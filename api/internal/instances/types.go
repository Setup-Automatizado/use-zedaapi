package instances

import (
	"time"

	"github.com/google/uuid"
)

type Instance struct {
	ID                 uuid.UUID        `json:"instanceId"`
	Name               string           `json:"name"`
	SessionName        string           `json:"sessionName"`
	InstanceToken      string           `json:"instanceToken"`
	StoreJID           *string          `json:"storeJid,omitempty"`
	IsDevice           bool             `json:"isDevice"`
	BusinessDevice     bool             `json:"businessDevice"`
	SubscriptionActive bool             `json:"subscriptionActive"`
	CallRejectAuto     bool             `json:"callRejectAuto"`
	CallRejectMessage  *string          `json:"callRejectMessage,omitempty"`
	AutoReadMessage    bool             `json:"autoReadMessage"`
	CanceledAt         *time.Time       `json:"canceledAt,omitempty"`
	CreatedAt          time.Time        `json:"createdAt"`
	UpdatedAt          time.Time        `json:"updatedAt"`
	Webhooks           *WebhookSettings `json:"webhooks,omitempty"`
	Middleware         string           `json:"middleware"`
	PhoneConnected     bool             `json:"phoneConnected"`
	WhatsappConnected  bool             `json:"whatsappConnected"`
	Due                *time.Time       `json:"due,omitempty"`
	ConnectionStatus   string           `json:"connectionStatus,omitempty"`
	LastConnectedAt    *time.Time       `json:"lastConnectedAt,omitempty"`
	WorkerID           *string          `json:"workerId,omitempty"`
	DesiredWorkerID    *string          `json:"desiredWorkerId,omitempty"`
}

type WebhookSettings struct {
	DeliveryURL         *string `json:"deliveryCallbackUrl,omitempty"`
	ReceivedURL         *string `json:"receivedCallbackUrl,omitempty"`
	ReceivedDeliveryURL *string `json:"receivedAndDeliveryCallbackUrl,omitempty"`
	DisconnectedURL     *string `json:"disconnectedCallbackUrl,omitempty"`
	ConnectedURL        *string `json:"connectedCallbackUrl,omitempty"`
	MessageStatusURL    *string `json:"messageStatusCallbackUrl,omitempty"`
	ChatPresenceURL     *string `json:"presenceChatCallbackUrl,omitempty"`
	HistorySyncURL      *string `json:"historySyncCallbackUrl,omitempty"`
	NotifySentByMe      bool    `json:"notifySentByMe"`
}

// Status represents the instance connection status response.
// fields: connected, error, smartphoneConnected
// Additional fields are hidden from JSON response but preserved for internal use.
type Status struct {
	Connected           bool       `json:"connected"`           // FUNNELCHAT: Indica se seu número está conectado ao FUNNELCHAT
	ConnectionStatus    string     `json:"-"`                   // Hidden: internal connection state tracking
	StoreJID            *string    `json:"-"`                   // Hidden: WhatsApp JID for internal use
	LastConnected       *time.Time `json:"-"`                   // Hidden: last connection timestamp for internal tracking
	InstanceID          uuid.UUID  `json:"-"`                   // Hidden: internal instance identifier
	AutoReconnect       bool       `json:"-"`                   // Hidden: internal reconnection flag
	WorkerAssigned      string     `json:"-"`                   // Hidden: internal worker assignment tracking
	SubscriptionActive  bool       `json:"-"`                   // Hidden: internal subscription state
	Error               string     `json:"error"`               // FUNNELCHAT: Informa detalhes caso algum dos atributos não esteja true
	SmartphoneConnected bool       `json:"smartphoneConnected"` // FUNNELCHAT: Indica se o celular está conectado à internet
}

// DeviceMetadata espelha o objeto "device" retornado pela FUNNELCHAT.
type DeviceMetadata struct {
	SessionName        string `json:"sessionName,omitempty"`
	DeviceModel        string `json:"device_model,omitempty"`
	WAVersion          string `json:"wa_version,omitempty"`
	MCC                string `json:"mcc,omitempty"`
	MNC                string `json:"mnc,omitempty"`
	OSVersion          string `json:"os_version,omitempty"`
	DeviceManufacturer string `json:"device_manufacturer,omitempty"`
	OSBuildNumber      string `json:"osbuildnumber,omitempty"`
	Platform           string `json:"platform,omitempty"`
}

// DeviceResponse replica o payload do GET /device da FUNNELCHAT.
type DeviceResponse struct {
	Phone          string         `json:"phone"`
	ImgURL         *string        `json:"imgUrl"`
	Name           string         `json:"name"`
	Device         DeviceMetadata `json:"device"`
	OriginalDevice string         `json:"originalDevice,omitempty"`
	SessionID      int            `json:"sessionId"`
	IsBusiness     bool           `json:"isBusiness"`
}

type ListFilter struct {
	Query      string
	Middleware string
	Page       int
	PageSize   int
}

type ListResult struct {
	Data     []Instance `json:"data"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
	Total    int64      `json:"total"`
}

type CreateParams struct {
	Name string `json:"name"`
}

// ProxyConfig holds per-instance proxy configuration.
type ProxyConfig struct {
	ProxyURL        *string    `json:"proxyUrl,omitempty"`
	Enabled         bool       `json:"proxyEnabled"`
	NoWebsocket     bool       `json:"noWebsocket"`
	OnlyLogin       bool       `json:"onlyLogin"`
	NoMedia         bool       `json:"noMedia"`
	HealthStatus    string     `json:"healthStatus"`
	LastHealthCheck *time.Time `json:"lastHealthCheck,omitempty"`
	HealthFailures  int        `json:"healthFailures"`
}

// ProxyTestResult holds the result of a proxy connectivity test.
type ProxyTestResult struct {
	Reachable bool   `json:"reachable"`
	LatencyMs int    `json:"latencyMs"`
	Error     string `json:"error,omitempty"`
}

// ProxyHealthLog represents a single health check entry.
type ProxyHealthLog struct {
	ID           int64     `json:"id"`
	InstanceID   uuid.UUID `json:"instanceId"`
	ProxyURL     string    `json:"proxyUrl"`
	Status       string    `json:"status"`
	LatencyMs    *int      `json:"latencyMs,omitempty"`
	ErrorMessage *string   `json:"errorMessage,omitempty"`
	CheckedAt    time.Time `json:"checkedAt"`
}

// ProxyInstance is a lightweight struct for health-check enumeration.
type ProxyInstance struct {
	InstanceID     uuid.UUID
	ProxyURL       string
	HealthStatus   string
	HealthFailures int
}
