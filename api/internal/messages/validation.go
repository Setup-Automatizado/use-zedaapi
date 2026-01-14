package messages

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	// Phone validation regex - E.164 format
	phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

	// Base64 image prefix patterns
	base64ImagePrefixes = []string{
		"data:image/jpeg;base64,",
		"data:image/jpg;base64,",
		"data:image/png;base64,",
		"data:image/gif;base64,",
		"data:image/webp;base64,",
	}

	// Base64 audio prefix patterns
	base64AudioPrefixes = []string{
		"data:audio/ogg;base64,",
		"data:audio/mpeg;base64,",
		"data:audio/mp3;base64,",
		"data:audio/aac;base64,",
		"data:audio/mp4;base64,",
	}

	// Base64 video prefix patterns
	base64VideoPrefixes = []string{
		"data:video/mp4;base64,",
		"data:video/3gpp;base64,",
		"data:video/quicktime;base64,",
	}
)

// Validator wraps go-playground validator
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new validator instance with custom validation rules
func NewValidator() *Validator {
	v := validator.New()

	// Register custom validation functions
	_ = v.RegisterValidation("e164", validateE164Phone)
	_ = v.RegisterValidation("media_url_or_base64", validateMediaURLOrBase64)

	return &Validator{validate: v}
}

// ValidateSendTextRequest validates a send text request
func (v *Validator) ValidateSendTextRequest(req *SendTextRequest) error {
	// Validate using struct tags
	if err := v.validate.Struct(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Additional validation for dual fields
	phone := req.GetPhone()
	if phone == "" {
		return fmt.Errorf("phone is required")
	}
	if !phoneRegex.MatchString(phone) {
		return fmt.Errorf("invalid phone format: must be E.164 format")
	}

	message := req.GetMessage()
	if message == "" {
		return fmt.Errorf("message is required")
	}
	if len(message) > 4096 {
		return fmt.Errorf("message too long: max 4096 characters")
	}

	return nil
}

// ValidateSendImageRequest validates a send image request
func (v *Validator) ValidateSendImageRequest(req *SendImageRequest) error {
	// Validate using struct tags
	if err := v.validate.Struct(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Additional validation for dual fields
	phone := req.GetPhone()
	if phone == "" {
		return fmt.Errorf("phone is required")
	}
	if !phoneRegex.MatchString(phone) {
		return fmt.Errorf("invalid phone format: must be E.164 format")
	}

	image := req.GetImage()
	if image == "" {
		return fmt.Errorf("image is required")
	}

	// Validate image is URL or base64
	if !isValidURL(image) && !isValidBase64Image(image) {
		return fmt.Errorf("image must be a valid URL or base64 encoded image")
	}

	caption := req.GetCaption()
	if len(caption) > 1024 {
		return fmt.Errorf("caption too long: max 1024 characters")
	}

	return nil
}

// ValidateSendAudioRequest validates a send audio request
func (v *Validator) ValidateSendAudioRequest(req *SendAudioRequest) error {
	// Validate using struct tags
	if err := v.validate.Struct(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Additional validation for dual fields
	phone := req.GetPhone()
	if phone == "" {
		return fmt.Errorf("phone is required")
	}
	if !phoneRegex.MatchString(phone) {
		return fmt.Errorf("invalid phone format: must be E.164 format")
	}

	audio := req.GetAudio()
	if audio == "" {
		return fmt.Errorf("audio is required")
	}

	// Validate audio is URL or base64
	if !isValidURL(audio) && !isValidBase64Audio(audio) {
		return fmt.Errorf("audio must be a valid URL or base64 encoded audio")
	}

	return nil
}

// ValidateSendVideoRequest validates a send video request
func (v *Validator) ValidateSendVideoRequest(req *SendVideoRequest) error {
	// Validate using struct tags
	if err := v.validate.Struct(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Additional validation for dual fields
	phone := req.GetPhone()
	if phone == "" {
		return fmt.Errorf("phone is required")
	}
	if !phoneRegex.MatchString(phone) {
		return fmt.Errorf("invalid phone format: must be E.164 format")
	}

	video := req.GetVideo()
	if video == "" {
		return fmt.Errorf("video is required")
	}

	// Validate video is URL or base64
	if !isValidURL(video) && !isValidBase64Video(video) {
		return fmt.Errorf("video must be a valid URL or base64 encoded video")
	}

	caption := req.GetCaption()
	if len(caption) > 1024 {
		return fmt.Errorf("caption too long: max 1024 characters")
	}

	return nil
}

// ValidateSendStickerRequest validates a send sticker request
func (v *Validator) ValidateSendStickerRequest(req *SendStickerRequest) error {
	// Validate using struct tags
	if err := v.validate.Struct(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Additional validation for dual fields
	phone := req.GetPhone()
	if phone == "" {
		return fmt.Errorf("phone is required")
	}
	if !phoneRegex.MatchString(phone) {
		return fmt.Errorf("invalid phone format: must be E.164 format")
	}

	sticker := req.GetSticker()
	if sticker == "" {
		return fmt.Errorf("sticker is required")
	}

	// Validate sticker is URL or base64
	if !isValidURL(sticker) && !isValidBase64Image(sticker) {
		return fmt.Errorf("sticker must be a valid URL or base64 encoded image")
	}

	return nil
}

// ValidateSendGifRequest validates a send gif request
func (v *Validator) ValidateSendGifRequest(req *SendGifRequest) error {
	// Validate using struct tags
	if err := v.validate.Struct(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Additional validation for dual fields
	phone := req.GetPhone()
	if phone == "" {
		return fmt.Errorf("phone is required")
	}
	if !phoneRegex.MatchString(phone) {
		return fmt.Errorf("invalid phone format: must be E.164 format")
	}

	gif := req.GetGif()
	if gif == "" {
		return fmt.Errorf("gif is required")
	}

	// Validate gif is URL or base64
	if !isValidURL(gif) && !isValidBase64Image(gif) {
		return fmt.Errorf("gif must be a valid URL or base64 encoded image")
	}

	caption := req.GetCaption()
	if len(caption) > 1024 {
		return fmt.Errorf("caption too long: max 1024 characters")
	}

	return nil
}

// Helper validation functions

// validateE164Phone is a custom validator for E.164 phone format
func validateE164Phone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	return phoneRegex.MatchString(phone)
}

// validateMediaURLOrBase64 is a custom validator for media URLs or base64
func validateMediaURLOrBase64(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return isValidURL(value) || isValidBase64Image(value) || isValidBase64Audio(value) || isValidBase64Video(value)
}

// isValidURL checks if a string is a valid URL
func isValidURL(str string) bool {
	if str == "" {
		return false
	}
	u, err := url.Parse(str)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

// isValidBase64Image checks if a string is a valid base64 encoded image
func isValidBase64Image(str string) bool {
	if str == "" {
		return false
	}

	// Check for data URI prefix
	for _, prefix := range base64ImagePrefixes {
		if strings.HasPrefix(str, prefix) {
			// Extract base64 content
			base64Content := strings.TrimPrefix(str, prefix)
			// Try to decode
			_, err := base64.StdEncoding.DecodeString(base64Content)
			return err == nil
		}
	}

	// Try direct base64 decode (without data URI prefix)
	_, err := base64.StdEncoding.DecodeString(str)
	return err == nil
}

// isValidBase64Audio checks if a string is a valid base64 encoded audio
func isValidBase64Audio(str string) bool {
	if str == "" {
		return false
	}

	// Check for data URI prefix
	for _, prefix := range base64AudioPrefixes {
		if strings.HasPrefix(str, prefix) {
			// Extract base64 content
			base64Content := strings.TrimPrefix(str, prefix)
			// Try to decode
			_, err := base64.StdEncoding.DecodeString(base64Content)
			return err == nil
		}
	}

	// Try direct base64 decode (without data URI prefix)
	_, err := base64.StdEncoding.DecodeString(str)
	return err == nil
}

// isValidBase64Video checks if a string is a valid base64 encoded video
func isValidBase64Video(str string) bool {
	if str == "" {
		return false
	}

	// Check for data URI prefix
	for _, prefix := range base64VideoPrefixes {
		if strings.HasPrefix(str, prefix) {
			// Extract base64 content
			base64Content := strings.TrimPrefix(str, prefix)
			// Try to decode
			_, err := base64.StdEncoding.DecodeString(base64Content)
			return err == nil
		}
	}

	// Try direct base64 decode (without data URI prefix)
	_, err := base64.StdEncoding.DecodeString(str)
	return err == nil
}

// NormalizePhone normalizes a phone number to E.164 format
func NormalizePhone(phone string) string {
	// Remove all non-digit characters except +
	cleaned := regexp.MustCompile(`[^\d+]`).ReplaceAllString(phone, "")

	// Ensure it starts with +
	if !strings.HasPrefix(cleaned, "+") {
		cleaned = "+" + cleaned
	}

	return cleaned
}
