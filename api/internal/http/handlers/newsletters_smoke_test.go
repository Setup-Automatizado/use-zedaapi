package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.mau.fi/whatsmeow/api/internal/instances"
	"go.mau.fi/whatsmeow/api/internal/newsletters"
)

// Smoke tests validate all 18 newsletter endpoints are callable and return Z-API compatible responses
// These tests use mocks to verify handler -> service integration without requiring WhatsApp connection

// mockNewslettersService implements newsletters.Service for smoke testing
type mockNewslettersService struct {
	listFunc                  func(ctx context.Context, instanceID uuid.UUID, params newsletters.ListParams) (newsletters.ListResult, error)
	createFunc                func(ctx context.Context, instanceID uuid.UUID, params newsletters.CreateParams) (newsletters.CreateResult, error)
	updatePictureFunc         func(ctx context.Context, instanceID uuid.UUID, params newsletters.UpdatePictureParams) (newsletters.OperationResult, error)
	updateNameFunc            func(ctx context.Context, instanceID uuid.UUID, params newsletters.UpdateNameParams) (newsletters.OperationResult, error)
	updateDescriptionFunc     func(ctx context.Context, instanceID uuid.UUID, params newsletters.UpdateDescriptionParams) (newsletters.OperationResult, error)
	followFunc                func(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error)
	unfollowFunc              func(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error)
	muteFunc                  func(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error)
	unmuteFunc                func(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error)
	deleteFunc                func(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error)
	getMetadataFunc           func(ctx context.Context, instanceID uuid.UUID, id string) (newsletters.MetadataResult, error)
	searchFunc                func(ctx context.Context, instanceID uuid.UUID, params newsletters.SearchParams) (newsletters.SearchResult, error)
	updateSettingsFunc        func(ctx context.Context, instanceID uuid.UUID, params newsletters.SettingsParams) (newsletters.OperationResult, error)
	sendAdminInviteFunc       func(ctx context.Context, instanceID uuid.UUID, params newsletters.AdminActionParams) (newsletters.OperationResult, error)
	acceptAdminInviteFunc     func(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error)
	removeAdminFunc           func(ctx context.Context, instanceID uuid.UUID, params newsletters.AdminActionParams) (newsletters.OperationResult, error)
	revokeAdminInviteFunc     func(ctx context.Context, instanceID uuid.UUID, params newsletters.AdminActionParams) (newsletters.OperationResult, error)
	transferOwnershipFunc     func(ctx context.Context, instanceID uuid.UUID, params newsletters.TransferOwnershipParams) (newsletters.OperationResult, error)
}

func (m *mockNewslettersService) List(ctx context.Context, instanceID uuid.UUID, params newsletters.ListParams) (newsletters.ListResult, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, instanceID, params)
	}
	return newsletters.ListResult{}, nil
}

func (m *mockNewslettersService) Create(ctx context.Context, instanceID uuid.UUID, params newsletters.CreateParams) (newsletters.CreateResult, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, instanceID, params)
	}
	return newsletters.CreateResult{}, nil
}

func (m *mockNewslettersService) UpdatePicture(ctx context.Context, instanceID uuid.UUID, params newsletters.UpdatePictureParams) (newsletters.OperationResult, error) {
	if m.updatePictureFunc != nil {
		return m.updatePictureFunc(ctx, instanceID, params)
	}
	return newsletters.OperationResult{}, nil
}

func (m *mockNewslettersService) UpdateName(ctx context.Context, instanceID uuid.UUID, params newsletters.UpdateNameParams) (newsletters.OperationResult, error) {
	if m.updateNameFunc != nil {
		return m.updateNameFunc(ctx, instanceID, params)
	}
	return newsletters.OperationResult{}, nil
}

func (m *mockNewslettersService) UpdateDescription(ctx context.Context, instanceID uuid.UUID, params newsletters.UpdateDescriptionParams) (newsletters.OperationResult, error) {
	if m.updateDescriptionFunc != nil {
		return m.updateDescriptionFunc(ctx, instanceID, params)
	}
	return newsletters.OperationResult{}, nil
}

