package queue

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// StickerConverter handles image format conversion to WhatsApp native WebP sticker format
// Converts various image formats (PNG, JPG, GIF, etc.) to WebP with proper sizing
// WhatsApp sticker requirements:
//   - Format: WebP
//   - Size: 512x512 pixels maximum (maintains aspect ratio)
//   - File size: Ideally under 100KB
//   - Background: Transparent (if PNG/WebP with alpha)
type StickerConverter struct {
	log *slog.Logger
}

// StickerConversionResult contains converted sticker data and metadata
type StickerConversionResult struct {
	Data     []byte // Converted image data in WebP format
	MimeType string // Always "image/webp" after conversion
	Width    int    // Final width in pixels
	Height   int    // Final height in pixels
}

// NewStickerConverter creates a new sticker converter
func NewStickerConverter(log *slog.Logger) *StickerConverter {
	return &StickerConverter{
		log: log.With(slog.String("component", "sticker_converter")),
	}
}

// IsWebPFormat checks if image is already in WebP format
func (c *StickerConverter) IsWebPFormat(mimeType string) bool {
	mimeType = strings.ToLower(mimeType)
	return strings.Contains(mimeType, "webp")
}

// Convert converts image to WhatsApp native WebP sticker format
// Parameters:
//   - imageData: Raw image file bytes
//   - originalMimeType: Original image MIME type (e.g., "image/png", "image/jpeg")
//
// Returns converted image in WebP format optimized for WhatsApp stickers:
//   - Format: WebP with lossless compression for transparency
//   - Size: 512x512 max (maintains aspect ratio)
//   - Quality: 80 (balance between quality and size)
func (c *StickerConverter) Convert(imageData []byte, originalMimeType string) (*StickerConversionResult, error) {
	// Check ffmpeg availability
	if !c.isFFmpegAvailable() {
		return nil, fmt.Errorf("ffmpeg not found - install ffmpeg to enable sticker conversion to WebP format")
	}

	// Always convert to ensure proper sizing and format
	// Even WebP images may need resizing to 512x512
	convertedData, width, height, err := c.convertWithFFmpeg(imageData, originalMimeType)
	if err != nil {
		return nil, fmt.Errorf("convert image to WebP sticker: %w", err)
	}

	return &StickerConversionResult{
		Data:     convertedData,
		MimeType: "image/webp",
		Width:    width,
		Height:   height,
	}, nil
}

// isFFmpegAvailable checks if ffmpeg is installed and available in PATH
func (c *StickerConverter) isFFmpegAvailable() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

// convertWithFFmpeg performs actual image conversion using ffmpeg
// Uses optimal parameters for WhatsApp stickers:
//   - WebP format with transparency support
//   - 512x512 max size (maintains aspect ratio)
//   - Quality 80 for good balance
//   - Lossless for images with alpha channel
func (c *StickerConverter) convertWithFFmpeg(imageData []byte, mimeType string) ([]byte, int, int, error) {
	// Create temp directory for conversion
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("sticker_conv_%d", time.Now().UnixNano()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, 0, 0, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Determine input file extension from MIME type
	inputExt := c.getExtensionForMimeType(mimeType)
	inputPath := filepath.Join(tempDir, "input"+inputExt)
	outputPath := filepath.Join(tempDir, "output.webp")

	// Write input image data to temp file
	if err := os.WriteFile(inputPath, imageData, 0644); err != nil {
		return nil, 0, 0, fmt.Errorf("write input file: %w", err)
	}

	// Build ffmpeg command with optimal WhatsApp sticker parameters
	// Scale to max 512x512 while maintaining aspect ratio
	// Use WebP format with good quality and transparency support
	cmd := exec.Command("ffmpeg",
		"-i", inputPath, // Input file
		"-vf", "scale='min(512,iw)':min'(512,ih)':force_original_aspect_ratio=decrease", // Scale to max 512x512
		"-c:v", "libwebp", // WebP codec
		"-quality", "80", // Quality level (0-100)
		"-lossless", "0", // Lossy compression (better size)
		"-compression_level", "6", // Compression level (0-6)
		"-preset", "picture", // Optimize for pictures
		"-pix_fmt", "yuva420p", // Support transparency
		"-y", // Overwrite output file
		outputPath)

	// Capture stderr for error reporting
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Execute conversion
	c.log.Debug("converting image to WebP sticker",
		slog.String("original_mime", mimeType),
		slog.Int("original_size_bytes", len(imageData)))

	if err := cmd.Run(); err != nil {
		// Try fallback without transparency if first attempt fails
		c.log.Debug("first conversion attempt failed, trying fallback",
			slog.String("error", err.Error()))
		return c.convertWithFFmpegFallback(inputPath, outputPath, mimeType, len(imageData))
	}

	// Read converted image file
	convertedData, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("read converted file: %w", err)
	}

	// Get image dimensions
	width, height := c.getImageDimensions(outputPath)

	// Log conversion success with metrics
	compressionRatio := float64(len(convertedData)) / float64(len(imageData)) * 100
	c.log.Info("image converted to WebP sticker successfully",
		slog.String("original_mime", mimeType),
		slog.Int("original_size_bytes", len(imageData)),
		slog.Int("converted_size_bytes", len(convertedData)),
		slog.Float64("compression_ratio_percent", compressionRatio),
		slog.Int("width", width),
		slog.Int("height", height))

	return convertedData, width, height, nil
}

