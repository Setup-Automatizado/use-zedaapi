package interactive

import (
	"encoding/json"
	"testing"
)

func TestNewProtoBuilder(t *testing.T) {
	b := NewProtoBuilder()
	if b == nil {
		t.Fatal("NewProtoBuilder returned nil")
	}
}

func TestBuildButtonListMessage(t *testing.T) {
	b := NewProtoBuilder()

	tests := []struct {
		name       string
		req        *SendButtonListRequest
		wantErr    bool
		checkProto func(t *testing.T, msgBytes []byte)
	}{
		{
			name: "basic button list",
			req: &SendButtonListRequest{
				Phone:   "5511999999999",
				Message: "Choose an option",
				ButtonList: ButtonListPayload{
					Buttons: []ButtonListItem{
						{ID: "btn1", Label: "Option 1"},
						{ID: "btn2", Label: "Option 2"},
					},
				},
			},
			wantErr: false,
			checkProto: func(t *testing.T, msgBytes []byte) {
				// Verify the message structure contains expected button params
				if len(msgBytes) == 0 {
					t.Error("expected non-empty message")
				}
			},
		},
		{
			name: "button list with title and footer",
			req: &SendButtonListRequest{
				Phone:   "5511999999999",
				Message: "Choose an option",
				Title:   ptrString("Title"),
				Footer:  ptrString("Footer"),
				ButtonList: ButtonListPayload{
					Buttons: []ButtonListItem{
						{ID: "btn1", Label: "Option 1"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "button list with three buttons",
			req: &SendButtonListRequest{
				Phone:   "5511999999999",
				Message: "Choose an option",
				ButtonList: ButtonListPayload{
					Buttons: []ButtonListItem{
						{ID: "btn1", Label: "Option 1"},
						{ID: "btn2", Label: "Option 2"},
						{ID: "btn3", Label: "Option 3"},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := b.BuildButtonListMessage(tt.req)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if msg == nil {
				t.Fatal("expected message, got nil")
			}
			if msg.InteractiveMessage == nil {
				t.Fatal("expected InteractiveMessage, got nil")
			}
			if msg.InteractiveMessage.Body == nil {
				t.Fatal("expected Body, got nil")
			}
			if msg.InteractiveMessage.Body.GetText() != tt.req.Message {
				t.Errorf("body text = %q, want %q", msg.InteractiveMessage.Body.GetText(), tt.req.Message)
			}
		})
	}
}

func TestBuildButtonActionsMessage(t *testing.T) {
	b := NewProtoBuilder()

	tests := []struct {
		name    string
		req     *SendButtonActionsRequest
		wantErr bool
	}{
		{
			name: "quick reply buttons",
			req: &SendButtonActionsRequest{
				Phone:   "5511999999999",
				Message: "Choose an action",
				ButtonActions: ButtonActionsPayload{
					Buttons: []ActionButton{
						{ID: "btn1", Label: "Reply 1", Type: ButtonTypeQuickReply},
						{ID: "btn2", Label: "Reply 2", Type: ButtonTypeQuickReply},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "URL button",
			req: &SendButtonActionsRequest{
				Phone:   "5511999999999",
				Message: "Visit website",
				ButtonActions: ButtonActionsPayload{
					Buttons: []ActionButton{
						{ID: "btn1", Label: "Visit", Type: ButtonTypeCTAURL, URL: ptrString("https://example.com")},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "call button",
			req: &SendButtonActionsRequest{
				Phone:   "5511999999999",
				Message: "Call us",
				ButtonActions: ButtonActionsPayload{
					Buttons: []ActionButton{
						{ID: "btn1", Label: "Call", Type: ButtonTypeCTACall, Phone: ptrString("5511999999999")},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "copy button",
			req: &SendButtonActionsRequest{
				Phone:   "5511999999999",
				Message: "Copy code",
				ButtonActions: ButtonActionsPayload{
					Buttons: []ActionButton{
						{ID: "btn1", Label: "Copy", Type: ButtonTypeCTACopy, CopyCode: ptrString("ABC123")},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "mixed button types - reply with action buttons should fail",
			req: &SendButtonActionsRequest{
				Phone:   "5511999999999",
				Message: "Choose an action",
				Title:   ptrString("Actions"),
				Footer:  ptrString("Tap a button"),
				ButtonActions: ButtonActionsPayload{
					Buttons: []ActionButton{
						{ID: "btn1", Label: "Reply", Type: ButtonTypeQuickReply},
						{ID: "btn2", Label: "Visit", Type: ButtonTypeCTAURL, URL: ptrString("https://example.com")},
						{ID: "btn3", Label: "Copy", Type: ButtonTypeCTACopy, CopyCode: ptrString("CODE")},
					},
				},
			},
			wantErr: true, // quick_reply buttons cannot be mixed with action buttons
		},
		{
			name: "multiple action buttons - same type allowed",
			req: &SendButtonActionsRequest{
				Phone:   "5511999999999",
				Message: "Choose an action",
				Title:   ptrString("Actions"),
				Footer:  ptrString("Tap a button"),
				ButtonActions: ButtonActionsPayload{
					Buttons: []ActionButton{
						{ID: "btn1", Label: "Visit Site", Type: ButtonTypeCTAURL, URL: ptrString("https://example.com")},
						{ID: "btn2", Label: "Copy Code", Type: ButtonTypeCTACopy, CopyCode: ptrString("CODE123")},
						{ID: "btn3", Label: "Call Us", Type: ButtonTypeCTACall, Phone: ptrString("+5511999999999")},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "unsupported button type",
			req: &SendButtonActionsRequest{
				Phone:   "5511999999999",
				Message: "Test",
				ButtonActions: ButtonActionsPayload{
					Buttons: []ActionButton{
						{ID: "btn1", Label: "Test", Type: "unsupported"},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := b.BuildButtonActionsMessage(tt.req)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if msg == nil {
				t.Fatal("expected message, got nil")
			}
			if msg.InteractiveMessage == nil {
				t.Fatal("expected InteractiveMessage, got nil")
			}
		})
	}
}

func TestBuildOptionListMessage(t *testing.T) {
	b := NewProtoBuilder()

	tests := []struct {
		name    string
		req     *SendOptionListRequest
		wantErr bool
	}{
		{
			name: "basic option list",
			req: &SendOptionListRequest{
				Phone:       "5511999999999",
				Message:     "Select an option",
				ButtonLabel: "Options",
				OptionList: OptionListPayload{
					Sections: []OptionSection{
						{
							Title: "Section 1",
							Rows: []OptionRow{
								{ID: "row1", Title: "Row 1"},
								{ID: "row2", Title: "Row 2"},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "option list with descriptions",
			req: &SendOptionListRequest{
				Phone:       "5511999999999",
				Message:     "Select an option",
				ButtonLabel: "Options",
				OptionList: OptionListPayload{
					Sections: []OptionSection{
						{
							Title: "Section 1",
							Rows: []OptionRow{
								{ID: "row1", Title: "Row 1", Description: ptrString("Description 1")},
								{ID: "row2", Title: "Row 2", Description: ptrString("Description 2")},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "option list with multiple sections",
			req: &SendOptionListRequest{
				Phone:       "5511999999999",
				Message:     "Select an option",
				Title:       ptrString("Menu"),
				Footer:      ptrString("Tap to select"),
				ButtonLabel: "Options",
				OptionList: OptionListPayload{
					Sections: []OptionSection{
						{
							Title: "Section 1",
							Rows: []OptionRow{
								{ID: "row1", Title: "Row 1"},
							},
						},
						{
							Title: "Section 2",
							Rows: []OptionRow{
								{ID: "row2", Title: "Row 2"},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := b.BuildOptionListMessage(tt.req)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if msg == nil {
				t.Fatal("expected message, got nil")
			}
			if msg.ListMessage == nil {
				t.Fatal("expected ListMessage, got nil")
			}
			if msg.ListMessage.GetDescription() != tt.req.Message {
				t.Errorf("description = %q, want %q", msg.ListMessage.GetDescription(), tt.req.Message)
			}
			if msg.ListMessage.GetButtonText() != tt.req.ButtonLabel {
				t.Errorf("buttonText = %q, want %q", msg.ListMessage.GetButtonText(), tt.req.ButtonLabel)
			}
		})
	}
}

func TestBuildPIXButtonMessage(t *testing.T) {
	b := NewProtoBuilder()

	tests := []struct {
		name    string
		req     *SendButtonPIXRequest
		wantErr bool
	}{
		{
			name: "basic PIX button",
			req: &SendButtonPIXRequest{
				Phone:  "5511999999999",
				PIXKey: "12345678901",
				Type:   PIXKeyTypeCPF,
				Name:   ptrString("Test Merchant"), // merchant name is required
			},
			wantErr: false,
		},
		{
			name: "PIX with message and amount",
			req: &SendButtonPIXRequest{
				Phone:   "5511999999999",
				Message: ptrString("Payment for order #123"),
				PIXKey:  "12345678901",
				Type:    PIXKeyTypeCPF,
				Amount:  ptrFloat64(99.90),
				Name:    ptrString("John Doe"),
			},
			wantErr: false,
		},
		{
			name: "PIX with transaction ID",
			req: &SendButtonPIXRequest{
				Phone:         "5511999999999",
				PIXKey:        "test@email.com",
				Type:          PIXKeyTypeEmail,
				TransactionID: ptrString("TX123456"),
				Name:          ptrString("Test Merchant"), // merchant name is required
			},
			wantErr: false,
		},
		{
			name: "PIX without merchant name should fail",
			req: &SendButtonPIXRequest{
				Phone:  "5511999999999",
				PIXKey: "12345678901",
				Type:   PIXKeyTypeCPF,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := b.BuildPIXButtonMessage(tt.req)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if msg == nil {
				t.Fatal("expected message, got nil")
			}
			if msg.InteractiveMessage == nil {
				t.Fatal("expected InteractiveMessage, got nil")
			}
		})
	}
}

func TestBuildOTPButtonMessage(t *testing.T) {
	b := NewProtoBuilder()

	tests := []struct {
		name    string
		req     *SendButtonOTPRequest
		wantErr bool
	}{
		{
			name: "basic OTP button",
			req: &SendButtonOTPRequest{
				Phone:   "5511999999999",
				Message: "Your verification code is:",
				Code:    "123456",
			},
			wantErr: false,
		},
		{
			name: "OTP with title and footer",
			req: &SendButtonOTPRequest{
				Phone:   "5511999999999",
				Message: "Your verification code is:",
				Code:    "ABC123",
				Title:   ptrString("Verification"),
				Footer:  ptrString("Code expires in 10 minutes"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := b.BuildOTPButtonMessage(tt.req)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if msg == nil {
				t.Fatal("expected message, got nil")
			}
			if msg.InteractiveMessage == nil {
				t.Fatal("expected InteractiveMessage, got nil")
			}
			if msg.InteractiveMessage.Body == nil {
				t.Fatal("expected Body, got nil")
			}
			if msg.InteractiveMessage.Body.GetText() != tt.req.Message {
				t.Errorf("body text = %q, want %q", msg.InteractiveMessage.Body.GetText(), tt.req.Message)
			}
		})
	}
}

func TestAddContextInfo(t *testing.T) {
	b := NewProtoBuilder()

	t.Run("add context to interactive message", func(t *testing.T) {
		req := &SendButtonListRequest{
			Phone:   "5511999999999",
			Message: "Test",
			ButtonList: ButtonListPayload{
				Buttons: []ButtonListItem{
					{ID: "btn1", Label: "Option 1"},
				},
			},
		}

		msg, err := b.BuildButtonListMessage(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		b.AddContextInfo(msg, "reply-message-id", "5511888888888")

		if msg.InteractiveMessage.ContextInfo == nil {
			t.Fatal("expected ContextInfo, got nil")
		}
		if msg.InteractiveMessage.ContextInfo.GetStanzaID() != "reply-message-id" {
			t.Errorf("stanzaId = %q, want %q", msg.InteractiveMessage.ContextInfo.GetStanzaID(), "reply-message-id")
		}
		if msg.InteractiveMessage.ContextInfo.GetParticipant() != "5511888888888" {
			t.Errorf("participant = %q, want %q", msg.InteractiveMessage.ContextInfo.GetParticipant(), "5511888888888")
		}
	})

	t.Run("add context to list message", func(t *testing.T) {
		req := &SendOptionListRequest{
			Phone:       "5511999999999",
			Message:     "Test",
			ButtonLabel: "Options",
			OptionList: OptionListPayload{
				Sections: []OptionSection{
					{
						Title: "Section",
						Rows: []OptionRow{
							{ID: "row1", Title: "Row 1"},
						},
					},
				},
			},
		}

		msg, err := b.BuildOptionListMessage(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		b.AddContextInfo(msg, "reply-message-id", "5511888888888")

		if msg.ListMessage.ContextInfo == nil {
			t.Fatal("expected ContextInfo, got nil")
		}
	})

	t.Run("empty reply ID does nothing", func(t *testing.T) {
		req := &SendButtonListRequest{
			Phone:   "5511999999999",
			Message: "Test",
			ButtonList: ButtonListPayload{
				Buttons: []ButtonListItem{
					{ID: "btn1", Label: "Option 1"},
				},
			},
		}

		msg, err := b.BuildButtonListMessage(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		b.AddContextInfo(msg, "", "5511888888888")

		if msg.InteractiveMessage.ContextInfo != nil {
			t.Error("expected no ContextInfo when replyToMessageID is empty")
		}
	})
}

func TestButtonParamsJSON(t *testing.T) {
	b := NewProtoBuilder()

	t.Run("quick reply button params", func(t *testing.T) {
		req := &SendButtonListRequest{
			Phone:   "5511999999999",
			Message: "Test",
			ButtonList: ButtonListPayload{
				Buttons: []ButtonListItem{
					{ID: "btn1", Label: "Option 1"},
				},
			},
		}

		msg, err := b.BuildButtonListMessage(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		nativeFlow := msg.InteractiveMessage.GetNativeFlowMessage()
		if nativeFlow == nil {
			t.Fatal("expected NativeFlowMessage, got nil")
		}
		if len(nativeFlow.Buttons) != 1 {
			t.Fatalf("expected 1 button, got %d", len(nativeFlow.Buttons))
		}

		button := nativeFlow.Buttons[0]
		if button.GetName() != "quick_reply" {
			t.Errorf("button name = %q, want %q", button.GetName(), "quick_reply")
		}

		// Verify button params JSON - now uses map[string]interface{} due to disabled field (bool)
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(button.GetButtonParamsJSON()), &params); err != nil {
			t.Fatalf("failed to unmarshal button params: %v", err)
		}
		if params["id"] != "btn1" {
			t.Errorf("params[id] = %v, want %q", params["id"], "btn1")
		}
		if params["display_text"] != "Option 1" {
			t.Errorf("params[display_text] = %v, want %q", params["display_text"], "Option 1")
		}
		// Verify disabled field is present and false
		if disabled, ok := params["disabled"]; !ok {
			t.Error("params[disabled] not present")
		} else if disabled != false {
			t.Errorf("params[disabled] = %v, want false", disabled)
		}
	})

	t.Run("CTA URL button params", func(t *testing.T) {
		req := &SendButtonActionsRequest{
			Phone:   "5511999999999",
			Message: "Test",
			ButtonActions: ButtonActionsPayload{
				Buttons: []ActionButton{
					{ID: "btn1", Label: "Visit", Type: ButtonTypeCTAURL, URL: ptrString("https://example.com")},
				},
			},
		}

		msg, err := b.BuildButtonActionsMessage(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		nativeFlow := msg.InteractiveMessage.GetNativeFlowMessage()
		if nativeFlow == nil {
			t.Fatal("expected NativeFlowMessage, got nil")
		}

		button := nativeFlow.Buttons[0]
		if button.GetName() != "cta_url" {
			t.Errorf("button name = %q, want %q", button.GetName(), "cta_url")
		}

		// Now uses map[string]interface{} due to disabled field (bool)
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(button.GetButtonParamsJSON()), &params); err != nil {
			t.Fatalf("failed to unmarshal button params: %v", err)
		}
		if params["url"] != "https://example.com" {
			t.Errorf("params[url] = %v, want %q", params["url"], "https://example.com")
		}
		// Verify disabled field is present and false
		if disabled, ok := params["disabled"]; !ok {
			t.Error("params[disabled] not present")
		} else if disabled != false {
			t.Errorf("params[disabled] = %v, want false", disabled)
		}
	})
}