func (m *mockNewslettersService) Follow(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error) {
	if m.followFunc != nil {
		return m.followFunc(ctx, instanceID, params)
	}
	return newsletters.OperationResult{}, nil
}

func (m *mockNewslettersService) Unfollow(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error) {
	if m.unfollowFunc != nil {
		return m.unfollowFunc(ctx, instanceID, params)
	}
	return newsletters.OperationResult{}, nil
}

func (m *mockNewslettersService) Mute(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error) {
	if m.muteFunc != nil {
		return m.muteFunc(ctx, instanceID, params)
	}
	return newsletters.OperationResult{}, nil
}

func (m *mockNewslettersService) Unmute(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error) {
	if m.unmuteFunc != nil {
		return m.unmuteFunc(ctx, instanceID, params)
	}
	return newsletters.OperationResult{}, nil
}

func (m *mockNewslettersService) Delete(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error) {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, instanceID, params)
	}
	return newsletters.OperationResult{}, nil
}

func (m *mockNewslettersService) Metadata(ctx context.Context, instanceID uuid.UUID, id string) (newsletters.MetadataResult, error) {
	if m.getMetadataFunc != nil {
		return m.getMetadataFunc(ctx, instanceID, id)
	}
	return newsletters.MetadataResult{}, nil
}

func (m *mockNewslettersService) Search(ctx context.Context, instanceID uuid.UUID, params newsletters.SearchParams) (newsletters.SearchResult, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, instanceID, params)
	}
	return newsletters.SearchResult{}, nil
}

func (m *mockNewslettersService) UpdateSettings(ctx context.Context, instanceID uuid.UUID, params newsletters.SettingsParams) (newsletters.OperationResult, error) {
	if m.updateSettingsFunc != nil {
		return m.updateSettingsFunc(ctx, instanceID, params)
	}
	return newsletters.OperationResult{}, nil
}

func (m *mockNewslettersService) SendAdminInvite(ctx context.Context, instanceID uuid.UUID, params newsletters.AdminActionParams) (newsletters.OperationResult, error) {
	if m.sendAdminInviteFunc != nil {
		return m.sendAdminInviteFunc(ctx, instanceID, params)
	}
	return newsletters.OperationResult{}, nil
}

func (m *mockNewslettersService) AcceptAdminInvite(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error) {
	if m.acceptAdminInviteFunc != nil {
		return m.acceptAdminInviteFunc(ctx, instanceID, params)
	}
	return newsletters.OperationResult{}, nil
}

func (m *mockNewslettersService) RemoveAdmin(ctx context.Context, instanceID uuid.UUID, params newsletters.AdminActionParams) (newsletters.OperationResult, error) {
	if m.removeAdminFunc != nil {
		return m.removeAdminFunc(ctx, instanceID, params)
	}
	return newsletters.OperationResult{}, nil
}

func (m *mockNewslettersService) RevokeAdminInvite(ctx context.Context, instanceID uuid.UUID, params newsletters.AdminActionParams) (newsletters.OperationResult, error) {
	if m.revokeAdminInviteFunc != nil {
		return m.revokeAdminInviteFunc(ctx, instanceID, params)
	}
	return newsletters.OperationResult{}, nil
}

func (m *mockNewslettersService) TransferOwnership(ctx context.Context, instanceID uuid.UUID, params newsletters.TransferOwnershipParams) (newsletters.OperationResult, error) {
	if m.transferOwnershipFunc != nil {
		return m.transferOwnershipFunc(ctx, instanceID, params)
	}
	return newsletters.OperationResult{}, nil
}

// mockInstanceService is a minimal mock for instance authentication
type mockInstanceService struct{}

