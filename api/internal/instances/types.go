package instances

import (
	"time"

	"github.com/google/uuid"
)

type Instance struct {
	ID                 uuid.UUID        `json:"instanceId"`
	Name               string           `json:"name"`
	SessionName        string           `json:"sessionName"`
	ClientToken        string           `json:"clientToken"`
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
}

type WebhookSettings struct {
	DeliveryURL         *string `json:"deliveryCallbackUrl,omitempty"`
	ReceivedURL         *string `json:"receivedCallbackUrl,omitempty"`
	ReceivedDeliveryURL *string `json:"receivedAndDeliveryCallbackUrl,omitempty"`
	DisconnectedURL     *string `json:"disconnectedCallbackUrl,omitempty"`
	ConnectedURL        *string `json:"connectedCallbackUrl,omitempty"`
	MessageStatusURL    *string `json:"messageStatusCallbackUrl,omitempty"`
	ChatPresenceURL     *string `json:"presenceChatCallbackUrl,omitempty"`
	NotifySentByMe      bool    `json:"notifySentByMe"`
}

type Status struct {
	Connected          bool      `json:"connected"`
	StoreJID           *string   `json:"storeJid,omitempty"`
	LastConnected      time.Time `json:"lastConnected,omitempty"`
	InstanceID         uuid.UUID `json:"instanceId"`
	AutoReconnect      bool      `json:"autoReconnect"`
	WorkerAssigned     string    `json:"workerAssigned"`
	SubscriptionActive bool      `json:"subscriptionActive"`
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
