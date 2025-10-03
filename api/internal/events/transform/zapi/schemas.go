package zapi

// Package zapi contains Z-API webhook schema definitions and transformers.
// These schemas match the Z-API webhook specifications exactly, including
// field names, types, and structures.
//
// IMPORTANT: Field names must match Z-API spec exactly, including typos
// (e.g., "momment" instead of "moment") to maintain compatibility.

// ReceivedCallback is the webhook payload for all received messages.
// This is the primary webhook type for incoming messages of all types.
type ReceivedCallback struct {
	// Common fields
	IsStatusReply    bool   `json:"isStatusReply"`
	SenderLid        string `json:"senderLid,omitempty"`
	ConnectedPhone   string `json:"connectedPhone"`
	WaitingMessage   bool   `json:"waitingMessage"`
	IsEdit           bool   `json:"isEdit"`
	IsGroup          bool   `json:"isGroup"`
	IsNewsletter     bool   `json:"isNewsletter"`
	InstanceID       string `json:"instanceId"`
	MessageID        string `json:"messageId"`
	Phone            string `json:"phone"`
	FromMe           bool   `json:"fromMe"`
	Momment          int64  `json:"momment"` // Note: typo in Z-API spec
	Status           string `json:"status"`
	ChatName         string `json:"chatName,omitempty"`
	SenderPhoto      string `json:"senderPhoto,omitempty"`
	SenderName       string `json:"senderName,omitempty"`
	ParticipantPhone string `json:"participantPhone,omitempty"`
	ParticipantLid   string `json:"participantLid,omitempty"`
	Photo            string `json:"photo,omitempty"`
	Broadcast        bool   `json:"broadcast"`
	Type             string `json:"type"` // Always "ReceivedCallback"

	// Optional fields based on message type
	ReferenceMessageID string   `json:"referenceMessageId,omitempty"`
	Forwarded          bool     `json:"forwarded,omitempty"`
	Mentioned          []string `json:"mentioned,omitempty"`
	RevokedMessageID   string   `json:"revokedMessageId,omitempty"`

	// Message content fields (only one will be present based on message type)
	Text                    *TextContent                `json:"text,omitempty"`
	Image                   *ImageContent               `json:"image,omitempty"`
	Audio                   *AudioContent               `json:"audio,omitempty"`
	Video                   *VideoContent               `json:"video,omitempty"`
	Document                *DocumentContent            `json:"document,omitempty"`
	Location                *LocationContent            `json:"location,omitempty"`
	Contact                 *ContactContent             `json:"contact,omitempty"`
	Sticker                 *StickerContent             `json:"sticker,omitempty"`
	Reaction                *ReactionContent            `json:"reaction,omitempty"`
	Poll                    *PollContent                `json:"poll,omitempty"`
	PollVote                *PollVoteContent            `json:"pollVote,omitempty"`
	ButtonsResponseMessage  *ButtonsResponseContent     `json:"buttonsResponseMessage,omitempty"`
	ListResponseMessage     *ListResponseContent        `json:"listResponseMessage,omitempty"`
	HydratedTemplate        *HydratedTemplateContent    `json:"hydratedTemplate,omitempty"`
	ButtonsMessage          *ButtonsMessageContent      `json:"buttonsMessage,omitempty"`
	PixKeyMessage           *PixKeyContent              `json:"pixKeyMessage,omitempty"`
	CarouselMessage         *CarouselContent            `json:"carouselMessage,omitempty"`
	Product                 *ProductContent             `json:"product,omitempty"`
	Order                   *OrderContent               `json:"order,omitempty"`
	ReviewAndPay            *ReviewAndPayContent        `json:"reviewAndPay,omitempty"`
	ReviewOrder             *ReviewOrderContent         `json:"reviewOrder,omitempty"`
	NewsletterAdminInvite   *NewsletterAdminInviteContent `json:"newsletterAdminInvite,omitempty"`
	PinMessage              *PinMessageContent          `json:"pinMessage,omitempty"`
	Event                   *EventContent               `json:"event,omitempty"`
	EventResponse           *EventResponseContent       `json:"eventResponse,omitempty"`
	RequestPayment          *RequestPaymentContent      `json:"requestPayment,omitempty"`
	SendPayment             *SendPaymentContent         `json:"sendPayment,omitempty"`
	ExternalAdReply         *ExternalAdReplyContent     `json:"externalAdReply,omitempty"`

	// Special notification fields
	Notification           string   `json:"notification,omitempty"`
	NotificationParameters []string `json:"notificationParameters,omitempty"`
	CallID                 string   `json:"callId,omitempty"`
	Code                   string   `json:"code,omitempty"`
	RequestMethod          string   `json:"requestMethod,omitempty"`
	ProfileName            string   `json:"profileName,omitempty"`
	UpdatedPhoto           string   `json:"updatedPhoto,omitempty"`

	// Message expiration
	MessageExpirationSeconds int `json:"messageExpirationSeconds,omitempty"`
	ViewOnce                 bool `json:"viewOnce,omitempty"`
}