func (m *mockInstanceService) GetStatus(ctx context.Context, instanceID uuid.UUID, clientToken, instanceToken string) (*instances.Status, error) {
	// Accept any instance ID and tokens for smoke tests
	storeJID := "5511999999999@s.whatsapp.net"
	return &instances.Status{
		Connected: true,
		StoreJID:  &storeJID,
	}, nil
}

// Using a valid UUID format for instanceId parameter
const testURLPrefix = "/instances/123e4567-e89b-12d3-a456-426614174000/token/test-token"

// setupTestRouter creates chi router with newsletter routes for testing
func setupTestRouter(service NewslettersService) *chi.Mux {
	r := chi.NewRouter()

	// Add route context to simulate instance-based routing
	r.Route("/instances/{instanceId}/token/{token}", func(r chi.Router) {
		handler := &NewslettersHandler{
			instanceService: &mockInstanceService{},
			service:         service,
		}
		handler.RegisterRoutes(r)
	})

	return r
}

// TestSmoke01_CreateNewsletter validates POST /create-newsletter
func TestSmoke01_CreateNewsletter(t *testing.T) {
	mockSvc := &mockNewslettersService{
		createFunc: func(ctx context.Context, instanceID uuid.UUID, params newsletters.CreateParams) (newsletters.CreateResult, error) {
			assert.Equal(t, "Test Newsletter", params.Name)
			return newsletters.CreateResult{
				ID: "123456789@newsletter",
			}, nil
		},
	}

	router := setupTestRouter(mockSvc)

	body := []byte(`{"name":"Test Newsletter","description":"Test Description"}`)
	req := httptest.NewRequest(http.MethodPost, testURLPrefix+"/create-newsletter", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Token", "test-client-token")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Logf("Response body: %s", rr.Body.String())
	}
	require.Equal(t, http.StatusCreated, rr.Code, "Expected 201 Created")

	var result newsletters.CreateResult
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	require.NoError(t, err, "Response should be valid JSON")
	assert.Equal(t, "123456789@newsletter", result.ID, "Should return newsletter ID")
}

// TestSmoke02_ListNewsletters validates GET /newsletter
func TestSmoke02_ListNewsletters(t *testing.T) {
	mockSvc := &mockNewslettersService{
		listFunc: func(ctx context.Context, instanceID uuid.UUID, params newsletters.ListParams) (newsletters.ListResult, error) {
			return newsletters.ListResult{
				Items: []newsletters.Summary{
					{
						ID:   "123456789@newsletter",
						Name: "Test Newsletter",
					},
				},
				Total: 1,
			}, nil
		},
	}

	router := setupTestRouter(mockSvc)

	req := httptest.NewRequest(http.MethodGet, testURLPrefix+"/newsletter", nil)
	req.Header.Set("Client-Token", "test-client-token")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "Expected 200 OK")

	var result []newsletters.Summary
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	require.NoError(t, err, "Response should be valid JSON array")
	assert.Len(t, result, 1, "Should return 1 newsletter")
	assert.Equal(t, "123456789@newsletter", result[0].ID, "Newsletter ID should match")
	assert.Equal(t, "Test Newsletter", result[0].Name, "Newsletter name should match")
}

