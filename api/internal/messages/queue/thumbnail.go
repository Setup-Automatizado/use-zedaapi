package queue

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"github.com/nfnt/resize"

	wameow "go.mau.fi/whatsmeow"

	"github.com/gen2brain/go-fitz"
)

// ThumbnailGenerator handles thumbnail generation and upload
type ThumbnailGenerator struct {
	client *wameow.Client
	log    *slog.Logger
}

// ThumbnailResult contains thumbnail data and upload metadata
type ThumbnailResult struct {
	Data          []byte
	Width         uint32
	Height        uint32
	DirectPath    string
	MediaKey      []byte
	FileEncSha256 []byte
	FileSha256    []byte
	FileLength    uint64
	PageCount     uint32 // PDF page count (0 for non-PDF)
}

const (
	// ImageThumbnailMaxSize is the max dimension for image thumbnails (72x72 matches WhatsApp multidevice)
	ImageThumbnailMaxSize = 72
	// VideoThumbnailWidth is the width for video thumbnails (100px wide, aspect ratio maintained)
	VideoThumbnailWidth = 100
	// PDFThumbnailMaxSize is the max dimension for PDF thumbnails
	PDFThumbnailMaxSize = 300

	// ImageThumbnailJPEGQuality is the JPEG quality for image thumbnails
	ImageThumbnailJPEGQuality = 85
	// VideoThumbnailJPEGQuality is the JPEG quality for video thumbnails
	VideoThumbnailJPEGQuality = 80
	// PDFThumbnailJPEGQuality is the JPEG quality for PDF thumbnails
	PDFThumbnailJPEGQuality = 95
)

// NewThumbnailGenerator creates a new thumbnail generator
func NewThumbnailGenerator(client *wameow.Client, log *slog.Logger) *ThumbnailGenerator {
	return &ThumbnailGenerator{
		client: client,
		log:    log.With(slog.String("component", "thumbnail")),
	}
}

// GenerateAndUploadImageThumbnail generates thumbnail from image data and uploads it
func (t *ThumbnailGenerator) GenerateAndUploadImageThumbnail(ctx context.Context, imageData []byte, mimeType string) (*ThumbnailResult, error) {
	// Decode image
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	t.log.Debug("decoded image for thumbnail",
		slog.String("format", format),
		slog.Int("width", img.Bounds().Dx()),
		slog.Int("height", img.Bounds().Dy()))

	// Resize to 72x72 max maintaining aspect ratio (matches WhatsApp multidevice thumbnail size)
	thumbnail := resize.Thumbnail(ImageThumbnailMaxSize, ImageThumbnailMaxSize, img, resize.Lanczos3)

	// Encode as JPEG
	thumbnailData, err := t.encodeJPEG(thumbnail, ImageThumbnailJPEGQuality)
	if err != nil {
		return nil, fmt.Errorf("encode thumbnail: %w", err)
	}

	t.log.Debug("generated thumbnail",
		slog.Int("width", thumbnail.Bounds().Dx()),
		slog.Int("height", thumbnail.Bounds().Dy()),
		slog.Int("size_bytes", len(thumbnailData)))

	// Upload thumbnail
	return t.uploadThumbnail(ctx, thumbnailData, thumbnail.Bounds().Dx(), thumbnail.Bounds().Dy())
}

// GenerateAndUploadVideoThumbnail generates thumbnail from video first frame using ffmpeg
func (t *ThumbnailGenerator) GenerateAndUploadVideoThumbnail(ctx context.Context, videoData []byte) (*ThumbnailResult, error) {
	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.log.Warn("ffmpeg not found, generating default video thumbnail")
		return t.generateDefaultVideoThumbnail(ctx)
	}

	// Generate thumbnail using ffmpeg (try in-memory first, fallback to temp files)
	thumbnailData, width, height, err := t.extractVideoFrameFFmpeg(videoData)
	if err != nil {
		t.log.Warn("failed to extract video frame",
			slog.String("error", err.Error()))
		return t.generateDefaultVideoThumbnail(ctx)
	}

	if len(thumbnailData) == 0 {
		return t.generateDefaultVideoThumbnail(ctx)
	}

	// Upload thumbnail
	return t.uploadThumbnail(ctx, thumbnailData, width, height)
}