// TextContent holds text message content.
type TextContent struct {
	Message      string `json:"message"`
	Description  string `json:"description,omitempty"`
	Title        string `json:"title,omitempty"`
	URL          string `json:"url,omitempty"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
}

// ImageContent holds image message content.
type ImageContent struct {
	MimeType     string `json:"mimeType"`
	ImageURL     string `json:"imageUrl"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
	Caption      string `json:"caption,omitempty"`
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
	IsGif        bool   `json:"isGif,omitempty"`
	IsAnimated   bool   `json:"isAnimated,omitempty"`
	ViewOnce     bool   `json:"viewOnce,omitempty"`
}

// AudioContent holds audio message content.
type AudioContent struct {
	MimeType string `json:"mimeType"`
	AudioURL string `json:"audioUrl"`
	PTT      bool   `json:"ptt,omitempty"` // Push-to-talk (voice message)
	Seconds  int    `json:"seconds,omitempty"`
	Waveform []byte `json:"waveform,omitempty"`
	ViewOnce bool   `json:"viewOnce,omitempty"`
}

// VideoContent holds video message content.
type VideoContent struct {
	VideoURL string `json:"videoUrl"`
	Caption  string `json:"caption,omitempty"`
	MimeType string `json:"mimeType"`
	Seconds  int    `json:"seconds,omitempty"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	IsGif    bool   `json:"isGif,omitempty"` // Video plays as GIF
	ViewOnce bool   `json:"viewOnce,omitempty"`
}

// DocumentContent holds document message content.
type DocumentContent struct {
	DocumentURL  string `json:"documentUrl"`
	MimeType     string `json:"mimeType"`
	Title        string `json:"title,omitempty"`
	PageCount    int    `json:"pageCount,omitempty"`
	FileName     string `json:"fileName,omitempty"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
	Caption      string `json:"caption,omitempty"`
}

// LocationContent holds location message content.
type LocationContent struct {
	Longitude    float64 `json:"longitude"`
	Latitude     float64 `json:"latitude"`
	Name         string  `json:"name,omitempty"`
	Address      string  `json:"address,omitempty"`
	URL          string  `json:"url,omitempty"`
	ThumbnailURL string  `json:"thumbnailUrl,omitempty"`
}

// ContactContent holds contact message content.
type ContactContent struct {
	DisplayName string   `json:"displayName"`
	VCard       string   `json:"vCard"`
	Phones      []string `json:"phones,omitempty"`
}

