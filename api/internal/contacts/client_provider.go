package contacts

import (
	"context"

	"go.mau.fi/whatsmeow/types"
)

// ClientProvider abstracts access to a whatsmeow Client instance for a specific WhatsApp instance.
// This interface allows the service layer to retrieve contacts without directly depending on
// the whatsmeow Client type.
//
// Deprecated: Use the Client and ClientProvider interfaces defined in service.go instead.
// This file is kept for backwards compatibility but will be removed in a future version.
type ContactsClientProvider interface {
	// GetAllContacts returns all contacts from the WhatsApp instance's contact store.
	// Returns a map where keys are JIDs and values are ContactInfo.
	GetAllContacts(ctx context.Context) (map[types.JID]types.ContactInfo, error)
}
