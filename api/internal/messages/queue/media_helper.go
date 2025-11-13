package queue

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
)

// MediaType represents supported media types for validation
type MediaType string

const (
	MediaTypeImage    MediaType = "image"
	MediaTypeVideo    MediaType = "video"
	MediaTypeAudio    MediaType = "audio"
	MediaTypeDocument MediaType = "document"
)

// MediaDownloader handles downloading media from URLs or decoding base64
type MediaDownloader struct {
	httpClient *http.Client
	maxSize    int64 // Maximum file size in bytes
}

// NewMediaDownloader creates a new media downloader
func NewMediaDownloader(maxSizeMB int) *MediaDownloader {
	return &MediaDownloader{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxSize: int64(maxSizeMB * 1024 * 1024),
	}
}

// Download downloads media from URL or decodes base64 data
// Supports: http(s):// URLs and data: URI scheme (data:image/png;base64,...)
func (d *MediaDownloader) Download(mediaURL string) ([]byte, string, error) {
	// Check if it's a data URI (base64 encoded)
	if strings.HasPrefix(mediaURL, "data:") {
		return d.decodeDataURI(mediaURL)
	}

	// Download from URL
	if strings.HasPrefix(mediaURL, "http://") || strings.HasPrefix(mediaURL, "https://") {
		return d.downloadFromURL(mediaURL)
	}

	return nil, "", fmt.Errorf("invalid media URL format: must be http(s):// or data: URI")
}

// downloadFromURL downloads media from HTTP(S) URL
func (d *MediaDownloader) downloadFromURL(url string) ([]byte, string, error) {
	resp, err := d.httpClient.Get(url)
	if err != nil {
		return nil, "", fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("http status %d: %s", resp.StatusCode, resp.Status)
	}

	// Check content length
	if resp.ContentLength > d.maxSize {
		return nil, "", fmt.Errorf("file too large: %d bytes (max %d)", resp.ContentLength, d.maxSize)
	}

	// Read with size limit
	limitedReader := io.LimitReader(resp.Body, d.maxSize+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, "", fmt.Errorf("read response: %w", err)
	}

	if int64(len(data)) > d.maxSize {
		return nil, "", fmt.Errorf("file too large: exceeds %d bytes", d.maxSize)
	}

	// Get MIME type from response
	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = detectMimeType(data)
	}

	return data, mimeType, nil
}

// decodeDataURI decodes data URI scheme (data:image/png;base64,...)
func (d *MediaDownloader) decodeDataURI(dataURI string) ([]byte, string, error) {
	// Format: data:[<mediatype>][;base64],<data>
	parts := strings.SplitN(dataURI, ",", 2)
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("invalid data URI format")
	}

	// Parse media type
	header := parts[0]
	mimeType := "application/octet-stream" // default
	isBase64 := false

	if strings.Contains(header, ";") {
		headerParts := strings.Split(header, ";")
		for _, part := range headerParts {
			part = strings.TrimSpace(part)
			if part == "base64" {
				isBase64 = true
			} else if strings.HasPrefix(part, "data:") {
				mimeType = strings.TrimPrefix(part, "data:")
			} else if !strings.HasPrefix(part, "data:") && part != "" {
				mimeType = part
			}
		}
	}

	// Decode data
	var data []byte
	var err error

	if isBase64 {
		data, err = base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, "", fmt.Errorf("decode base64: %w", err)
		}
	} else {
		data = []byte(parts[1])
	}

	// Check size
	if int64(len(data)) > d.maxSize {
		return nil, "", fmt.Errorf("file too large: %d bytes (max %d)", len(data), d.maxSize)
	}

	return data, mimeType, nil
}

