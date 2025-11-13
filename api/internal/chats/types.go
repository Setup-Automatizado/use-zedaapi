package chats

import (
	"time"

	watypes "go.mau.fi/whatsmeow/types"
)

// Chat represents a WhatsApp conversation (can be a contact or a group).
// This follows the Z-API principle that "everything is a chat".
type Chat struct {
	Phone            string    `json:"phone"`            // WhatsApp JID (e.g., "5511999999999@s.whatsapp.net" or group JID)
	Name             string    `json:"name"`             // Contact name or group subject
	Unread           int       `json:"unread"`           // Unread message count (not available from whatsmeow store)
	LastMessageTime  time.Time `json:"lastMessageTime"`  // Timestamp of last message (not available from whatsmeow store)
	IsMuted          bool      `json:"isMuted"`          // Whether chat is muted
	MuteEndTime      time.Time `json:"muteEndTime"`      // When mute expires (zero if not muted or muted forever)
	IsMarkedSpam     bool      `json:"isMarkedSpam"`     // Whether chat is marked as spam (not available from whatsmeow store)
	Archived         bool      `json:"archived"`         // Whether chat is archived
	Pinned           bool      `json:"pinned"`           // Whether chat is pinned
	MessagesUnread   int       `json:"messagesUnread"`   // Alternative field for unread count (Z-API compatibility)
	IsGroup          bool      `json:"isGroup"`          // Whether this is a group chat
}

// ListParams defines parameters for listing chats with pagination.
type ListParams struct {
	Page     int `json:"page"`     // Page number (1-indexed)
	PageSize int `json:"pageSize"` // Number of items per page
}

// ListResult contains paginated chat results and metadata.
type ListResult struct {
	Chats      []Chat `json:"chats"`      // Array of chat objects
	TotalCount int    `json:"totalCount"` // Total number of chats available
	Page       int    `json:"page"`       // Current page number
	PageSize   int    `json:"pageSize"`   // Items per page
}

// fromContactInfo creates a Chat from a contact entry.
func fromContactInfo(jid watypes.JID, info watypes.ContactInfo, settings watypes.LocalChatSettings) Chat {
	var muteEndTime time.Time
	isMuted := false
	if !settings.MutedUntil.IsZero() && settings.MutedUntil.After(time.Now()) {
		isMuted = true
		muteEndTime = settings.MutedUntil
	}

	name := info.FullName
	if name == "" {
		name = info.PushName
	}
	if name == "" {
		name = info.BusinessName
	}

	return Chat{
		Phone:            jid.String(),
		Name:             name,
		Unread:           0, // Not available from whatsmeow store
		LastMessageTime:  time.Time{}, // Not available from whatsmeow store
		IsMuted:          isMuted,
		MuteEndTime:      muteEndTime,
		IsMarkedSpam:     false, // Not available from whatsmeow store
		Archived:         settings.Archived,
		Pinned:           settings.Pinned,
		MessagesUnread:   0, // Not available from whatsmeow store
		IsGroup:          false,
	}
}

// fromGroupInfo creates a Chat from a group entry.
func fromGroupInfo(info *watypes.GroupInfo, settings watypes.LocalChatSettings) Chat {
	var muteEndTime time.Time
	isMuted := false
	if !settings.MutedUntil.IsZero() && settings.MutedUntil.After(time.Now()) {
		isMuted = true
		muteEndTime = settings.MutedUntil
	}

	// Use GroupCreated as lastMessageTime fallback since store doesn't expose message history
	lastMessageTime := time.Time{}
	if info.GroupCreated.Unix() > 0 {
		lastMessageTime = info.GroupCreated
	}

	return Chat{
		Phone:            info.JID.String(),
		Name:             info.Name,
		Unread:           0, // Not available from whatsmeow store
		LastMessageTime:  lastMessageTime,
		IsMuted:          isMuted,
		MuteEndTime:      muteEndTime,
		IsMarkedSpam:     false, // Not available from whatsmeow store
		Archived:         settings.Archived,
		Pinned:           settings.Pinned,
		MessagesUnread:   0, // Not available from whatsmeow store
		IsGroup:          true,
	}
}
