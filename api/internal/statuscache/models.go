package statuscache

import (
	"time"
)

// StatusType represents the type of message status
type StatusType string

const (
	StatusSent      StatusType = "sent"
	StatusDelivered StatusType = "delivered"
	StatusRead      StatusType = "read"
	StatusPlayed    StatusType = "played"
)

// ValidStatusTypes contains all valid status types
var ValidStatusTypes = map[StatusType]bool{
	StatusSent:      true,
	StatusDelivered: true,
	StatusRead:      true,
	StatusPlayed:    true,
}

// IsValidStatusType checks if a status type is valid
func IsValidStatusType(s string) bool {
	return ValidStatusTypes[StatusType(s)]
}

// ParticipantStatus represents the status of a single participant
type ParticipantStatus struct {
	Phone     string `json:"phone"`
	Status    string `json:"status"`    // sent, delivered, read, played
	Timestamp int64  `json:"timestamp"` // Unix timestamp in milliseconds
	Device    int    `json:"device,omitempty"`
}

// StatusCacheEntry represents a cached message status entry
type StatusCacheEntry struct {
	MessageID    string                        `json:"messageId"`
	InstanceID   string                        `json:"instanceId"`
	GroupID      string                        `json:"groupId,omitempty"`
	Phone        string                        `json:"phone"`
	IsGroup      bool                          `json:"isGroup"`
	CreatedAt    int64                         `json:"createdAt"` // Unix timestamp in milliseconds
	UpdatedAt    int64                         `json:"updatedAt"` // Unix timestamp in milliseconds
	Participants map[string]*ParticipantStatus `json:"participants"`
}

// NewStatusCacheEntry creates a new StatusCacheEntry
func NewStatusCacheEntry(instanceID, messageID, phone, groupID string, isGroup bool) *StatusCacheEntry {
	now := time.Now().UnixMilli()
	return &StatusCacheEntry{
		MessageID:    messageID,
		InstanceID:   instanceID,
		GroupID:      groupID,
		Phone:        phone,
		IsGroup:      isGroup,
		CreatedAt:    now,
		UpdatedAt:    now,
		Participants: make(map[string]*ParticipantStatus),
	}
}

// AddParticipant adds or updates a participant status
func (e *StatusCacheEntry) AddParticipant(phone, status string, timestamp int64, device int) {
	e.Participants[phone] = &ParticipantStatus{
		Phone:     phone,
		Status:    status,
		Timestamp: timestamp,
		Device:    device,
	}
	e.UpdatedAt = time.Now().UnixMilli()
}

// GetStatusCounts returns the count of each status type
func (e *StatusCacheEntry) GetStatusCounts() map[string]int {
	counts := map[string]int{
		string(StatusSent):      0,
		string(StatusDelivered): 0,
		string(StatusRead):      0,
		string(StatusPlayed):    0,
	}
	for _, p := range e.Participants {
		if _, ok := counts[p.Status]; ok {
			counts[p.Status]++
		}
	}
	return counts
}

// AggregatedStatus is the API response for status queries
type AggregatedStatus struct {
	MessageID         string              `json:"messageId"`
	InstanceID        string              `json:"instanceId"`
	GroupID           string              `json:"groupId,omitempty"`
	Phone             string              `json:"phone"`
	IsGroup           bool                `json:"isGroup"`
	TotalParticipants int                 `json:"totalParticipants"`
	StatusCounts      map[string]int      `json:"status"`
	Participants      []ParticipantStatus `json:"participants,omitempty"`
	CreatedAt         int64               `json:"createdAt"`
	UpdatedAt         int64               `json:"updatedAt"`
}

// ToAggregated converts a StatusCacheEntry to AggregatedStatus
func (e *StatusCacheEntry) ToAggregated(includeParticipants bool) *AggregatedStatus {
	agg := &AggregatedStatus{
		MessageID:         e.MessageID,
		InstanceID:        e.InstanceID,
		GroupID:           e.GroupID,
		Phone:             e.Phone,
		IsGroup:           e.IsGroup,
		TotalParticipants: len(e.Participants),
		StatusCounts:      e.GetStatusCounts(),
		CreatedAt:         e.CreatedAt,
		UpdatedAt:         e.UpdatedAt,
	}

	if includeParticipants {
		agg.Participants = make([]ParticipantStatus, 0, len(e.Participants))
		for _, p := range e.Participants {
			agg.Participants = append(agg.Participants, *p)
		}
	}

	return agg
}

