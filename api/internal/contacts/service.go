package contacts

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow/types"

	"go.mau.fi/whatsmeow/api/internal/logging"
)

// Client exposes the WhatsApp operations required by the contacts service.
type Client interface {
	GetAllContacts(ctx context.Context) (map[types.JID]types.ContactInfo, error)
}

// ClientProvider resolves a WhatsApp client for a given instance.
type ClientProvider interface {
	Get(ctx context.Context, instanceID uuid.UUID) (Client, error)
}

// Service provides business logic for contacts operations.
type Service struct {
	provider ClientProvider
	log      *slog.Logger
}

// NewService creates a new contacts service.
func NewService(
	provider ClientProvider,
	log *slog.Logger,
) *Service {
	return &Service{
		provider: provider,
		log:      log,
	}
}

// List retrieves contacts for a WhatsApp instance with pagination.
// It fetches all contacts from the store, sorts them, and returns the requested page.
func (s *Service) List(ctx context.Context, instanceID uuid.UUID, params ListParams) (ListResult, error) {
	logger := logging.ContextLogger(ctx, s.log)

	logger.InfoContext(ctx, "listing contacts",
		slog.String("instance_id", instanceID.String()),
		slog.Int("page", params.Page),
		slog.Int("page_size", params.PageSize))

	// Get whatsmeow client for this instance
	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get client",
			slog.String("error", err.Error()))
		return ListResult{}, fmt.Errorf("get client: %w", err)
	}

	// Fetch all contacts from store
	contactsMap, err := client.GetAllContacts(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get contacts from store",
			slog.String("error", err.Error()))
		return ListResult{}, fmt.Errorf("get all contacts: %w", err)
	}

	logger.DebugContext(ctx, "retrieved contacts from store",
		slog.Int("total_contacts", len(contactsMap)))

	// Convert map to slice for sorting and pagination
	contacts := make([]Contact, 0, len(contactsMap))
	for jid, info := range contactsMap {
		contact := s.toContact(jid, info)
		contacts = append(contacts, contact)
	}

	// Sort contacts by phone number for consistent ordering
	sort.Slice(contacts, func(i, j int) bool {
		return contacts[i].Phone < contacts[j].Phone
	})

	total := len(contacts)

	// Apply pagination
	start := (params.Page - 1) * params.PageSize
	if start >= total {
		return ListResult{Items: []Contact{}, Total: total}, nil
	}

	end := start + params.PageSize
	if end > total {
		end = total
	}

	pageItems := contacts[start:end]

	logger.InfoContext(ctx, "contacts listed successfully",
		slog.Int("total_contacts", total),
		slog.Int("page_items", len(pageItems)),
		slog.Int("page", params.Page),
		slog.Int("page_size", params.PageSize))

	return ListResult{Items: pageItems, Total: total}, nil
}

// toContact converts whatsmeow types to Z-API compatible Contact format.
// It handles optional fields by using pointers and setting them to nil when data is missing.
func (s *Service) toContact(jid types.JID, info types.ContactInfo) Contact {
	// Extract phone number from JID
	phone := jid.User

	// Z-API expects fields to be nil if not available (not empty strings)
	var name *string
	if info.FullName != "" {
		name = &info.FullName
	}

	var short *string
	if info.FirstName != "" {
		short = &info.FirstName
	}

	var notify *string
	if info.PushName != "" {
		notify = &info.PushName
	}

	var vname *string
	if info.FullName != "" {
		vname = &info.FullName
	}

	// Handle business contacts - prefer business name
	if info.BusinessName != "" {
		businessName := info.BusinessName
		name = &businessName
		vname = &businessName
	}

	// Handle redacted phones (LID members in groups)
	// When a phone is redacted, WhatsApp provides a partial number
	if info.RedactedPhone != "" {
		redacted := info.RedactedPhone
		// Remove the + prefix if present
		phone = strings.TrimPrefix(redacted, "+")
	}

	return Contact{
		Phone:  phone,
		Name:   name,
		Short:  short,
		Notify: notify,
		Vname:  vname,
	}
}