// TestSmoke03_UpdateNewsletterPicture validates POST /update-newsletter-picture
func TestSmoke03_UpdateNewsletterPicture(t *testing.T) {
	mockSvc := &mockNewslettersService{
		updatePictureFunc: func(ctx context.Context, instanceID uuid.UUID, params newsletters.UpdatePictureParams) (newsletters.OperationResult, error) {
			return newsletters.OperationResult{Value: true}, nil
		},
	}

	router := setupTestRouter(mockSvc)

	body := []byte(`{"id":"123456789@newsletter","pictureUrl":"base64data"}`)
	req := httptest.NewRequest(http.MethodPost, testURLPrefix+"/update-newsletter-picture", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Token", "test-client-token")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code, "Expected 201 Created")

	var result newsletters.OperationResult
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	require.NoError(t, err, "Response should be valid JSON")
	assert.True(t, result.Value, "Value should be true")
}

// TestSmoke04_UpdateNewsletterName validates POST /update-newsletter-name
func TestSmoke04_UpdateNewsletterName(t *testing.T) {
	mockSvc := &mockNewslettersService{
		updateNameFunc: func(ctx context.Context, instanceID uuid.UUID, params newsletters.UpdateNameParams) (newsletters.OperationResult, error) {
			assert.Equal(t, "New Name", params.Name)
			return newsletters.OperationResult{Value: true}, nil
		},
	}

	router := setupTestRouter(mockSvc)

	body := []byte(`{"id":"123456789@newsletter","name":"New Name"}`)
	req := httptest.NewRequest(http.MethodPost, testURLPrefix+"/update-newsletter-name", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Token", "test-client-token")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code, "Expected 201 Created")

	var result newsletters.OperationResult
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	require.NoError(t, err, "Response should be valid JSON")
	assert.True(t, result.Value, "Value should be true")
}

// TestSmoke05_UpdateNewsletterDescription validates POST /update-newsletter-description
func TestSmoke05_UpdateNewsletterDescription(t *testing.T) {
	mockSvc := &mockNewslettersService{
		updateDescriptionFunc: func(ctx context.Context, instanceID uuid.UUID, params newsletters.UpdateDescriptionParams) (newsletters.OperationResult, error) {
			assert.Equal(t, "New Description", params.Description)
			return newsletters.OperationResult{Value: true}, nil
		},
	}

	router := setupTestRouter(mockSvc)

	body := []byte(`{"id":"123456789@newsletter","description":"New Description"}`)
	req := httptest.NewRequest(http.MethodPost, testURLPrefix+"/update-newsletter-description", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Token", "test-client-token")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code, "Expected 201 Created")

	var result newsletters.OperationResult
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	require.NoError(t, err, "Response should be valid JSON")
	assert.True(t, result.Value, "Value should be true")
}

// TestSmoke06_FollowNewsletter validates PUT /follow-newsletter
func TestSmoke06_FollowNewsletter(t *testing.T) {
	mockSvc := &mockNewslettersService{
		followFunc: func(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error) {
			return newsletters.OperationResult{Value: true}, nil
		},
	}

	router := setupTestRouter(mockSvc)

	body := []byte(`{"id":"123456789@newsletter"}`)
	req := httptest.NewRequest(http.MethodPut, testURLPrefix+"/follow-newsletter", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Token", "test-client-token")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "Expected 200 OK")

	var result newsletters.OperationResult
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	require.NoError(t, err, "Response should be valid JSON")
	assert.True(t, result.Value, "Value should be true")
}

// TestSmoke07_UnfollowNewsletter validates PUT /unfollow-newsletter
func TestSmoke07_UnfollowNewsletter(t *testing.T) {
	mockSvc := &mockNewslettersService{
		unfollowFunc: func(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error) {
			return newsletters.OperationResult{Value: true}, nil
		},
	}

	router := setupTestRouter(mockSvc)

	body := []byte(`{"id":"123456789@newsletter"}`)
	req := httptest.NewRequest(http.MethodPut, testURLPrefix+"/unfollow-newsletter", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Token", "test-client-token")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "Expected 200 OK")

	var result newsletters.OperationResult
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	require.NoError(t, err, "Response should be valid JSON")
	assert.True(t, result.Value, "Value should be true")
}

// TestSmoke08_MuteNewsletter validates PUT /mute-newsletter
func TestSmoke08_MuteNewsletter(t *testing.T) {
	mockSvc := &mockNewslettersService{
		muteFunc: func(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error) {
			return newsletters.OperationResult{Value: true}, nil
		},
	}

	router := setupTestRouter(mockSvc)

	body := []byte(`{"id":"123456789@newsletter"}`)
	req := httptest.NewRequest(http.MethodPut, testURLPrefix+"/mute-newsletter", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Token", "test-client-token")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "Expected 200 OK")

	var result newsletters.OperationResult
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	require.NoError(t, err, "Response should be valid JSON")
	assert.True(t, result.Value, "Value should be true")
}

// TestSmoke09_UnmuteNewsletter validates PUT /unmute-newsletter
func TestSmoke09_UnmuteNewsletter(t *testing.T) {
	mockSvc := &mockNewslettersService{
		unmuteFunc: func(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error) {
			return newsletters.OperationResult{Value: true}, nil
		},
	}

	router := setupTestRouter(mockSvc)

	body := []byte(`{"id":"123456789@newsletter"}`)
	req := httptest.NewRequest(http.MethodPut, testURLPrefix+"/unmute-newsletter", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Token", "test-client-token")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "Expected 200 OK")

	var result newsletters.OperationResult
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	require.NoError(t, err, "Response should be valid JSON")
	assert.True(t, result.Value, "Value should be true")
}

// TestSmoke10_DeleteNewsletter validates DELETE /delete-newsletter
func TestSmoke10_DeleteNewsletter(t *testing.T) {
	mockSvc := &mockNewslettersService{
		deleteFunc: func(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error) {
			return newsletters.OperationResult{Value: true}, nil
		},
	}

	router := setupTestRouter(mockSvc)

	body := []byte(`{"id":"123456789@newsletter"}`)
	req := httptest.NewRequest(http.MethodDelete, testURLPrefix+"/delete-newsletter", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Token", "test-client-token")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "Expected 200 OK")

	var result newsletters.OperationResult
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	require.NoError(t, err, "Response should be valid JSON")
	assert.True(t, result.Value, "Value should be true")
}

// TestSmoke11_GetNewsletterMetadata validates GET /newsletter/metadata/{newsletterId}
func TestSmoke11_GetNewsletterMetadata(t *testing.T) {
	mockSvc := &mockNewslettersService{
		getMetadataFunc: func(ctx context.Context, instanceID uuid.UUID, id string) (newsletters.MetadataResult, error) {
			assert.Equal(t, "123456789@newsletter", id)
			return newsletters.MetadataResult{
				ID:   "123456789@newsletter",
				Name: "Test Newsletter",
			}, nil
		},
	}

	router := setupTestRouter(mockSvc)

	req := httptest.NewRequest(http.MethodGet, testURLPrefix+"/newsletter/metadata/123456789@newsletter", nil)
	req.Header.Set("Client-Token", "test-client-token")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "Expected 200 OK")

	var result newsletters.Summary
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	require.NoError(t, err, "Response should be valid JSON")
	assert.Equal(t, "123456789@newsletter", result.ID, "Should return newsletter ID")
}

// TestSmoke12_SearchNewsletters validates POST /search-newsletter
func TestSmoke12_SearchNewsletters(t *testing.T) {
	mockSvc := &mockNewslettersService{
		searchFunc: func(ctx context.Context, instanceID uuid.UUID, params newsletters.SearchParams) (newsletters.SearchResult, error) {
			require.NotNil(t, params.SearchText)
			assert.Equal(t, "tech", *params.SearchText)
			return newsletters.SearchResult{
				Data: []newsletters.SearchItem{
					{
						ID:   "123456789@newsletter",
						Name: "Tech News",
					},
				},
			}, nil
		},
	}

	router := setupTestRouter(mockSvc)

	body := []byte(`{"searchText":"tech"}`)
	req := httptest.NewRequest(http.MethodPost, testURLPrefix+"/search-newsletter", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Token", "test-client-token")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "Expected 200 OK")

	var result newsletters.SearchResult
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	require.NoError(t, err, "Response should be valid JSON")
	assert.Len(t, result.Data, 1, "Should return 1 result")
}

// TestSmoke13_UpdateNewsletterSettings validates POST /newsletter/settings/{newsletterId}
func TestSmoke13_UpdateNewsletterSettings(t *testing.T) {
	mockSvc := &mockNewslettersService{
		updateSettingsFunc: func(ctx context.Context, instanceID uuid.UUID, params newsletters.SettingsParams) (newsletters.OperationResult, error) {
			return newsletters.OperationResult{Value: true}, nil
		},
	}

	router := setupTestRouter(mockSvc)

	body := []byte(`{"reactionCodes":"👍,❤️"}`)
	req := httptest.NewRequest(http.MethodPost, testURLPrefix+"/newsletter/settings/123456789@newsletter", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Token", "test-client-token")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code, "Expected 201 Created")

	var result newsletters.OperationResult
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	require.NoError(t, err, "Response should be valid JSON")
	assert.True(t, result.Value, "Value should be true")
}

// TestSmoke14_SendNewsletterAdminInvite validates POST /send-newsletter-admin-invite (BONUS endpoint)
func TestSmoke14_SendNewsletterAdminInvite(t *testing.T) {
	mockSvc := &mockNewslettersService{
		sendAdminInviteFunc: func(ctx context.Context, instanceID uuid.UUID, params newsletters.AdminActionParams) (newsletters.OperationResult, error) {
			assert.Equal(t, "123456789@newsletter", params.ID)
			assert.Equal(t, "5511999999999", params.Phone)
			return newsletters.OperationResult{Value: true}, nil
		},
	}

	router := setupTestRouter(mockSvc)

	body := []byte(`{"id":"123456789@newsletter","phone":"5511999999999"}`)
	req := httptest.NewRequest(http.MethodPost, testURLPrefix+"/send-newsletter-admin-invite", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Token", "test-client-token")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code, "Expected 201 Created")

	var result newsletters.OperationResult
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	require.NoError(t, err, "Response should be valid JSON")
	assert.True(t, result.Value, "Value should be true")
}

// TestSmoke15_AcceptNewsletterAdminInvite validates POST /newsletter/accept-admin-invite/{newsletterId}
func TestSmoke15_AcceptNewsletterAdminInvite(t *testing.T) {
	mockSvc := &mockNewslettersService{
		acceptAdminInviteFunc: func(ctx context.Context, instanceID uuid.UUID, params newsletters.IDParams) (newsletters.OperationResult, error) {
			return newsletters.OperationResult{Value: true}, nil
		},
	}

	router := setupTestRouter(mockSvc)

	body := []byte(`{"inviteCode":"abc123"}`)
	req := httptest.NewRequest(http.MethodPost, testURLPrefix+"/newsletter/accept-admin-invite/123456789@newsletter", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Token", "test-client-token")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code, "Expected 201 Created")

	var result newsletters.OperationResult
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	require.NoError(t, err, "Response should be valid JSON")
	assert.True(t, result.Value, "Value should be true")
}

// TestSmoke16_RemoveNewsletterAdmin validates POST /newsletter/remove-admin/{newsletterId}
func TestSmoke16_RemoveNewsletterAdmin(t *testing.T) {
	mockSvc := &mockNewslettersService{
		removeAdminFunc: func(ctx context.Context, instanceID uuid.UUID, params newsletters.AdminActionParams) (newsletters.OperationResult, error) {
			return newsletters.OperationResult{Value: true}, nil
		},
	}

	router := setupTestRouter(mockSvc)

	body := []byte(`{"phone":"5511999999999"}`)
	req := httptest.NewRequest(http.MethodPost, testURLPrefix+"/newsletter/remove-admin/123456789@newsletter", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Token", "test-client-token")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code, "Expected 201 Created")

	var result newsletters.OperationResult
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	require.NoError(t, err, "Response should be valid JSON")
	assert.True(t, result.Value, "Value should be true")
}

// TestSmoke17_RevokeNewsletterAdminInvite validates POST /newsletter/revoke-admin-invite/{newsletterId}
func TestSmoke17_RevokeNewsletterAdminInvite(t *testing.T) {
	mockSvc := &mockNewslettersService{
		revokeAdminInviteFunc: func(ctx context.Context, instanceID uuid.UUID, params newsletters.AdminActionParams) (newsletters.OperationResult, error) {
			return newsletters.OperationResult{Value: true}, nil
		},
	}

	router := setupTestRouter(mockSvc)

	body := []byte(`{"phone":"5511999999999"}`)
	req := httptest.NewRequest(http.MethodPost, testURLPrefix+"/newsletter/revoke-admin-invite/123456789@newsletter", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Token", "test-client-token")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code, "Expected 201 Created")

	var result newsletters.OperationResult
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	require.NoError(t, err, "Response should be valid JSON")
	assert.True(t, result.Value, "Value should be true")
}

// TestSmoke18_TransferNewsletterOwnership validates POST /newsletter/transfer-ownership/{newsletterId}
func TestSmoke18_TransferNewsletterOwnership(t *testing.T) {
	mockSvc := &mockNewslettersService{
		transferOwnershipFunc: func(ctx context.Context, instanceID uuid.UUID, params newsletters.TransferOwnershipParams) (newsletters.OperationResult, error) {
			assert.Equal(t, "5511999999999", params.Phone)
			return newsletters.OperationResult{Value: true}, nil
		},
	}

	router := setupTestRouter(mockSvc)

	body := []byte(`{"phone":"5511999999999"}`)
	req := httptest.NewRequest(http.MethodPost, testURLPrefix+"/newsletter/transfer-ownership/123456789@newsletter", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Token", "test-client-token")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code, "Expected 201 Created")

	var result newsletters.OperationResult
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	require.NoError(t, err, "Response should be valid JSON")
	assert.True(t, result.Value, "Value should be true")
}

// TestSmokeAll_EndpointCount verifies all 18 endpoints are registered
func TestSmokeAll_EndpointCount(t *testing.T) {
	mockSvc := &mockNewslettersService{}
	router := setupTestRouter(mockSvc)

	// Count registered routes by attempting each HTTP method+path combination
	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/create-newsletter"},
		{http.MethodGet, "/newsletter"},
		{http.MethodPost, "/update-newsletter-picture"},
		{http.MethodPost, "/update-newsletter-name"},
		{http.MethodPost, "/update-newsletter-description"},
		{http.MethodPut, "/follow-newsletter"},
		{http.MethodPut, "/unfollow-newsletter"},
		{http.MethodPut, "/mute-newsletter"},
		{http.MethodPut, "/unmute-newsletter"},
		{http.MethodDelete, "/delete-newsletter"},
		{http.MethodGet, "/newsletter/metadata/123456789@newsletter"},
		{http.MethodPost, "/search-newsletter"},
		{http.MethodPost, "/newsletter/settings/123456789@newsletter"},
		{http.MethodPost, "/send-newsletter-admin-invite"},
		{http.MethodPost, "/newsletter/accept-admin-invite/123456789@newsletter"},
		{http.MethodPost, "/newsletter/remove-admin/123456789@newsletter"},
		{http.MethodPost, "/newsletter/revoke-admin-invite/123456789@newsletter"},
		{http.MethodPost, "/newsletter/transfer-ownership/123456789@newsletter"},
	}

	registeredCount := 0
	for _, ep := range endpoints {
		req := httptest.NewRequest(ep.method, testURLPrefix+ep.path, nil)
		req.Header.Set("Client-Token", "test-client-token")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// 405 Method Not Allowed means route exists but method mismatch
		// 404 Not Found means route doesn't exist
		// Anything else (including 400, 500) means route is registered
		if rr.Code != http.StatusNotFound {
			registeredCount++
		}
	}

	assert.Equal(t, 18, registeredCount, "All 18 newsletter endpoints must be registered (17 Z-API + 1 bonus)")
}
