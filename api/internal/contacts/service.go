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
	IsOnWhatsApp(ctx context.Context, phones []string) ([]types.IsOnWhatsAppResponse, error)
	GetUserInfo(ctx context.Context, jids []types.JID) (map[types.JID]types.UserInfo, error)
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

// toContact converts whatsmeow types to Contact format.
// It handles optional fields by using pointers and setting them to nil when data is missing.
func (s *Service) toContact(jid types.JID, info types.ContactInfo) Contact {
	// Extract phone number from JID
	phone := jid.User

	// Zé da API expects fields to be nil if not available (not empty strings)
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

// IsOnWhatsApp checks if a phone number is registered on WhatsApp.
// Returns PhoneExistsResponse with exists status, phone number, and LID if available.
func (s *Service) IsOnWhatsApp(ctx context.Context, instanceID uuid.UUID, phone string) (PhoneExistsResponse, error) {
	logger := logging.ContextLogger(ctx, s.log)

	logger.InfoContext(ctx, "checking if phone is on whatsapp",
		slog.String("instance_id", instanceID.String()),
		slog.String("phone", phone))

	// Get whatsmeow client for this instance
	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get client",
			slog.String("error", err.Error()))
		return PhoneExistsResponse{Exists: false}, fmt.Errorf("get client: %w", err)
	}

	// Check if phone is on WhatsApp
	responses, err := client.IsOnWhatsApp(ctx, []string{phone})
	if err != nil {
		logger.ErrorContext(ctx, "failed to check whatsapp",
			slog.String("error", err.Error()))
		return PhoneExistsResponse{Exists: false}, fmt.Errorf("check whatsapp: %w", err)
	}

	if len(responses) == 0 {
		logger.InfoContext(ctx, "no response from whatsapp check")
		return PhoneExistsResponse{Exists: false}, nil
	}

	resp := responses[0]
	exists := resp.IsIn

	// If phone doesn't exist on WhatsApp, return early
	if !exists {
		logger.InfoContext(ctx, "phone not on whatsapp",
			slog.String("phone", phone))
		return PhoneExistsResponse{Exists: false}, nil
	}

	// Phone exists - get the canonical phone number from JID
	phoneNumber := resp.JID.User
	result := PhoneExistsResponse{
		Exists: true,
		Phone:  &phoneNumber,
	}

	// Try to get LID using GetUserInfo
	userInfoMap, err := client.GetUserInfo(ctx, []types.JID{resp.JID})
	if err != nil {
		// Log the error but don't fail - LID is optional
		logger.WarnContext(ctx, "failed to get user info for LID",
			slog.String("error", err.Error()),
			slog.String("phone", phone))
	} else if userInfo, ok := userInfoMap[resp.JID]; ok && !userInfo.LID.IsEmpty() {
		lidStr := userInfo.LID.String()
		result.LID = &lidStr
		logger.InfoContext(ctx, "phone check completed with LID",
			slog.Bool("exists", exists),
			slog.String("phone", phoneNumber),
			slog.String("lid", lidStr))
	} else {
		logger.InfoContext(ctx, "phone check completed without LID",
			slog.Bool("exists", exists),
			slog.String("phone", phoneNumber))
	}

	return result, nil
}

// IsOnWhatsAppBatch checks if multiple phone numbers are registered on WhatsApp.
// Returns a slice of batch responses with validation results for each phone.
// Maximum batch size is 50,000 numbers per request (Zé da API limit).
func (s *Service) IsOnWhatsAppBatch(ctx context.Context, instanceID uuid.UUID, phones []string) ([]PhoneExistsBatchResponse, error) {
	logger := logging.ContextLogger(ctx, s.log)

	logger.InfoContext(ctx, "checking batch phones on whatsapp",
		slog.String("instance_id", instanceID.String()),
		slog.Int("phone_count", len(phones)))

	// Get whatsmeow client for this instance
	client, err := s.provider.Get(ctx, instanceID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get client",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("get client: %w", err)
	}

	// Check if phones are on WhatsApp
	responses, err := client.IsOnWhatsApp(ctx, phones)
	if err != nil {
		logger.ErrorContext(ctx, "failed to check whatsapp batch",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("check whatsapp batch: %w", err)
	}

	// Build a map of query -> response for fast lookup
	responseMap := make(map[string]types.IsOnWhatsAppResponse, len(responses))
	for _, resp := range responses {
		responseMap[resp.Query] = resp
	}

	// Build results maintaining input order
	results := make([]PhoneExistsBatchResponse, len(phones))
	for i, phone := range phones {
		resp, found := responseMap[phone]
		if found {
			// Use the JID user part as the output phone (formatted by WhatsApp)
			outputPhone := resp.JID.User
			if outputPhone == "" {
				outputPhone = phone
			}
			results[i] = PhoneExistsBatchResponse{
				Exists:      resp.IsIn,
				InputPhone:  phone,
				OutputPhone: outputPhone,
			}
		} else {
			// Phone not found in response - mark as not existing
			results[i] = PhoneExistsBatchResponse{
				Exists:      false,
				InputPhone:  phone,
				OutputPhone: phone,
			}
		}
	}

	logger.InfoContext(ctx, "batch phone check completed",
		slog.Int("total_checked", len(phones)),
		slog.Int("results_count", len(results)))

	return results, nil
}
