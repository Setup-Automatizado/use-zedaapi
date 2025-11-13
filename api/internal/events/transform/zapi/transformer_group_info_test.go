package zapi

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	internaltypes "go.mau.fi/whatsmeow/api/internal/events/types"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func TestTransformGroupInfoMembershipRequest(t *testing.T) {
	transformer := NewTransformer("5511999999999", false, "", nil)

	groupJID := types.NewJID("120363420116051586", types.GroupServer)
	memberJID := types.NewJID("227294266302623", types.HiddenUserServer)

	info := &events.GroupInfo{
		JID:                       groupJID,
		MembershipRequestsCreated: []types.JID{memberJID},
		MembershipRequestMethod:   "invite_link",
		Timestamp:                 time.Now(),
	}

	event := &internaltypes.InternalEvent{
		InstanceID: uuid.New(),
		EventID:    uuid.New(),
		EventType:  "group_info",
		Metadata: map[string]string{
			"timestamp":                     fmt.Sprintf("%d", time.Now().Unix()),
			"group_id":                      groupJID.String(),
			"membership_request_created_pn": `["5521971532700@s.whatsapp.net"]`,
		},
		RawPayload: info,
		CapturedAt: time.Now(),
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	callback, err := transformer.transformGroupInfoEvent(context.Background(), logger, event)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if callback.Notification != "MEMBERSHIP_APPROVAL_REQUEST" {
		t.Fatalf("unexpected notification %s", callback.Notification)
	}
	if len(callback.NotificationParameters) != 1 {
		t.Fatalf("expected 1 participant parameter, got %d", len(callback.NotificationParameters))
	}
	expectedParam := conversationIdentifierFromJID(memberJID)
	if callback.NotificationParameters[0] != expectedParam {
		t.Fatalf("unexpected participant identifier %s", callback.NotificationParameters[0])
	}
	if callback.RequestMethod != "invite_link" {
		t.Fatalf("expected request method invite_link, got %s", callback.RequestMethod)
	}
	if callback.ParticipantPhone != "5521971532700" {
		t.Fatalf("expected participant phone fallback, got %s", callback.ParticipantPhone)
	}
}

func TestTransformGroupInfoMembershipRequestRevoked(t *testing.T) {
	transformer := NewTransformer("5511999999999", false, "", nil)

	groupJID := types.NewJID("120363420116051586", types.GroupServer)
	memberJID := types.NewJID("227294266302623", types.HiddenUserServer)

	info := &events.GroupInfo{
		JID:                       groupJID,
		MembershipRequestsRevoked: []types.JID{memberJID},
		Timestamp:                 time.Now(),
	}

	event := &internaltypes.InternalEvent{
		InstanceID: uuid.New(),
		EventID:    uuid.New(),
		EventType:  "group_info",
		Metadata: map[string]string{
			"timestamp":                     fmt.Sprintf("%d", time.Now().Unix()),
			"group_id":                      groupJID.String(),
			"membership_request_revoked_pn": `["5521971532700@s.whatsapp.net"]`,
		},
		RawPayload: info,
		CapturedAt: time.Now(),
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	callback, err := transformer.transformGroupInfoEvent(context.Background(), logger, event)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if callback.Notification != "REVOKED_MEMBERSHIP_REQUESTS" {
		t.Fatalf("unexpected notification %s", callback.Notification)
	}
	if callback.RequestMethod != "" {
		t.Fatalf("expected empty request method, got %s", callback.RequestMethod)
	}
	if callback.ParticipantPhone != "5521971532700" {
		t.Fatalf("expected participant phone fallback, got %s", callback.ParticipantPhone)
	}
	if len(callback.NotificationParameters) != 1 {
		t.Fatalf("expected 1 participant parameter, got %d", len(callback.NotificationParameters))
	}
}