// convertWithFFmpegFallback performs conversion without alpha channel support
// Used as fallback when primary conversion fails
func (c *StickerConverter) convertWithFFmpegFallback(inputPath, outputPath, mimeType string, originalSize int) ([]byte, int, int, error) {
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-vf", "scale=512:512:force_original_aspect_ratio=decrease,pad=512:512:(ow-iw)/2:(oh-ih)/2:color=0x00000000",
		"-c:v", "libwebp",
		"-quality", "80",
		"-lossless", "0",
		"-y",
		outputPath)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, 0, 0, fmt.Errorf("ffmpeg conversion failed: %w - stderr: %s", err, stderr.String())
	}

	convertedData, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("read converted file: %w", err)
	}

	width, height := c.getImageDimensions(outputPath)

	compressionRatio := float64(len(convertedData)) / float64(originalSize) * 100
	c.log.Info("image converted to WebP sticker (fallback) successfully",
		slog.String("original_mime", mimeType),
		slog.Int("original_size_bytes", originalSize),
		slog.Int("converted_size_bytes", len(convertedData)),
		slog.Float64("compression_ratio_percent", compressionRatio),
		slog.Int("width", width),
		slog.Int("height", height))

	return convertedData, width, height, nil
}

// getExtensionForMimeType returns appropriate file extension for image MIME type
func (c *StickerConverter) getExtensionForMimeType(mimeType string) string {
	extensions := map[string]string{
		"image/png":     ".png",
		"image/jpeg":    ".jpg",
		"image/jpg":     ".jpg",
		"image/gif":     ".gif",
		"image/webp":    ".webp",
		"image/bmp":     ".bmp",
		"image/tiff":    ".tiff",
		"image/svg+xml": ".svg",
	}

	if ext, ok := extensions[mimeType]; ok {
		return ext
	}

	// Fallback to generic extension
	return ".dat"
}

// getImageDimensions extracts image dimensions using ffprobe
// Returns width and height in pixels, or 0 if detection fails
func (c *StickerConverter) getImageDimensions(filePath string) (int, int) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=s=x:p=0",
		filePath)

	output, err := cmd.Output()
	if err != nil {
		c.log.Warn("failed to get image dimensions",
			slog.String("error", err.Error()))
		return 512, 512 // Default sticker size
	}

	var width, height int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(output)), "%dx%d", &width, &height); err != nil {
		c.log.Warn("failed to parse image dimensions",
			slog.String("output", string(output)))
		return 512, 512
	}

	return width, height
}

// ConvertAnimatedToStatic converts animated images (GIF/APNG) to static WebP
// Takes the first frame of the animation
func (c *StickerConverter) ConvertAnimatedToStatic(imageData []byte, originalMimeType string) (*StickerConversionResult, error) {
	if !c.isFFmpegAvailable() {
		return nil, fmt.Errorf("ffmpeg not found")
	}

	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("sticker_anim_%d", time.Now().UnixNano()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	inputExt := c.getExtensionForMimeType(originalMimeType)
	inputPath := filepath.Join(tempDir, "input"+inputExt)
	outputPath := filepath.Join(tempDir, "output.webp")

	if err := os.WriteFile(inputPath, imageData, 0644); err != nil {
		return nil, fmt.Errorf("write input file: %w", err)
	}

	// Extract first frame and convert to WebP
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-vf", "select=eq(n\\,0),scale=512:512:force_original_aspect_ratio=decrease",
		"-frames:v", "1",
		"-c:v", "libwebp",
		"-quality", "80",
		"-lossless", "0",
		"-y",
		outputPath)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg animated conversion failed: %w - stderr: %s", err, stderr.String())
	}

	convertedData, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("read converted file: %w", err)
	}

	width, height := c.getImageDimensions(outputPath)

	c.log.Info("animated image converted to static WebP sticker",
		slog.String("original_mime", originalMimeType),
		slog.Int("original_size_bytes", len(imageData)),
		slog.Int("converted_size_bytes", len(convertedData)),
		slog.Int("width", width),
		slog.Int("height", height))

	return &StickerConversionResult{
		Data:     convertedData,
		MimeType: "image/webp",
		Width:    width,
		Height:   height,
	}, nil
}
