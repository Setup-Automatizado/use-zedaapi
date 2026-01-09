package interactive

import "strings"

// ButtonType defines the type of button action
type ButtonType string

const (
	ButtonTypeQuickReply   ButtonType = "quick_reply"
	ButtonTypeCTAURL       ButtonType = "cta_url"
	ButtonTypeCTACall      ButtonType = "cta_call"
	ButtonTypeCTACopy      ButtonType = "cta_copy"
	ButtonTypePaymentInfo  ButtonType = "payment_info"
	ButtonTypeReviewAndPay ButtonType = "review_and_pay"
)

// NormalizeButtonType converts FUNNELCHAT uppercase format to internal format
// Accepts both FUNNELCHAT style (CALL, URL, COPY) and internal style (cta_call, cta_url)
func NormalizeButtonType(t string) ButtonType {
	switch strings.ToUpper(t) {
	case "CALL", "CTA_CALL":
		return ButtonTypeCTACall
	case "URL", "CTA_URL":
		return ButtonTypeCTAURL
	case "COPY", "CTA_COPY":
		return ButtonTypeCTACopy
	case "QUICK_REPLY", "REPLY":
		return ButtonTypeQuickReply
	case "PAYMENT_INFO":
		return ButtonTypePaymentInfo
	case "REVIEW_AND_PAY":
		return ButtonTypeReviewAndPay
	default:
		return ButtonType(strings.ToLower(t))
	}
}

// PIXKeyType defines PIX key types for Brazilian payment
type PIXKeyType string

const (
	PIXKeyTypeCPF   PIXKeyType = "CPF"
	PIXKeyTypeCNPJ  PIXKeyType = "CNPJ"
	PIXKeyTypeEmail PIXKeyType = "EMAIL"
	PIXKeyTypePhone PIXKeyType = "PHONE"
	PIXKeyTypeEVP   PIXKeyType = "EVP" // Random key
)

// SendButtonListRequest - FUNNELCHAT /send-button-list format
type SendButtonListRequest struct {
	Phone        string            `json:"phone" validate:"required"`
	Message      string            `json:"message" validate:"required,max=4096"`
	ButtonList   ButtonListPayload `json:"buttonList" validate:"required"`
	Title        *string           `json:"title,omitempty"`
	Footer       *string           `json:"footer,omitempty"`
	Image        *string           `json:"image,omitempty"`        // URL or base64
	Video        *string           `json:"video,omitempty"`        // URL or base64
	DelayMessage *int              `json:"delayMessage,omitempty"` // Delay in milliseconds
	MessageId    *string           `json:"messageId,omitempty"`    // Custom tracking ID
}

// ButtonListPayload contains the list of buttons for button list messages
type ButtonListPayload struct {
	Buttons []ButtonListItem `json:"buttons" validate:"required,min=1,max=3,dive"`
}

// ButtonListItem represents a single button in button list
type ButtonListItem struct {
	ID    string `json:"id" validate:"required,max=256"`
	Label string `json:"label" validate:"required,max=20"`
}

// SendButtonActionsRequest - FUNNELCHAT /send-button-actions format
type SendButtonActionsRequest struct {
	Phone         string               `json:"phone" validate:"required"`
	Message       string               `json:"message" validate:"required,max=4096"`
	Title         *string              `json:"title,omitempty" validate:"omitempty,max=60"`
	Footer        *string              `json:"footer,omitempty" validate:"omitempty,max=60"`
	ButtonActions ButtonActionsPayload `json:"buttonActions" validate:"required"`
	DelayMessage  *int                 `json:"delayMessage,omitempty"`
	MessageId     *string              `json:"messageId,omitempty"`
}

// ButtonActionsPayload contains the list of action buttons
type ButtonActionsPayload struct {
	Buttons []ActionButton `json:"buttons" validate:"required,min=1,max=3,dive"`
}

// ActionButton represents a single action button with type-specific fields
type ActionButton struct {
	ID       string     `json:"id" validate:"required,max=256"`
	Label    string     `json:"label" validate:"required,max=20"`
	Type     ButtonType `json:"type" validate:"required"`
	URL      *string    `json:"url,omitempty" validate:"omitempty,url"`
	Phone    *string    `json:"phone,omitempty"`
	CopyCode *string    `json:"copyCode,omitempty"`
	// Payment button fields
	Currency      *string       `json:"currency,omitempty"`      // Currency code (default: BRL)
	Amount        *int64        `json:"amount,omitempty"`        // Amount in cents
	TotalAmount   *int64        `json:"totalAmount,omitempty"`   // Total amount for review_and_pay
	ReferenceID   *string       `json:"referenceId,omitempty"`   // Payment reference
	PaymentStatus *string       `json:"paymentStatus,omitempty"` // pending, completed, failed
	MerchantName  *string       `json:"merchantName,omitempty"`  // Merchant name for payments
	Order         *PaymentOrder `json:"order,omitempty"`         // Order details for review_and_pay
	ExpiresAt     *int64        `json:"expiresAt,omitempty"`     // Expiration timestamp
	// PIX-specific fields for payment_info and review_and_pay buttons
	PIXKey     *string `json:"pixKey,omitempty"`     // PIX key value
	PIXKeyType *string `json:"pixKeyType,omitempty"` // PIX key type: PHONE, CPF, CNPJ, EMAIL, EVP
}