// detectMimeType attempts to detect MIME type from file data using magic numbers (file signatures)
// Reference: https://en.wikipedia.org/wiki/List_of_file_signatures
func detectMimeType(data []byte) string {
	if len(data) < 12 {
		return "application/octet-stream"
	}

	// === IMAGE FORMATS ===
	// JPEG (FF D8 FF)
	if bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}) {
		return "image/jpeg"
	}

	// PNG (89 50 4E 47 0D 0A 1A 0A)
	if bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return "image/png"
	}

	// GIF (47 49 46 38)
	if bytes.HasPrefix(data, []byte{0x47, 0x49, 0x46, 0x38}) {
		return "image/gif"
	}

	// WebP (52 49 46 46 ... 57 45 42 50)
	if len(data) >= 20 && bytes.HasPrefix(data, []byte{0x52, 0x49, 0x46, 0x46}) &&
		bytes.Equal(data[8:12], []byte{0x57, 0x45, 0x42, 0x50}) {
		return "image/webp"
	}

	// BMP (42 4D)
	if bytes.HasPrefix(data, []byte{0x42, 0x4D}) {
		return "image/bmp"
	}

	// TIFF (49 49 2A 00 or 4D 4D 00 2A)
	if bytes.HasPrefix(data, []byte{0x49, 0x49, 0x2A, 0x00}) ||
		bytes.HasPrefix(data, []byte{0x4D, 0x4D, 0x00, 0x2A}) {
		return "image/tiff"
	}

	// === VIDEO FORMATS ===
	// MP4/M4A/M4V (00 00 00 [size] ftyp)
	if len(data) >= 12 && bytes.HasPrefix(data, []byte{0x00, 0x00, 0x00}) {
		if bytes.Contains(data[4:12], []byte("ftyp")) {
			// Check for specific subtypes
			if bytes.Contains(data[8:16], []byte("M4A")) || bytes.Contains(data[8:16], []byte("mp42")) {
				return "audio/mp4"
			}
			return "video/mp4"
		}
	}

	// WebM/MKV (1A 45 DF A3)
	if bytes.HasPrefix(data, []byte{0x1A, 0x45, 0xDF, 0xA3}) {
		// Distinguish between WebM (video) and MKV (video)
		return "video/webm" // WhatsApp prefers webm for web compatibility
	}

	// AVI (52 49 46 46 ... 41 56 49 20)
	if len(data) >= 12 && bytes.HasPrefix(data, []byte{0x52, 0x49, 0x46, 0x46}) &&
		bytes.Equal(data[8:12], []byte{0x41, 0x56, 0x49, 0x20}) {
		return "video/x-msvideo"
	}

	// FLV (46 4C 56 01)
	if bytes.HasPrefix(data, []byte{0x46, 0x4C, 0x56, 0x01}) {
		return "video/x-flv"
	}

	// MPEG (00 00 01 BA or 00 00 01 B3)
	if bytes.HasPrefix(data, []byte{0x00, 0x00, 0x01, 0xBA}) ||
		bytes.HasPrefix(data, []byte{0x00, 0x00, 0x01, 0xB3}) {
		return "video/mpeg"
	}

	// WMV (30 26 B2 75 8E 66 CF 11)
	if bytes.HasPrefix(data, []byte{0x30, 0x26, 0xB2, 0x75, 0x8E, 0x66, 0xCF, 0x11}) {
		return "video/x-ms-wmv"
	}

	// 3GP (00 00 00 [size] ftyp3gp)
	if len(data) >= 12 && bytes.HasPrefix(data, []byte{0x00, 0x00, 0x00}) &&
		bytes.Contains(data[4:12], []byte("3gp")) {
		return "video/3gpp"
	}

	// QuickTime/MOV (00 00 00 [size] moov/mdat/free/wide)
	if len(data) >= 12 && bytes.HasPrefix(data, []byte{0x00, 0x00, 0x00}) {
		if bytes.Contains(data[4:12], []byte("moov")) ||
			bytes.Contains(data[4:12], []byte("mdat")) ||
			bytes.Contains(data[4:12], []byte("wide")) ||
			bytes.Contains(data[4:12], []byte("free")) {
			return "video/quicktime"
		}
	}

	// === AUDIO FORMATS ===
	// OGG (4F 67 67 53)
	if bytes.HasPrefix(data, []byte{0x4F, 0x67, 0x67, 0x53}) {
		return "audio/ogg" // May contain Opus or Vorbis codec
	}

	// MP3 (FF FB or FF F3 or FF F2 or ID3)
	if bytes.HasPrefix(data, []byte{0xFF, 0xFB}) ||
		bytes.HasPrefix(data, []byte{0xFF, 0xF3}) ||
		bytes.HasPrefix(data, []byte{0xFF, 0xF2}) ||
		bytes.HasPrefix(data, []byte{0x49, 0x44, 0x33}) { // ID3
		return "audio/mpeg"
	}

	// WAV (52 49 46 46 ... 57 41 56 45)
	if len(data) >= 12 && bytes.HasPrefix(data, []byte{0x52, 0x49, 0x46, 0x46}) &&
		bytes.Equal(data[8:12], []byte{0x57, 0x41, 0x56, 0x45}) {
		return "audio/wav"
	}

	// FLAC (66 4C 61 43)
	if bytes.HasPrefix(data, []byte{0x66, 0x4C, 0x61, 0x43}) {
		return "audio/flac"
	}

	// AAC/ADTS (FF F1 or FF F9)
	if bytes.HasPrefix(data, []byte{0xFF, 0xF1}) || bytes.HasPrefix(data, []byte{0xFF, 0xF9}) {
		return "audio/aac"
	}

	// AMR (23 21 41 4D 52)
	if bytes.HasPrefix(data, []byte{0x23, 0x21, 0x41, 0x4D, 0x52}) {
		return "audio/amr"
	}

	// === DOCUMENT FORMATS ===
	// PDF (25 50 44 46)
	if bytes.HasPrefix(data, []byte{0x25, 0x50, 0x44, 0x46}) {
		return "application/pdf"
	}

	// ZIP and ZIP-based formats (50 4B 03 04 or 50 4B 05 06 or 50 4B 07 08)
	if bytes.HasPrefix(data, []byte{0x50, 0x4B, 0x03, 0x04}) ||
		bytes.HasPrefix(data, []byte{0x50, 0x4B, 0x05, 0x06}) ||
		bytes.HasPrefix(data, []byte{0x50, 0x4B, 0x07, 0x08}) {
		return detectZipBasedFormat(data)
	}

	// RAR (52 61 72 21 1A 07)
	if bytes.HasPrefix(data, []byte{0x52, 0x61, 0x72, 0x21, 0x1A, 0x07}) {
		return "application/x-rar-compressed"
	}

	// 7z (37 7A BC AF 27 1C)
	if bytes.HasPrefix(data, []byte{0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C}) {
		return "application/x-7z-compressed"
	}

	// GZIP (1F 8B)
	if bytes.HasPrefix(data, []byte{0x1F, 0x8B}) {
		return "application/gzip"
	}

	// TAR (ustar at offset 257)
	if len(data) >= 262 && bytes.Equal(data[257:262], []byte{0x75, 0x73, 0x74, 0x61, 0x72}) {
		return "application/x-tar"
	}

	// Microsoft Office 97-2003 (DOC, XLS, PPT, MDB) - OLE2/CFB (D0 CF 11 E0 A1 B1 1A E1)
	if bytes.HasPrefix(data, []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}) {
		return detectOldOfficeFormat(data)
	}

	// RTF (7B 5C 72 74 66)
	if bytes.HasPrefix(data, []byte{0x7B, 0x5C, 0x72, 0x74, 0x66}) {
		return "application/rtf"
	}

	// Fallback to http.DetectContentType for remaining formats
	mimeType := http.DetectContentType(data)

	// If still generic, return octet-stream
	if mimeType == "application/octet-stream" || mimeType == "text/plain; charset=utf-8" {
		return "application/octet-stream"
	}

	return mimeType
}

