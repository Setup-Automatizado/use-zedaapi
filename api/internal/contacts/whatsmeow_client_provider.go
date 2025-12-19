package contacts

import (
	"context"
	"fmt"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

// whatsmeowClient wraps a whatsmeow.Client to implement the Client interface.
type whatsmeowClient struct {
	client *whatsmeow.Client
}

// NewWhatsmeowClientProvider creates a Client that wraps a whatsmeow.Client.
func NewWhatsmeowClientProvider(client *whatsmeow.Client) Client {
	return &whatsmeowClient{client: client}
}

// GetAllContacts retrieves all contacts from the whatsmeow client's contact store.
func (w *whatsmeowClient) GetAllContacts(ctx context.Context) (map[types.JID]types.ContactInfo, error) {
	if w.client == nil {
		return nil, fmt.Errorf("whatsmeow client is nil")
	}

	if w.client.Store == nil {
		return nil, fmt.Errorf("whatsmeow client store is nil")
	}

	if w.client.Store.Contacts == nil {
		return nil, fmt.Errorf("whatsmeow client contacts store is nil")
	}

	contacts, err := w.client.Store.Contacts.GetAllContacts(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all contacts from store: %w", err)
	}

	return contacts, nil
}

// IsOnWhatsApp checks if the given phone numbers are registered on WhatsApp.
func (w *whatsmeowClient) IsOnWhatsApp(ctx context.Context, phones []string) ([]types.IsOnWhatsAppResponse, error) {
	if w.client == nil {
		return nil, fmt.Errorf("whatsmeow client is nil")
	}

	return w.client.IsOnWhatsApp(ctx, phones)
}

// GetUserInfo gets basic user info (avatar, status, verified business name, device list, LID).
func (w *whatsmeowClient) GetUserInfo(ctx context.Context, jids []types.JID) (map[types.JID]types.UserInfo, error) {
	if w.client == nil {
		return nil, fmt.Errorf("whatsmeow client is nil")
	}

	return w.client.GetUserInfo(ctx, jids)
}
