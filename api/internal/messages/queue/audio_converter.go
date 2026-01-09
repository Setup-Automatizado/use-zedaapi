package queue

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// AudioConverter handles audio format conversion to WhatsApp native Opus/OGG format
// Converts various audio formats (MP3, WAV, M4A, AAC, FLAC, etc.) to Opus codec in OGG container
// This ensures all audio messages appear as native WhatsApp voice notes with waveform display
type AudioConverter struct {
	log *slog.Logger
}

// ConversionResult contains converted audio data and metadata
type ConversionResult struct {
	Data     []byte // Converted audio data in Opus/OGG format
	MimeType string // Always "audio/ogg; codecs=opus" after conversion
	Duration int64  // Duration in seconds (0 if detection fails)
	Waveform []byte // Audio waveform data for WhatsApp voice note visualization (64 samples)
}

// NewAudioConverter creates a new audio converter
func NewAudioConverter(log *slog.Logger) *AudioConverter {
	return &AudioConverter{
		log: log.With(slog.String("component", "audio_converter")),
	}
}

// IsOpusFormat checks if audio is already in Opus/OGG format
// Returns true for audio/ogg or audio/ogg; codecs=opus
func (c *AudioConverter) IsOpusFormat(mimeType string) bool {
	mimeType = strings.ToLower(mimeType)
	// Check for Opus codec or OGG container
	// Note: Some OGG files may contain Vorbis, but we assume OGG = Opus for WhatsApp
	return strings.Contains(mimeType, "opus") ||
		(strings.Contains(mimeType, "ogg") && !strings.Contains(mimeType, "video"))
}

// Convert converts audio to WhatsApp native Opus/OGG format
// Parameters:
//   - audioData: Raw audio file bytes
//   - originalMimeType: Original audio MIME type (e.g., "audio/mpeg", "audio/wav")
//
// Returns converted audio in Opus/OGG format optimized for WhatsApp voice notes:
//   - Codec: libopus
//   - Bitrate: 24 kbps (optimal for voice)
//   - Sample Rate: 16000 Hz
//   - Channels: 1 (mono)
//   - Container: OGG
func (c *AudioConverter) Convert(audioData []byte, originalMimeType string) (*ConversionResult, error) {
	// Check if already in Opus/OGG format
	if c.IsOpusFormat(originalMimeType) {
		c.log.Debug("audio already in Opus/OGG format, no conversion needed",
			slog.String("mime_type", originalMimeType))
		return &ConversionResult{
			Data:     audioData,
			MimeType: "audio/ogg; codecs=opus",
		}, nil
	}

	// Check ffmpeg availability
	if !c.isFFmpegAvailable() {
		return nil, fmt.Errorf("ffmpeg not found - install ffmpeg to enable audio conversion for native WhatsApp voice notes")
	}

	// Perform conversion
	convertedData, duration, waveform, err := c.convertWithFFmpeg(audioData, originalMimeType)
	if err != nil {
		return nil, fmt.Errorf("convert audio to Opus: %w", err)
	}

	return &ConversionResult{
		Data:     convertedData,
		MimeType: "audio/ogg; codecs=opus",
		Duration: duration,
		Waveform: waveform,
	}, nil
}

// isFFmpegAvailable checks if ffmpeg is installed and available in PATH
func (c *AudioConverter) isFFmpegAvailable() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