// GenerateAndUploadDocumentThumbnail generates thumbnail from document (PDF first page)
func (t *ThumbnailGenerator) GenerateAndUploadDocumentThumbnail(ctx context.Context, documentData []byte, mimeType string) (*ThumbnailResult, error) {
	// Only support PDF for now
	if mimeType != "application/pdf" {
		t.log.Debug("document thumbnail only supported for PDF",
			slog.String("mime_type", mimeType))
		return nil, nil
	}

	// Extract first page as image + page count
	thumbnailData, width, height, pageCount, err := t.extractPDFFirstPage(documentData)
	if err != nil {
		t.log.Warn("failed to extract PDF first page",
			slog.String("error", err.Error()))
		return nil, nil
	}

	if len(thumbnailData) == 0 {
		return nil, nil
	}

	// Upload thumbnail
	result, err := t.uploadThumbnail(ctx, thumbnailData, width, height)
	if err != nil {
		return nil, err
	}

	// Set page count from PDF
	result.PageCount = uint32(pageCount)

	return result, nil
}

// extractVideoFrameFFmpeg extracts a frame from video using ffmpeg
func (t *ThumbnailGenerator) extractVideoFrameFFmpeg(videoData []byte) ([]byte, int, int, error) {
	// Create temp files
	tempVideoPath := fmt.Sprintf("/tmp/video_%d.mp4", time.Now().UnixNano())
	tempFramePath := fmt.Sprintf("/tmp/frame_%d.jpg", time.Now().UnixNano())

	// Save video to temp file
	if err := os.WriteFile(tempVideoPath, videoData, 0644); err != nil {
		return nil, 0, 0, fmt.Errorf("write temp video: %w", err)
	}
	defer os.Remove(tempVideoPath)
	defer os.Remove(tempFramePath)

	// Extract frame at 1 second using ffmpeg (no scaling, raw frame)
	cmd := exec.Command("ffmpeg",
		"-i", tempVideoPath,
		"-ss", "00:00:01.000",
		"-vframes", "1",
		tempFramePath)

	// Run ffmpeg
	if err := cmd.Run(); err != nil {
		return nil, 0, 0, fmt.Errorf("ffmpeg extract frame: %w", err)
	}

	// Read extracted frame
	frameData, err := os.ReadFile(tempFramePath)
	if err != nil || len(frameData) == 0 {
		return nil, 0, 0, fmt.Errorf("read extracted frame: %w", err)
	}

	// Decode to get dimensions
	img, _, err := image.Decode(bytes.NewReader(frameData))
	if err != nil {
		// Return frame data anyway, use default dimensions
		return frameData, 100, 75, nil
	}

	// Resize to 100px width maintaining aspect ratio (matches WhatsApp multidevice)
	resized := resize.Resize(VideoThumbnailWidth, 0, img, resize.Lanczos3)

	// Re-encode as JPEG
	thumbnailData, err := t.encodeJPEG(resized, VideoThumbnailJPEGQuality)
	if err != nil {
		return frameData, img.Bounds().Dx(), img.Bounds().Dy(), nil
	}

	return thumbnailData, resized.Bounds().Dx(), resized.Bounds().Dy(), nil
}

// extractPDFFirstPage extracts first page of PDF as image and returns page count
func (t *ThumbnailGenerator) extractPDFFirstPage(pdfData []byte) ([]byte, int, int, int, error) {
	// Open PDF document
	doc, err := fitz.NewFromMemory(pdfData)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("open pdf: %w", err)
	}
	defer doc.Close()

	// Get page count
	pageCount := doc.NumPage()
	if pageCount == 0 {
		return nil, 0, 0, 0, fmt.Errorf("pdf has no pages")
	}

	// Render first page as image
	img, err := doc.Image(0) // Page 0 (first page)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("render pdf page: %w", err)
	}

	// Resize maintaining aspect ratio (300px max matches WhatsApp PDF thumbnail size)
	resized := t.resizeImageMaintainAspect(img, PDFThumbnailMaxSize, PDFThumbnailMaxSize)

	// Encode as JPEG
	thumbnailData, err := t.encodeJPEG(resized, PDFThumbnailJPEGQuality)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("encode pdf thumbnail: %w", err)
	}

	bounds := resized.Bounds()
	return thumbnailData, bounds.Dx(), bounds.Dy(), pageCount, nil
}