// detectZipBasedFormat detects specific ZIP-based formats (DOCX, XLSX, PPTX, ODT, ODS, ODP, EPUB, etc.)
func detectZipBasedFormat(data []byte) string {
	// Simple heuristic: check for specific strings in first 1KB
	dataStr := string(data[:min(len(data), 1024)])

	// Microsoft Office Open XML
	if strings.Contains(dataStr, "word/") || strings.Contains(dataStr, "_rels") {
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document" // DOCX
	}
	if strings.Contains(dataStr, "xl/") {
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" // XLSX
	}
	if strings.Contains(dataStr, "ppt/") {
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation" // PPTX
	}

	// OpenDocument Format (LibreOffice)
	if strings.Contains(dataStr, "mimetype") {
		if strings.Contains(dataStr, "application/vnd.oasis.opendocument.text") {
			return "application/vnd.oasis.opendocument.text" // ODT
		}
		if strings.Contains(dataStr, "application/vnd.oasis.opendocument.spreadsheet") {
			return "application/vnd.oasis.opendocument.spreadsheet" // ODS
		}
		if strings.Contains(dataStr, "application/vnd.oasis.opendocument.presentation") {
			return "application/vnd.oasis.opendocument.presentation" // ODP
		}
	}

	// EPUB
	if strings.Contains(dataStr, "EPUB") || strings.Contains(dataStr, "application/epub+zip") {
		return "application/epub+zip"
	}

	// Generic ZIP if no specific format detected
	return "application/zip"
}

// detectOldOfficeFormat detects legacy Microsoft Office formats (DOC, XLS, PPT, MDB)
func detectOldOfficeFormat(data []byte) string {
	// These formats use OLE2 (Compound File Binary) format
	// Detailed detection requires parsing the directory structure
	// For simplicity, we return generic office format

	// Check for Word-specific markers (rough heuristic)
	dataStr := string(data[:min(len(data), 8192)])
	if strings.Contains(dataStr, "Microsoft Word") || strings.Contains(dataStr, "Word.Document") {
		return "application/msword" // DOC
	}
	if strings.Contains(dataStr, "Microsoft Excel") || strings.Contains(dataStr, "Workbook") {
		return "application/vnd.ms-excel" // XLS
	}
	if strings.Contains(dataStr, "Microsoft PowerPoint") || strings.Contains(dataStr, "PowerPoint") {
		return "application/vnd.ms-powerpoint" // PPT
	}
	if strings.Contains(dataStr, "Standard Jet DB") {
		return "application/vnd.ms-access" // MDB
	}

	// Default to generic Office format
	return "application/msword" // Conservative default
}