// convertWithFFmpeg performs actual audio conversion using ffmpeg
// Uses optimal parameters for WhatsApp voice notes:
//   - libopus codec with VBR (variable bitrate)
//   - 24 kbps bitrate (excellent quality for voice)
//   - 16kHz mono (WhatsApp standard)
//   - Maximum compression for smaller file size
//   - 20ms frame duration (WhatsApp standard)
func (c *AudioConverter) convertWithFFmpeg(audioData []byte, mimeType string) ([]byte, int64, []byte, error) {
	// Create temp directory for conversion
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("audio_conv_%d", time.Now().UnixNano()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, 0, nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Determine input file extension from MIME type
	inputExt := c.getExtensionForMimeType(mimeType)
	inputPath := filepath.Join(tempDir, "input"+inputExt)
	outputPath := filepath.Join(tempDir, "output.ogg")

	// Write input audio data to temp file
	if err := os.WriteFile(inputPath, audioData, 0644); err != nil {
		return nil, 0, nil, fmt.Errorf("write input file: %w", err)
	}

	// Build ffmpeg command with optimal WhatsApp voice note parameters
	// Reference: https://trac.ffmpeg.org/wiki/Encode/HighQualityAudio#Opus
	cmd := exec.Command("ffmpeg",
		"-i", inputPath, // Input file
		"-c:a", "libopus", // Opus audio codec
		"-b:a", "24k", // 24 kbps bitrate (optimal for voice)
		"-vbr", "on", // Variable bitrate (better quality)
		"-compression_level", "10", // Maximum compression (0-10)
		"-frame_duration", "20", // 20ms frames (WhatsApp standard)
		"-ar", "16000", // 16kHz sample rate (mono voice)
		"-ac", "1", // 1 channel (mono)
		"-application", "voip", // Optimize for voice (not music)
		"-f", "ogg", // OGG container format
		"-y", // Overwrite output file
		outputPath)

	// Capture stderr for error reporting
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Execute conversion
	c.log.Debug("converting audio to Opus/OGG",
		slog.String("original_mime", mimeType),
		slog.Int("original_size_bytes", len(audioData)))

	if err := cmd.Run(); err != nil {
		return nil, 0, nil, fmt.Errorf("ffmpeg conversion failed: %w - stderr: %s", err, stderr.String())
	}

	// Read converted audio file
	convertedData, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("read converted file: %w", err)
	}

	// Get audio duration using ffprobe
	duration := c.getAudioDuration(outputPath)

	// Generate waveform data for WhatsApp voice note visualization
	waveform := c.generateWaveform(outputPath)

	// Log conversion success with metrics
	compressionRatio := float64(len(convertedData)) / float64(len(audioData)) * 100
	c.log.Info("audio converted to Opus/OGG successfully",
		slog.String("original_mime", mimeType),
		slog.Int("original_size_bytes", len(audioData)),
		slog.Int("converted_size_bytes", len(convertedData)),
		slog.Float64("compression_ratio_percent", compressionRatio),
		slog.Int64("duration_seconds", duration),
		slog.Int("waveform_samples", len(waveform)))

	return convertedData, duration, waveform, nil
}

// getExtensionForMimeType returns appropriate file extension for audio MIME type
// Used to write temp input file with correct extension for ffmpeg
func (c *AudioConverter) getExtensionForMimeType(mimeType string) string {
	extensions := map[string]string{
		"audio/mpeg":     ".mp3",
		"audio/mp3":      ".mp3",
		"audio/mp4":      ".m4a",
		"audio/m4a":      ".m4a",
		"audio/aac":      ".aac",
		"audio/wav":      ".wav",
		"audio/x-wav":    ".wav",
		"audio/wave":     ".wav",
		"audio/flac":     ".flac",
		"audio/x-flac":   ".flac",
		"audio/webm":     ".webm",
		"audio/3gpp":     ".3gp",
		"audio/3gp":      ".3gp",
		"audio/amr":      ".amr",
		"audio/x-ms-wma": ".wma",
		"audio/ogg":      ".ogg",
	}

	if ext, ok := extensions[mimeType]; ok {
		return ext
	}

	// Fallback to generic extension
	return ".dat"
}

// getAudioDuration extracts audio duration using ffprobe
// Returns duration in seconds, or 0 if detection fails
func (c *AudioConverter) getAudioDuration(filePath string) int64 {
	cmd := exec.Command("ffprobe",
		"-v", "error", // Only show errors
		"-show_entries", "format=duration", // Extract duration
		"-of", "default=noprint_wrappers=1:nokey=1", // Output format
		filePath)

	output, err := cmd.Output()
	if err != nil {
		c.log.Warn("failed to get audio duration",
			slog.String("error", err.Error()))
		return 0
	}

	// Parse duration from output
	var duration float64
	if _, err := fmt.Sscanf(string(output), "%f", &duration); err != nil {
		c.log.Warn("failed to parse audio duration",
			slog.String("output", string(output)))
		return 0
	}

	return int64(duration)
}