// GetNormalizedType returns the normalized button type
// Accepts both FUNNELCHAT uppercase (CALL, URL) and internal lowercase (cta_call, cta_url)
func (b *ActionButton) GetNormalizedType() ButtonType {
	return NormalizeButtonType(string(b.Type))
}

// PaymentOrder represents order details for review_and_pay buttons
type PaymentOrder struct {
	ID    string        `json:"id"`
	Items []PaymentItem `json:"items"`
}

// PaymentItem represents an item in a payment order
type PaymentItem struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Price    int64  `json:"price"` // Price in cents
}

// SendOptionListRequest - FUNNELCHAT /send-option-list format
type SendOptionListRequest struct {
	Phone        string            `json:"phone" validate:"required"`
	Message      string            `json:"message" validate:"required,max=4096"`
	Title        *string           `json:"title,omitempty" validate:"omitempty,max=60"`
	Footer       *string           `json:"footer,omitempty" validate:"omitempty,max=60"`
	ButtonLabel  string            `json:"buttonLabel" validate:"required,max=20"`
	OptionList   OptionListPayload `json:"optionList" validate:"required"`
	DelayMessage *int              `json:"delayMessage,omitempty"`
	MessageId    *string           `json:"messageId,omitempty"`
}

// OptionListPayload contains sections for list messages
type OptionListPayload struct {
	Sections []OptionSection `json:"sections" validate:"required,min=1,max=10,dive"`
}

// OptionSection represents a section in option list
type OptionSection struct {
	Title string      `json:"title" validate:"required,max=24"`
	Rows  []OptionRow `json:"rows" validate:"required,min=1,max=10,dive"`
}

// OptionRow represents a row in option list section
type OptionRow struct {
	ID          string  `json:"id" validate:"required,max=200"`
	Title       string  `json:"title" validate:"required,max=24"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=72"`
}

// SendButtonPIXRequest - FUNNELCHAT /send-button-pix format
type SendButtonPIXRequest struct {
	Phone         string     `json:"phone" validate:"required"`
	Message       *string    `json:"message,omitempty" validate:"omitempty,max=4096"`
	PIXKey        string     `json:"pixKey" validate:"required"`
	Type          PIXKeyType `json:"type" validate:"required,oneof=CPF CNPJ EMAIL PHONE EVP"`
	Name          *string    `json:"name,omitempty"`          // Beneficiary name (merchant name)
	TransactionID *string    `json:"transactionId,omitempty"` // PIX transaction ID
	Amount        *float64   `json:"amount,omitempty"`        // Amount in BRL
	Title         *string    `json:"title,omitempty" validate:"omitempty,max=60"`
	Footer        *string    `json:"footer,omitempty" validate:"omitempty,max=60"`
	DelayMessage  *int       `json:"delayMessage,omitempty"`
	MessageId     *string    `json:"messageId,omitempty"`
}

// SendButtonOTPRequest - FUNNELCHAT /send-button-otp format
type SendButtonOTPRequest struct {
	Phone        string  `json:"phone" validate:"required"`
	Message      string  `json:"message" validate:"required,max=4096"`
	Code         string  `json:"code" validate:"required,max=20"`
	Title        *string `json:"title,omitempty" validate:"omitempty,max=60"`
	Footer       *string `json:"footer,omitempty" validate:"omitempty,max=60"`
	DelayMessage *int    `json:"delayMessage,omitempty"`
	MessageId    *string `json:"messageId,omitempty"`
}

// Carousel types for /send-carousel endpoint

// CarouselCardType defines the carousel display type
type CarouselCardType string

const (
	CarouselCardTypeHScroll CarouselCardType = "HSCROLL_CARDS"
	CarouselCardTypeAlbum   CarouselCardType = "ALBUM_IMAGE"
)

// SendCarouselRequest - FUNNELCHAT /send-carousel format
type SendCarouselRequest struct {
	Phone        string           `json:"phone" validate:"required"`
	Cards        []CarouselCard   `json:"cards" validate:"required,min=1,max=10,dive"`
	CardType     CarouselCardType `json:"cardType,omitempty"` // defaults to HSCROLL_CARDS
	DelayMessage *int             `json:"delayMessage,omitempty"`
	MessageId    *string          `json:"messageId,omitempty"`
}

// CarouselCard represents a single card in a carousel message
type CarouselCard struct {
	Header      *CarouselHeader `json:"header,omitempty"`
	Body        CarouselBody    `json:"body" validate:"required"`
	Footer      *CarouselFooter `json:"footer,omitempty"`
	Buttons     []ActionButton  `json:"buttons" validate:"required,min=1,max=3,dive"`
	MediaURL    *string         `json:"mediaUrl,omitempty"`    // Card image/video URL
	MediaBase64 *string         `json:"mediaBase64,omitempty"` // Alternative to URL
	MediaType   *string         `json:"mediaType,omitempty"`   // image or video
}