// StickerContent holds sticker message content.
type StickerContent struct {
	StickerURL string `json:"stickerUrl"`
	MimeType   string `json:"mimeType"`
	IsAnimated bool   `json:"isAnimated,omitempty"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
}

// ReactionContent holds reaction message content.
type ReactionContent struct {
	Value              string          `json:"value"`
	Time               int64           `json:"time"`
	ReactionBy         string          `json:"reactionBy"`
	ReferencedMessage  *MessageRef     `json:"referencedMessage"`
}

// MessageRef references another message.
type MessageRef struct {
	MessageID   string  `json:"messageId"`
	FromMe      bool    `json:"fromMe"`
	Phone       string  `json:"phone"`
	Participant *string `json:"participant"`
}

// PollContent holds poll message content.
type PollContent struct {
	Question       string        `json:"question"`
	PollMaxOptions int           `json:"pollMaxOptions"`
	Options        []PollOption  `json:"options"`
}

// PollOption is a poll option.
type PollOption struct {
	Name string `json:"name"`
}

// PollVoteContent holds poll vote content.
type PollVoteContent struct {
	PollMessageID string       `json:"pollMessageId"`
	Options       []PollOption `json:"options"`
}

// ButtonsResponseContent holds button response content.
type ButtonsResponseContent struct {
	ButtonID string `json:"buttonId"`
	Message  string `json:"message"`
}

// ListResponseContent holds list response content.
type ListResponseContent struct {
	Message       string `json:"message"`
	Title         string `json:"title"`
	SelectedRowID string `json:"selectedRowId"`
}

// HydratedTemplateContent holds hydrated template content.
type HydratedTemplateContent struct {
	Header          *TemplateHeader   `json:"header,omitempty"`
	Message         string            `json:"message"`
	Footer          string            `json:"footer,omitempty"`
	Title           string            `json:"title,omitempty"`
	TemplateID      string            `json:"templateId,omitempty"`
	HydratedButtons []HydratedButton  `json:"hydratedButtons,omitempty"`
}

// TemplateHeader is a template header.
type TemplateHeader struct {
	Image    *ImageContent    `json:"image,omitempty"`
	Video    *VideoContent    `json:"video,omitempty"`
	Document *DocumentContent `json:"document,omitempty"`
	Location *LocationContent `json:"location,omitempty"`
}

// HydratedButton is a hydrated button.
type HydratedButton struct {
	Index            int               `json:"index"`
	URLButton        *URLButton        `json:"urlButton,omitempty"`
	QuickReplyButton *QuickReplyButton `json:"quickReplyButton,omitempty"`
}

// URLButton is a URL button.
type URLButton struct {
	DisplayText string `json:"displayText"`
	URL         string `json:"url"`
}

// QuickReplyButton is a quick reply button.
type QuickReplyButton struct {
	DisplayText string `json:"displayText"`
	ID          string `json:"id"`
}

// ButtonsMessageContent holds buttons message content.
type ButtonsMessageContent struct {
	ImageURL string   `json:"imageUrl,omitempty"`
	VideoURL string   `json:"videoUrl,omitempty"`
	Message  string   `json:"message"`
	Buttons  []Button `json:"buttons"`
}

// Button is a button.
type Button struct {
	ButtonID   string      `json:"buttonId"`
	Type       int         `json:"type"`
	ButtonText *ButtonText `json:"buttonText"`
}

// ButtonText is button text.
type ButtonText struct {
	DisplayText string `json:"displayText"`
}

// PixKeyContent holds PIX key content.
type PixKeyContent struct {
	Currency     string `json:"currency"`
	ReferenceID  string `json:"referenceId"`
	Key          string `json:"key"`
	KeyType      string `json:"keyType"`
	MerchantName string `json:"merchantName"`
}

// CarouselContent holds carousel content.
type CarouselContent struct {
	Text  string         `json:"text"`
	Cards []CarouselCard `json:"cards"`
}

// CarouselCard is a carousel card.
type CarouselCard struct {
	Header          *TemplateHeader  `json:"header,omitempty"`
	Message         string           `json:"message"`
	Footer          string           `json:"footer,omitempty"`
	Title           string           `json:"title,omitempty"`
	HydratedButtons []HydratedButton `json:"hydratedButtons,omitempty"`
}

// ProductContent holds product content.
type ProductContent struct {
	ProductImage      string  `json:"productImage"`
	BusinessOwnerJID  string  `json:"businessOwnerJid"`
	CurrencyCode      string  `json:"currencyCode"`
	ProductID         string  `json:"productId"`
	Description       string  `json:"description,omitempty"`
	ProductImageCount int     `json:"productImageCount"`
	Price             float64 `json:"price"`
	URL               string  `json:"url,omitempty"`
	RetailerID        string  `json:"retailerId,omitempty"`
	FirstImageID      string  `json:"firstImageId,omitempty"`
	Title             string  `json:"title"`
}

// OrderContent holds order content.
type OrderContent struct {
	ItemCount    int            `json:"itemCount"`
	OrderID      string         `json:"orderId"`
	Message      string         `json:"message,omitempty"`
	OrderTitle   string         `json:"orderTitle"`
	SellerJID    string         `json:"sellerJid"`
	ThumbnailURL string         `json:"thumbnailUrl,omitempty"`
	Token        string         `json:"token"`
	Currency     string         `json:"currency"`
	Total        int64          `json:"total"`
	SubTotal     int64          `json:"subTotal"`
	Products     []OrderProduct `json:"products"`
}

// OrderProduct is an order product.
type OrderProduct struct {
	Quantity     int    `json:"quantity"`
	Name         string `json:"name"`
	ProductID    string `json:"productId"`
	RetailerID   string `json:"retailerId"`
	Price        int64  `json:"price"`
	CurrencyCode string `json:"currencyCode"`
}

// ReviewAndPayContent holds review and pay content.
type ReviewAndPayContent struct {
	Type            string         `json:"type"`
	Currency        string         `json:"currency"`
	ReferenceID     string         `json:"referenceId"`
	OrderRequestID  string         `json:"orderRequestId"`
	OrderStatus     string         `json:"orderStatus"`
	PaymentStatus   string         `json:"paymentStatus"`
	Total           int64          `json:"total"`
	SubTotal        int64          `json:"subTotal"`
	Discount        int64          `json:"discount,omitempty"`
	Shipping        int64          `json:"shipping,omitempty"`
	Tax             int64          `json:"tax,omitempty"`
	Products        []OrderProduct `json:"products"`
}

// ReviewOrderContent holds review order content.
type ReviewOrderContent struct {
	Currency       string         `json:"currency"`
	ReferenceID    string         `json:"referenceId"`
	OrderRequestID string         `json:"orderRequestId"`
	OrderStatus    string         `json:"orderStatus"`
	PaymentStatus  string         `json:"paymentStatus"`
	Total          int64          `json:"total"`
	SubTotal       int64          `json:"subTotal"`
	Discount       int64          `json:"discount,omitempty"`
	Shipping       int64          `json:"shipping,omitempty"`
	Tax            int64          `json:"tax,omitempty"`
	Products       []OrderProduct `json:"products"`
}

// NewsletterAdminInviteContent holds newsletter admin invite content.
type NewsletterAdminInviteContent struct {
	NewsletterID     string `json:"newsletterId"`
	NewsletterName   string `json:"newsletterName"`
	Text             string `json:"text"`
	InviteExpiration int64  `json:"inviteExpiration"`
}

// PinMessageContent holds pin message content.
type PinMessageContent struct {
	Action             string      `json:"action"`
	PinDurationInSecs  int         `json:"pinDurationInSecs,omitempty"`
	ReferencedMessage  *MessageRef `json:"referencedMessage"`
}

// EventContent holds event content.
type EventContent struct {
	Name         string            `json:"name"`
	Description  string            `json:"description,omitempty"`
	Canceled     bool              `json:"canceled"`
	JoinLink     string            `json:"joinLink,omitempty"`
	ScheduleTime int64             `json:"scheduleTime"`
	Location     map[string]string `json:"location,omitempty"`
}

// EventResponseContent holds event response content.
type EventResponseContent struct {
	Response          string      `json:"response"`
	ResponseFrom      string      `json:"responseFrom"`
	Time              int64       `json:"time"`
	ReferencedMessage *MessageRef `json:"referencedMessage"`
}

// RequestPaymentContent holds request payment content.
type RequestPaymentContent struct {
	Value        float64      `json:"value"`
	CurrencyCode string       `json:"currencyCode"`
	Expiration   int64        `json:"expiration"`
	RequestPhone string       `json:"requestPhone"`
	PaymentInfo  *PaymentInfo `json:"paymentInfo,omitempty"`
}

// SendPaymentContent holds send payment content.
type SendPaymentContent struct {
	PaymentInfo *PaymentInfo `json:"paymentInfo"`
}

// PaymentInfo holds payment information.
type PaymentInfo struct {
	ReceiverPhone     string  `json:"receiverPhone"`
	Value             float64 `json:"value"`
	CurrencyCode      string  `json:"currencyCode"`
	Status            string  `json:"status"`
	TransactionStatus string  `json:"transactionStatus"`
}

// ExternalAdReplyContent holds external ad reply content.
type ExternalAdReplyContent struct {
	Title                  string `json:"title"`
	Body                   string `json:"body"`
	MediaType              int    `json:"mediaType"`
	ThumbnailURL           string `json:"thumbnailUrl,omitempty"`
	SourceType             string `json:"sourceType"`
	SourceID               string `json:"sourceId"`
	SourceURL              string `json:"sourceUrl,omitempty"`
	ContainsAutoReply      bool   `json:"containsAutoReply"`
	RenderLargerThumbnail  bool   `json:"renderLargerThumbnail"`
	ShowAdAttribution      bool   `json:"showAdAttribution"`
}

// MessageStatusCallback is the webhook payload for message status updates.
type MessageStatusCallback struct {
	InstanceID  string   `json:"instanceId"`
	Status      string   `json:"status"` // SENT, RECEIVED, READ, PLAYED
	IDs         []string `json:"ids"`
	Momment     int64    `json:"momment"`
	PhoneDevice int      `json:"phoneDevice,omitempty"`
	Phone       string   `json:"phone"`
	Type        string   `json:"type"` // Always "MessageStatusCallback"
	IsGroup     bool     `json:"isGroup"`
}

// PresenceChatCallback is the webhook payload for chat presence updates.
type PresenceChatCallback struct {
	Type       string  `json:"type"` // Always "PresenceChatCallback"
	Phone      string  `json:"phone"`
	Status     string  `json:"status"` // UNAVAILABLE, AVAILABLE, COMPOSING, PAUSED, RECORDING
	LastSeen   *int64  `json:"lastSeen"`
	InstanceID string  `json:"instanceId"`
}

// ConnectedCallback is the webhook payload for connection events.
type ConnectedCallback struct {
	Type       string `json:"type"` // Always "ConnectedCallback"
	Connected  bool   `json:"connected"`
	Momment    int64  `json:"momment"`
	InstanceID string `json:"instanceId"`
	Phone      string `json:"phone,omitempty"`
}

// DisconnectedCallback is the webhook payload for disconnection events.
type DisconnectedCallback struct {
	Momment      int64  `json:"momment"`
	Error        string `json:"error,omitempty"`
	Disconnected bool   `json:"disconnected"`
	Type         string `json:"type"` // Always "DisconnectedCallback"
	InstanceID   string `json:"instanceId"`
}

// DeliveryCallback is the webhook payload for message delivery confirmations.
type DeliveryCallback struct {
	Phone      string `json:"phone"`
	ZaapID     string `json:"zaapId"`
	MessageID  string `json:"messageId"`
	Type       string `json:"type"` // Always "DeliveryCallback"
	InstanceID string `json:"instanceId"`
}