// generateWaveform generates waveform data for WhatsApp voice note visualization
// WhatsApp expects 64 samples of amplitude data (0-100 range) for the waveform display
// Uses ffmpeg to extract audio samples and calculates RMS amplitude for each segment
func (c *AudioConverter) generateWaveform(audioPath string) []byte {
	const waveformSamples = 64 // WhatsApp uses 64 samples for waveform

	// Use ffmpeg to extract raw PCM samples
	// -f s16le: signed 16-bit little-endian
	// -ac 1: mono
	// -ar 8000: 8kHz sample rate (enough for waveform)
	cmd := exec.Command("ffmpeg",
		"-i", audioPath,
		"-f", "s16le",
		"-ac", "1",
		"-ar", "8000",
		"-")

	output, err := cmd.Output()
	if err != nil {
		c.log.Warn("failed to extract PCM for waveform",
			slog.String("error", err.Error()))
		return c.generateDefaultWaveform()
	}

	if len(output) < 2 {
		c.log.Warn("PCM output too short for waveform")
		return c.generateDefaultWaveform()
	}

	// Convert bytes to int16 samples
	numSamples := len(output) / 2
	samples := make([]int16, numSamples)
	for i := 0; i < numSamples; i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(output[i*2 : i*2+2]))
	}

	// Calculate samples per waveform segment
	samplesPerSegment := numSamples / waveformSamples
	if samplesPerSegment < 1 {
		samplesPerSegment = 1
	}

	// Calculate RMS amplitude for each segment
	waveform := make([]byte, waveformSamples)
	var maxAmplitude float64

	amplitudes := make([]float64, waveformSamples)
	for i := 0; i < waveformSamples; i++ {
		start := i * samplesPerSegment
		end := start + samplesPerSegment
		if end > numSamples {
			end = numSamples
		}
		if start >= numSamples {
			break
		}

		// Calculate RMS (Root Mean Square) for this segment
		var sumSquares float64
		for j := start; j < end; j++ {
			sumSquares += float64(samples[j]) * float64(samples[j])
		}
		rms := math.Sqrt(sumSquares / float64(end-start))
		amplitudes[i] = rms

		if rms > maxAmplitude {
			maxAmplitude = rms
		}
	}

	// Normalize to 0-100 range (WhatsApp expects values in this range)
	if maxAmplitude > 0 {
		for i := 0; i < waveformSamples; i++ {
			// Normalize and apply some compression for better visualization
			normalized := amplitudes[i] / maxAmplitude
			// Apply slight compression to make quiet parts more visible
			compressed := math.Pow(normalized, 0.7)
			waveform[i] = byte(compressed * 100)
		}
	}

	c.log.Debug("generated waveform",
		slog.Int("total_samples", numSamples),
		slog.Int("samples_per_segment", samplesPerSegment),
		slog.Int("waveform_samples", waveformSamples))

	return waveform
}

// generateDefaultWaveform creates a simple default waveform pattern
// Used as fallback when audio analysis fails
func (c *AudioConverter) generateDefaultWaveform() []byte {
	const waveformSamples = 64
	waveform := make([]byte, waveformSamples)

	// Generate a simple sine wave pattern as default
	for i := 0; i < waveformSamples; i++ {
		// Create a gentle wave pattern
		t := float64(i) / float64(waveformSamples) * math.Pi * 4
		waveform[i] = byte(30 + 20*math.Sin(t))
	}

	return waveform
}

// GenerateWaveformFromData generates waveform directly from audio data
// Used when audio is already in correct format and no conversion is needed
func (c *AudioConverter) GenerateWaveformFromData(audioData []byte, mimeType string) []byte {
	// Create temp file for analysis
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("waveform_%d", time.Now().UnixNano()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		c.log.Warn("failed to create temp dir for waveform",
			slog.String("error", err.Error()))
		return c.generateDefaultWaveform()
	}
	defer os.RemoveAll(tempDir)

	ext := c.getExtensionForMimeType(mimeType)
	tempPath := filepath.Join(tempDir, "audio"+ext)

	if err := os.WriteFile(tempPath, audioData, 0644); err != nil {
		c.log.Warn("failed to write temp audio for waveform",
			slog.String("error", err.Error()))
		return c.generateDefaultWaveform()
	}

	return c.generateWaveform(tempPath)
}

// GetDurationFromData extracts duration from audio data
// Used when audio is already in correct format
func (c *AudioConverter) GetDurationFromData(audioData []byte, mimeType string) int64 {
	// Create temp file for analysis
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("duration_%d", time.Now().UnixNano()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return 0
	}
	defer os.RemoveAll(tempDir)

	ext := c.getExtensionForMimeType(mimeType)
	tempPath := filepath.Join(tempDir, "audio"+ext)

	if err := os.WriteFile(tempPath, audioData, 0644); err != nil {
		return 0
	}

	return c.getAudioDuration(tempPath)
}

// GetDurationFromFile extracts duration from audio file using ffprobe
// Returns duration as string suitable for display (MM:SS format)
func (c *AudioConverter) GetDurationString(filePath string) string {
	duration := c.getAudioDuration(filePath)
	if duration <= 0 {
		return "0:00"
	}

	minutes := duration / 60
	seconds := duration % 60
	return strconv.FormatInt(minutes, 10) + ":" + fmt.Sprintf("%02d", seconds)
}