// CarouselHeader represents the header section of a carousel card
type CarouselHeader struct {
	Text string `json:"text,omitempty" validate:"max=60"`
}

// CarouselBody represents the body section of a carousel card
type CarouselBody struct {
	Text string `json:"text" validate:"required,max=1024"`
}

// CarouselFooter represents the footer section of a carousel card
type CarouselFooter struct {
	Text string `json:"text,omitempty" validate:"max=60"`
}

// ContextInfo provides context for quoted/forwarded messages
type ContextInfo struct {
	StanzaID        *string        `json:"stanzaId,omitempty"`        // Message ID being replied to
	Participant     *string        `json:"participant,omitempty"`     // Participant JID
	QuotedMessage   *QuotedMessage `json:"quotedMessage,omitempty"`   // Quoted message content
	IsForwarded     bool           `json:"isForwarded,omitempty"`     // Whether message is forwarded
	ForwardingScore int            `json:"forwardingScore,omitempty"` // Forwarding count
}

// QuotedMessage represents the content of a quoted message
type QuotedMessage struct {
	Conversation string `json:"conversation,omitempty"` // Text content of quoted message
}

// ========================================
// PIX Payment Proto Structures (whatsmeow pattern)
// ========================================

// PIXTotalAmount represents amount with value and offset (decimal places)
// Example: R$ 100.00 = {Value: 10000, Offset: 100} or {Value: 100000, Offset: 1000}
type PIXTotalAmount struct {
	Value  int `json:"value"`
	Offset int `json:"offset"`
}

// PIXStaticCode contains PIX key information
type PIXStaticCode struct {
	Key          string `json:"key"`
	MerchantName string `json:"merchant_name"`
	KeyType      string `json:"key_type"` // CPF, CNPJ, EMAIL, PHONE, EVP
}

// PIXCards represents card payment settings (usually disabled for PIX)
type PIXCards struct {
	Enabled bool `json:"enabled"`
}

// PIXPaymentSettings represents payment configuration for PIX
type PIXPaymentSettings struct {
	Type          string         `json:"type"` // "pix_static_code"
	PixStaticCode *PIXStaticCode `json:"pix_static_code,omitempty"`
	Cards         *PIXCards      `json:"cards,omitempty"`
}

// PIXItem represents an item in a PIX order
type PIXItem struct {
	Amount     PIXTotalAmount `json:"amount"`
	Name       string         `json:"name"`
	RetailerID string         `json:"retailer_id"`
	Quantity   int            `json:"quantity"`
}

// PIXOrder represents the order structure for PIX payments
type PIXOrder struct {
	Status    string         `json:"status"` // "payment_requested" or "pending"
	Items     []PIXItem      `json:"items"`
	Subtotal  PIXTotalAmount `json:"subtotal"`
	OrderType *string        `json:"order_type,omitempty"` // "ORDER_WITHOUT_AMOUNT" for PIX
}

// PIXButtonParamsJSON is the complete structure for payment_info button
// This matches the whatsmeow ButtonPixOrderParamsJSON pattern
type PIXButtonParamsJSON struct {
	Currency        string               `json:"currency"`
	Type            string               `json:"type"` // "physical-goods"
	TotalAmount     PIXTotalAmount       `json:"total_amount"`
	ReferenceID     string               `json:"reference_id"`
	PaymentSettings []PIXPaymentSettings `json:"payment_settings"`
	Order           PIXOrder             `json:"order"`
}

// ReviewAndPayParamsJSON is the complete structure for review_and_pay button
// This matches the whatsmeow ButtonReviewAndPayParamsJSON pattern
type ReviewAndPayParamsJSON struct {
	ReferenceID          string               `json:"reference_id"`
	Type                 string               `json:"type"`                  // "physical-goods"
	PaymentType          string               `json:"payment_type"`          // "br"
	PaymentConfiguration string               `json:"payment_configuration"` // "merchant_categorization_code"
	PaymentSettings      []PIXPaymentSettings `json:"payment_settings"`
	Currency             string               `json:"currency"`
	TotalAmount          PIXTotalAmount       `json:"total_amount"`
	Order                PIXOrder             `json:"order"`
	SharePaymentStatus   bool                 `json:"share_payment_status"`
	Referral             string               `json:"referral"` // "chat_attachment"
}

// InteractiveMessageResponse is the standard response for queued interactive messages
type InteractiveMessageResponse struct {
	Status    string  `json:"status"`              // QUEUED, SENT, FAILED
	ZapiID    string  `json:"zapiId,omitempty"`    // ID
	QueueID   string  `json:"queueId,omitempty"`   // Internal queue ID (deprecated, use zapiId)
	MessageID *string `json:"messageId,omitempty"` // Custom message ID if provided
}