// min returns minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ValidateMediaType checks if media type is supported for message type
func (d *MediaDownloader) ValidateMediaType(mimeType string, mediaType MediaType) error {
	switch mediaType {
	case MediaTypeImage:
		if !strings.HasPrefix(mimeType, "image/") {
			return fmt.Errorf("invalid image type: %s", mimeType)
		}
	case MediaTypeVideo:
		if !strings.HasPrefix(mimeType, "video/") {
			return fmt.Errorf("invalid video type: %s", mimeType)
		}
	case MediaTypeAudio:
		if !strings.HasPrefix(mimeType, "audio/") {
			return fmt.Errorf("invalid audio type: %s", mimeType)
		}
	case MediaTypeDocument:
		// Documents can be any type
		return nil
	}
	return nil
}

// GetMediaDimensions returns actual width and height for images and videos
// Supports all image formats (JPEG, PNG, GIF, WebP, BMP, TIFF) and all video formats (via ffprobe)
// Uses image.DecodeConfig() for fast dimension detection without decoding full image
func GetMediaDimensions(data []byte, mimeType string) (width, height int, err error) {
	// Detect if image or video
	if strings.HasPrefix(mimeType, "image/") {
		return getImageDimensions(data)
	}

	if strings.HasPrefix(mimeType, "video/") {
		return getVideoDimensions(data)
	}

	// Unknown type, return error
	return 0, 0, fmt.Errorf("unsupported media type for dimension detection: %s", mimeType)
}

// getImageDimensions detects image dimensions using image.DecodeConfig
// Much faster than image.Decode as it only reads headers, not pixel data
func getImageDimensions(data []byte) (width, height int, err error) {
	reader := bytes.NewReader(data)

	// Try standard library formats first (JPEG, PNG, GIF)
	config, format, err := image.DecodeConfig(reader)
	if err == nil {
		return config.Width, config.Height, nil
	}

	// Reset reader for additional attempts
	reader.Seek(0, 0)

	// Try WebP (requires golang.org/x/image/webp)
	config, err = webp.DecodeConfig(reader)
	if err == nil {
		return config.Width, config.Height, nil
	}

	// Reset reader
	reader.Seek(0, 0)

	// Try BMP (requires golang.org/x/image/bmp)
	config, err = bmp.DecodeConfig(reader)
	if err == nil {
		return config.Width, config.Height, nil
	}

	// Reset reader
	reader.Seek(0, 0)

	// Try TIFF (requires golang.org/x/image/tiff)
	config, err = tiff.DecodeConfig(reader)
	if err == nil {
		return config.Width, config.Height, nil
	}

	// All attempts failed
	return 0, 0, fmt.Errorf("failed to decode image dimensions (format: %s): %w", format, err)
}

// getVideoDimensions extracts video dimensions using ffprobe
// Requires ffmpeg/ffprobe to be installed on the system
func getVideoDimensions(data []byte) (width, height int, err error) {
	// Check if ffprobe is available
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return 0, 0, fmt.Errorf("ffprobe not found - install ffmpeg to detect video dimensions")
	}

	// Create temp file for video data
	tempFile, err := os.CreateTemp("", "video_*")
	if err != nil {
		return 0, 0, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write video data
	if _, err := tempFile.Write(data); err != nil {
		return 0, 0, fmt.Errorf("write temp file: %w", err)
	}

	// Close file before ffprobe reads it
	tempFile.Close()

	// Use ffprobe to extract dimensions
	// -v error: only show errors
	// -select_streams v:0: select first video stream
	// -show_entries stream=width,height: only show width and height
	// -of csv=p=0: output as CSV without headers (width,height)
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=p=0",
		tempFile.Name())

	output, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("ffprobe execution failed: %w", err)
	}

	// Parse output (format: "width,height")
	dimensions := strings.TrimSpace(string(output))
	parts := strings.Split(dimensions, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid ffprobe output format: %s", dimensions)
	}

	// Convert to integers
	if _, err := fmt.Sscanf(parts[0], "%d", &width); err != nil {
		return 0, 0, fmt.Errorf("parse width: %w", err)
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &height); err != nil {
		return 0, 0, fmt.Errorf("parse height: %w", err)
	}

	// Validate dimensions
	if width <= 0 || height <= 0 {
		return 0, 0, fmt.Errorf("invalid dimensions: %dx%d", width, height)
	}

	return width, height, nil
}