// generateDefaultVideoThumbnail returns a minimal 1x1 pixel JPEG placeholder
// This matches WhatsApp multidevice behavior for fallback thumbnails
func (t *ThumbnailGenerator) generateDefaultVideoThumbnail(_ context.Context) (*ThumbnailResult, error) {
	// Minimal valid 1x1 JPEG (matches WhatsApp multidevice default)
	data := []byte{
		0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01,
		0x01, 0x01, 0x00, 0x48, 0x00, 0x48, 0x00, 0x00, 0xFF, 0xC0, 0x00, 0x11,
		0x08, 0x00, 0x01, 0x00, 0x01, 0x01, 0x01, 0x11, 0x00, 0x02, 0x11, 0x01,
		0x03, 0x11, 0x01, 0xFF, 0xC4, 0x00, 0x14, 0x00, 0x01, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x08, 0xFF, 0xDA, 0x00, 0x08, 0x01, 0x01, 0x00, 0x00, 0x3F, 0x00, 0x00,
		0xFF, 0xD9,
	}
	return &ThumbnailResult{
		Data:   data,
		Width:  1,
		Height: 1,
	}, nil
}

// resizeImageMaintainAspect resizes image maintaining aspect ratio
func (t *ThumbnailGenerator) resizeImageMaintainAspect(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	currentWidth := bounds.Dx()
	currentHeight := bounds.Dy()

	// If image is already smaller, return original
	if currentWidth <= maxWidth && currentHeight <= maxHeight {
		return img
	}

	// Calculate new dimensions maintaining aspect ratio
	var newWidth, newHeight int
	aspectRatio := float64(currentWidth) / float64(currentHeight)

	if aspectRatio > 1.0 {
		// Landscape
		newWidth = maxWidth
		newHeight = int(float64(maxWidth) / aspectRatio)
		if newHeight > maxHeight {
			newHeight = maxHeight
			newWidth = int(float64(maxHeight) * aspectRatio)
		}
	} else {
		// Portrait
		newHeight = maxHeight
		newWidth = int(float64(maxHeight) * aspectRatio)
		if newWidth > maxWidth {
			newWidth = maxWidth
			newHeight = int(float64(maxWidth) / aspectRatio)
		}
	}

	// Use nfnt/resize for better quality
	return resize.Resize(uint(newWidth), uint(newHeight), img, resize.Lanczos3)
}

// encodeJPEG encodes image as JPEG with specified quality
func (t *ThumbnailGenerator) encodeJPEG(img image.Image, quality int) ([]byte, error) {
	var buf bytes.Buffer

	err := jpeg.Encode(&buf, img, &jpeg.Options{
		Quality: quality,
	})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// uploadThumbnail uploads thumbnail to WhatsApp servers
func (t *ThumbnailGenerator) uploadThumbnail(ctx context.Context, thumbnailData []byte, width, height int) (*ThumbnailResult, error) {
	// Upload using whatsmeow client
	uploadResp, err := t.client.Upload(ctx, thumbnailData, wameow.MediaImage)
	if err != nil {
		return nil, fmt.Errorf("upload thumbnail: %w", err)
	}

	t.log.Info("uploaded thumbnail",
		slog.Int("size_bytes", len(thumbnailData)),
		slog.String("direct_path", uploadResp.DirectPath))

	return &ThumbnailResult{
		Data:          thumbnailData,
		Width:         uint32(width),
		Height:        uint32(height),
		DirectPath:    uploadResp.DirectPath,
		MediaKey:      uploadResp.MediaKey,
		FileEncSha256: uploadResp.FileEncSHA256,
		FileSha256:    uploadResp.FileSHA256,
		FileLength:    uploadResp.FileLength,
	}, nil
}

// DecodeImage is a helper to decode various image formats
func DecodeImage(data []byte) (image.Image, string, error) {
	// Try PNG first
	img, err := png.Decode(bytes.NewReader(data))
	if err == nil {
		return img, "png", nil
	}

	// Try JPEG
	img, err = jpeg.Decode(bytes.NewReader(data))
	if err == nil {
		return img, "jpeg", nil
	}

	// Use generic decoder
	return image.Decode(bytes.NewReader(data))
}
