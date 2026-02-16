package messages

import (
	"time"

	"github.com/google/uuid"
)

// SendTextRequest represents Zé da API send-text request with dual field names for compatibility
type SendTextRequest struct {
	Phone        string `json:"phone" validate:"required,e164"`
	PhoneAlt     string `json:"Phone,omitempty"`
	Message      string `json:"message" validate:"required,min=1"`
	MessageAlt   string `json:"Message,omitempty"`
	DelayMessage int    `json:"delayMessage,omitempty"`
	DelayAlt     int    `json:"DelayMessage,omitempty"`
}

// GetPhone returns the phone number (supporting dual field names)
func (r *SendTextRequest) GetPhone() string {
	if r.Phone != "" {
		return r.Phone
	}
	return r.PhoneAlt
}

// GetMessage returns the message text (supporting dual field names)
func (r *SendTextRequest) GetMessage() string {
	if r.Message != "" {
		return r.Message
	}
	return r.MessageAlt
}

// GetDelay returns the delay in milliseconds (supporting dual field names)
func (r *SendTextRequest) GetDelay() int {
	if r.DelayMessage > 0 {
		return r.DelayMessage
	}
	return r.DelayAlt
}

// Media represents media attachment information
type Media struct {
	Image       string `json:"image,omitempty"`
	ImageAlt    string `json:"Image,omitempty"`
	Audio       string `json:"audio,omitempty"`
	AudioAlt    string `json:"Audio,omitempty"`
	Video       string `json:"video,omitempty"`
	VideoAlt    string `json:"Video,omitempty"`
	Document    string `json:"document,omitempty"`
	DocumentAlt string `json:"Document,omitempty"`
	Caption     string `json:"caption,omitempty"`
	CaptionAlt  string `json:"Caption,omitempty"`
	FileName    string `json:"fileName,omitempty"`
	FileNameAlt string `json:"FileName,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
	MimeTypeAlt string `json:"MimeType,omitempty"`
}

// GetImage returns the image URL or base64 (supporting dual field names)
func (m *Media) GetImage() string {
	if m.Image != "" {
		return m.Image
	}
	return m.ImageAlt
}

// GetAudio returns the audio URL or base64 (supporting dual field names)
func (m *Media) GetAudio() string {
	if m.Audio != "" {
		return m.Audio
	}
	return m.AudioAlt
}

// GetVideo returns the video URL or base64 (supporting dual field names)
func (m *Media) GetVideo() string {
	if m.Video != "" {
		return m.Video
	}
	return m.VideoAlt
}

// GetDocument returns the document URL or base64 (supporting dual field names)
func (m *Media) GetDocument() string {
	if m.Document != "" {
		return m.Document
	}
	return m.DocumentAlt
}

// GetCaption returns the caption (supporting dual field names)
func (m *Media) GetCaption() string {
	if m.Caption != "" {
		return m.Caption
	}
	return m.CaptionAlt
}

// GetFileName returns the file name (supporting dual field names)
func (m *Media) GetFileName() string {
	if m.FileName != "" {
		return m.FileName
	}
	return m.FileNameAlt
}

// GetMimeType returns the MIME type (supporting dual field names)
func (m *Media) GetMimeType() string {
	if m.MimeType != "" {
		return m.MimeType
	}
	return m.MimeTypeAlt
}

// SendImageRequest represents Zé da API send-image request
type SendImageRequest struct {
	Phone        string `json:"phone" validate:"required,e164"`
	PhoneAlt     string `json:"Phone,omitempty"`
	Image        string `json:"image" validate:"required"`
	ImageAlt     string `json:"Image,omitempty"`
	Caption      string `json:"caption,omitempty"`
	CaptionAlt   string `json:"Caption,omitempty"`
	DelayMessage int    `json:"delayMessage,omitempty"`
	DelayAlt     int    `json:"DelayMessage,omitempty"`
}

// GetPhone returns the phone number (supporting dual field names)
func (r *SendImageRequest) GetPhone() string {
	if r.Phone != "" {
		return r.Phone
	}
	return r.PhoneAlt
}

// GetImage returns the image URL or base64 (supporting dual field names)
func (r *SendImageRequest) GetImage() string {
	if r.Image != "" {
		return r.Image
	}
	return r.ImageAlt
}

// GetCaption returns the caption (supporting dual field names)
func (r *SendImageRequest) GetCaption() string {
	if r.Caption != "" {
		return r.Caption
	}
	return r.CaptionAlt
}

// GetDelay returns the delay in milliseconds (supporting dual field names)
func (r *SendImageRequest) GetDelay() int {
	if r.DelayMessage > 0 {
		return r.DelayMessage
	}
	return r.DelayAlt
}

// SendAudioRequest represents Zé da API send-audio request
type SendAudioRequest struct {
	Phone        string `json:"phone" validate:"required,e164"`
	PhoneAlt     string `json:"Phone,omitempty"`
	Audio        string `json:"audio" validate:"required"`
	AudioAlt     string `json:"Audio,omitempty"`
	DelayMessage int    `json:"delayMessage,omitempty"`
	DelayAlt     int    `json:"DelayMessage,omitempty"`
}

// GetPhone returns the phone number (supporting dual field names)
func (r *SendAudioRequest) GetPhone() string {
	if r.Phone != "" {
		return r.Phone
	}
	return r.PhoneAlt
}

// GetAudio returns the audio URL or base64 (supporting dual field names)
func (r *SendAudioRequest) GetAudio() string {
	if r.Audio != "" {
		return r.Audio
	}
	return r.AudioAlt
}

// GetDelay returns the delay in milliseconds (supporting dual field names)
func (r *SendAudioRequest) GetDelay() int {
	if r.DelayMessage > 0 {
		return r.DelayMessage
	}
	return r.DelayAlt
}

// SendVideoRequest represents Zé da API send-video request
type SendVideoRequest struct {
	Phone        string `json:"phone" validate:"required,e164"`
	PhoneAlt     string `json:"Phone,omitempty"`
	Video        string `json:"video" validate:"required"`
	VideoAlt     string `json:"Video,omitempty"`
	Caption      string `json:"caption,omitempty"`
	CaptionAlt   string `json:"Caption,omitempty"`
	DelayMessage int    `json:"delayMessage,omitempty"`
	DelayAlt     int    `json:"DelayMessage,omitempty"`
}

// GetPhone returns the phone number (supporting dual field names)
func (r *SendVideoRequest) GetPhone() string {
	if r.Phone != "" {
		return r.Phone
	}
	return r.PhoneAlt
}

// GetVideo returns the video URL or base64 (supporting dual field names)
func (r *SendVideoRequest) GetVideo() string {
	if r.Video != "" {
		return r.Video
	}
	return r.VideoAlt
}

// GetCaption returns the caption (supporting dual field names)
func (r *SendVideoRequest) GetCaption() string {
	if r.Caption != "" {
		return r.Caption
	}
	return r.CaptionAlt
}

// GetDelay returns the delay in milliseconds (supporting dual field names)
func (r *SendVideoRequest) GetDelay() int {
	if r.DelayMessage > 0 {
		return r.DelayMessage
	}
	return r.DelayAlt
}

// SendStickerRequest represents Zé da API send-sticker request
type SendStickerRequest struct {
	Phone        string `json:"phone" validate:"required,e164"`
	PhoneAlt     string `json:"Phone,omitempty"`
	Sticker      string `json:"sticker" validate:"required"`
	StickerAlt   string `json:"Sticker,omitempty"`
	DelayMessage int    `json:"delayMessage,omitempty"`
	DelayAlt     int    `json:"DelayMessage,omitempty"`
}

// GetPhone returns the phone number (supporting dual field names)
func (r *SendStickerRequest) GetPhone() string {
	if r.Phone != "" {
		return r.Phone
	}
	return r.PhoneAlt
}

// GetSticker returns the sticker URL or base64 (supporting dual field names)
func (r *SendStickerRequest) GetSticker() string {
	if r.Sticker != "" {
		return r.Sticker
	}
	return r.StickerAlt
}

// GetDelay returns the delay in milliseconds (supporting dual field names)
func (r *SendStickerRequest) GetDelay() int {
	if r.DelayMessage > 0 {
		return r.DelayMessage
	}
	return r.DelayAlt
}

// SendGifRequest represents Zé da API send-gif request
type SendGifRequest struct {
	Phone        string `json:"phone" validate:"required,e164"`
	PhoneAlt     string `json:"Phone,omitempty"`
	Gif          string `json:"gif" validate:"required"`
	GifAlt       string `json:"Gif,omitempty"`
	Caption      string `json:"caption,omitempty"`
	CaptionAlt   string `json:"Caption,omitempty"`
	DelayMessage int    `json:"delayMessage,omitempty"`
	DelayAlt     int    `json:"DelayMessage,omitempty"`
}

// GetPhone returns the phone number (supporting dual field names)
func (r *SendGifRequest) GetPhone() string {
	if r.Phone != "" {
		return r.Phone
	}
	return r.PhoneAlt
}

// GetGif returns the GIF URL or base64 (supporting dual field names)
func (r *SendGifRequest) GetGif() string {
	if r.Gif != "" {
		return r.Gif
	}
	return r.GifAlt
}

// GetCaption returns the caption (supporting dual field names)
func (r *SendGifRequest) GetCaption() string {
	if r.Caption != "" {
		return r.Caption
	}
	return r.CaptionAlt
}

// GetDelay returns the delay in milliseconds (supporting dual field names)
func (r *SendGifRequest) GetDelay() int {
	if r.DelayMessage > 0 {
		return r.DelayMessage
	}
	return r.DelayAlt
}

// SendMessageResult represents the response for send message operations
type SendMessageResult struct {
	ZaapID      string    `json:"zaapId"`
	MessageID   string    `json:"messageId"`
	ID          string    `json:"id"`
	Status      string    `json:"status,omitempty"`
	QueuedAt    time.Time `json:"queuedAt,omitempty"`
	DeliveredAt time.Time `json:"deliveredAt,omitempty"`
}

// QueuedMessage represents an internal message in the queue
type QueuedMessage struct {
	ID             uuid.UUID  `json:"id"`
	InstanceID     uuid.UUID  `json:"instance_id"`
	MessageID      string     `json:"message_id"`
	Type           string     `json:"type"`
	Phone          string     `json:"phone"`
	Payload        []byte     `json:"payload"`
	Status         string     `json:"status"`
	Attempts       int        `json:"attempts"`
	MaxAttempts    int        `json:"max_attempts"`
	ScheduledAt    time.Time  `json:"scheduled_at"`
	NextAttemptAt  time.Time  `json:"next_attempt_at"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
	LastError      string     `json:"last_error,omitempty"`
	SequenceNumber int64      `json:"sequence_number"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