// StatusEvent represents an incoming status event to be cached
type StatusEvent struct {
	InstanceID  string   `json:"instanceId"`
	MessageIDs  []string `json:"messageIds"`
	Status      string   `json:"status"` // sent, delivered, read, played
	Phone       string   `json:"phone"`
	Participant string   `json:"participant"`
	IsGroup     bool     `json:"isGroup"`
	GroupID     string   `json:"groupId,omitempty"`
	Timestamp   int64    `json:"timestamp"` // Unix timestamp in milliseconds
	Device      int      `json:"device,omitempty"`
}

// QueryParams contains parameters for querying the status cache
type QueryParams struct {
	Limit               int  `json:"limit"`
	Offset              int  `json:"offset"`
	IncludeParticipants bool `json:"includeParticipants"`
}

// DefaultQueryParams returns default query parameters
func DefaultQueryParams() QueryParams {
	return QueryParams{
		Limit:               100,
		Offset:              0,
		IncludeParticipants: false,
	}
}

// QueryResult contains the result of a status cache query
type QueryResult struct {
	Data []*AggregatedStatus `json:"data"`
	Meta QueryMeta           `json:"meta"`
}

// QueryMeta contains metadata for query results
type QueryMeta struct {
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

// FlushRequest contains parameters for flushing cached statuses
type FlushRequest struct {
	MessageID string `json:"messageId,omitempty"`
	All       bool   `json:"all,omitempty"`
}

// FlushResult contains the result of a flush operation
type FlushResult struct {
	Flushed           int64 `json:"flushed"`
	WebhooksTriggered int64 `json:"webhooksTriggered"`
}

// PendingWebhook represents a webhook that was suppressed and can be flushed
type PendingWebhook struct {
	MessageID   string `json:"messageId"`
	InstanceID  string `json:"instanceId"`
	Phone       string `json:"phone"`
	Participant string `json:"participant"`
	Status      string `json:"status"`
	Timestamp   int64  `json:"timestamp"`
	Device      int    `json:"device,omitempty"`
	Payload     []byte `json:"payload"` // Original FUNNELCHAT payload for flush
}

// CacheStats contains statistics about the status cache
type CacheStats struct {
	TotalEntries      int64            `json:"totalEntries"`
	EntriesByInstance map[string]int64 `json:"entriesByInstance"`
	PendingWebhooks   int64            `json:"pendingWebhooks"`
	OldestEntry       int64            `json:"oldestEntry,omitempty"` // Unix timestamp
	NewestEntry       int64            `json:"newestEntry,omitempty"` // Unix timestamp
}

// RawStatusPayload matches the FUNNELCHAT webhook payload format exactly
// This allows systems to process cached data identically to webhooks
type RawStatusPayload struct {
	Status         string   `json:"status"`
	Ids            []string `json:"ids"`
	Momment        int64    `json:"momment"` // FUNNELCHAT uses "momment" (typo preserved for compatibility)
	Phone          string   `json:"phone"`
	ChatLid        string   `json:"chatLid,omitempty"`
	ChatPN         string   `json:"chat_pn,omitempty"`
	ParticipantLid string   `json:"participantLid,omitempty"`
	ParticipantPN  string   `json:"participant_pn,omitempty"`
	IsGroup        bool     `json:"isGroup"`
	InstanceId     string   `json:"instanceId"`
	Type           string   `json:"type"` // "MessageStatusCallback"
	FromMe         bool     `json:"fromMe"`
	WaitingMessage bool     `json:"waitingMessage,omitempty"`
	SenderDevice   int      `json:"senderDevice,omitempty"`
}

// RawQueryResult contains the result of a raw status query
type RawQueryResult struct {
	Data []*RawStatusPayload `json:"data"`
	Meta QueryMeta           `json:"meta"`
}

// ToRawPayloads converts a StatusCacheEntry to raw payloads
func (e *StatusCacheEntry) ToRawPayloads() []*RawStatusPayload {
	payloads := make([]*RawStatusPayload, 0, len(e.Participants))

	for _, p := range e.Participants {
		payload := &RawStatusPayload{
			Status:       p.Status,
			Ids:          []string{e.MessageID},
			Momment:      p.Timestamp,
			Phone:        e.Phone,
			IsGroup:      e.IsGroup,
			InstanceId:   e.InstanceID,
			Type:         "MessageStatusCallback",
			FromMe:       true,
			SenderDevice: p.Device,
		}

		// Set participant info for groups
		if e.IsGroup {
			payload.ParticipantPN = p.Phone
		}

		payloads = append(payloads, payload)
	}

	return payloads
}
