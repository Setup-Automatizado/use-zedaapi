package interactive

import (
	"strings"
	"testing"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator()
	if v == nil {
		t.Fatal("NewValidator returned nil")
	}
	if v.validate == nil {
		t.Fatal("validator instance is nil")
	}
}

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"digits only", "5511999999999", "5511999999999"},
		{"with country code +", "+5511999999999", "5511999999999"},
		{"with dashes", "55-11-99999-9999", "5511999999999"},
		{"with parentheses", "(55)11999999999", "5511999999999"},
		{"with spaces", "55 11 99999 9999", "5511999999999"},
		{"mixed format", "+55 (11) 99999-9999", "5511999999999"},
		{"empty string", "", ""},
		{"letters only", "abc", ""},
		{"mixed letters and digits", "55abc11999", "5511999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizePhone(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizePhone(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Helper function to create string pointers
func ptrString(s string) *string {
	return &s
}

// Helper function to create float64 pointers
func ptrFloat64(f float64) *float64 {
	return &f
}

func TestValidateButtonList(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name        string
		req         *SendButtonListRequest
		wantErr     bool
		errContains string // substring to check in error message
	}{
		{
			name: "valid request with one button",
			req: &SendButtonListRequest{
				Phone:   "5511999999999",
				Message: "Choose an option",
				ButtonList: ButtonListPayload{
					Buttons: []ButtonListItem{
						{ID: "btn1", Label: "Option 1"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid request with three buttons",
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
		{
			name: "valid request with title and footer",
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
			name: "invalid phone - too short",
			req: &SendButtonListRequest{
				Phone:   "123",
				Message: "Choose an option",
				ButtonList: ButtonListPayload{
					Buttons: []ButtonListItem{
						{ID: "btn1", Label: "Option 1"},
					},
				},
			},
			wantErr:     true,
			errContains: "phone",
		},
		{
			name: "empty message",
			req: &SendButtonListRequest{
				Phone:   "5511999999999",
				Message: "",
				ButtonList: ButtonListPayload{
					Buttons: []ButtonListItem{
						{ID: "btn1", Label: "Option 1"},
					},
				},
			},
			wantErr:     true,
			errContains: "Message",
		},
		{
			name: "no buttons",
			req: &SendButtonListRequest{
				Phone:   "5511999999999",
				Message: "Choose an option",
				ButtonList: ButtonListPayload{
					Buttons: []ButtonListItem{},
				},
			},
			wantErr:     true,
			errContains: "Buttons",
		},
		{
			name: "too many buttons",
			req: &SendButtonListRequest{
				Phone:   "5511999999999",
				Message: "Choose an option",
				ButtonList: ButtonListPayload{
					Buttons: []ButtonListItem{
						{ID: "btn1", Label: "Option 1"},
						{ID: "btn2", Label: "Option 2"},
						{ID: "btn3", Label: "Option 3"},
						{ID: "btn4", Label: "Option 4"},
					},
				},
			},
			wantErr:     true,
			errContains: "Buttons",
		},
		{
			name: "button without ID",
			req: &SendButtonListRequest{
				Phone:   "5511999999999",
				Message: "Choose an option",
				ButtonList: ButtonListPayload{
					Buttons: []ButtonListItem{
						{ID: "", Label: "Option 1"},
					},
				},
			},
			wantErr:     true,
			errContains: "ID",
		},
		{
			name: "button without label",
			req: &SendButtonListRequest{
				Phone:   "5511999999999",
				Message: "Choose an option",
				ButtonList: ButtonListPayload{
					Buttons: []ButtonListItem{
						{ID: "btn1", Label: ""},
					},
				},
			},
			wantErr:     true,
			errContains: "Label",
		},
		{
			name: "button label too long",
			req: &SendButtonListRequest{
				Phone:   "5511999999999",
				Message: "Choose an option",
				ButtonList: ButtonListPayload{
					Buttons: []ButtonListItem{
						{ID: "btn1", Label: strings.Repeat("a", 21)},
					},
				},
			},
			wantErr:     true,
			errContains: "Label",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateButtonList(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateButtonList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
				}
			}
		})
	}
}

func TestValidateButtonActions(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name        string
		req         *SendButtonActionsRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "valid quick reply button",
			req: &SendButtonActionsRequest{
				Phone:   "5511999999999",
				Message: "Choose an action",
				ButtonActions: ButtonActionsPayload{
					Buttons: []ActionButton{
						{ID: "btn1", Label: "Reply", Type: ButtonTypeQuickReply},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid URL button",
			req: &SendButtonActionsRequest{
				Phone:   "5511999999999",
				Message: "Visit our site",
				ButtonActions: ButtonActionsPayload{
					Buttons: []ActionButton{
						{ID: "btn1", Label: "Visit", Type: ButtonTypeCTAURL, URL: ptrString("https://example.com")},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid call button",
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
			name: "valid copy button",
			req: &SendButtonActionsRequest{
				Phone:   "5511999999999",
				Message: "Copy this code",
				ButtonActions: ButtonActionsPayload{
					Buttons: []ActionButton{
						{ID: "btn1", Label: "Copy", Type: ButtonTypeCTACopy, CopyCode: ptrString("PROMO123")},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "URL button without URL",
			req: &SendButtonActionsRequest{
				Phone:   "5511999999999",
				Message: "Visit our site",
				ButtonActions: ButtonActionsPayload{
					Buttons: []ActionButton{
						{ID: "btn1", Label: "Visit", Type: ButtonTypeCTAURL},
					},
				},
			},
			wantErr:     true,
			errContains: "url is required",
		},
		{
			name: "call button without phone",
			req: &SendButtonActionsRequest{
				Phone:   "5511999999999",
				Message: "Call us",
				ButtonActions: ButtonActionsPayload{
					Buttons: []ActionButton{
						{ID: "btn1", Label: "Call", Type: ButtonTypeCTACall},
					},
				},
			},
			wantErr:     true,
			errContains: "phone is required",
		},
		{
			name: "copy button without code",
			req: &SendButtonActionsRequest{
				Phone:   "5511999999999",
				Message: "Copy this code",
				ButtonActions: ButtonActionsPayload{
					Buttons: []ActionButton{
						{ID: "btn1", Label: "Copy", Type: ButtonTypeCTACopy},
					},
				},
			},
			wantErr:     true,
			errContains: "copyCode is required",
		},
		{
			name: "empty message",
			req: &SendButtonActionsRequest{
				Phone:   "5511999999999",
				Message: "",
				ButtonActions: ButtonActionsPayload{
					Buttons: []ActionButton{
						{ID: "btn1", Label: "Reply", Type: ButtonTypeQuickReply},
					},
				},
			},
			wantErr:     true,
			errContains: "Message",
		},
		{
			name: "no buttons",
			req: &SendButtonActionsRequest{
				Phone:   "5511999999999",
				Message: "Choose",
				ButtonActions: ButtonActionsPayload{
					Buttons: []ActionButton{},
				},
			},
			wantErr:     true,
			errContains: "Buttons",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateButtonActions(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateButtonActions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
				}
			}
		})
	}
}

func TestValidateOptionList(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name        string
		req         *SendOptionListRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "valid request with one section",
			req: &SendOptionListRequest{
				Phone:       "5511999999999",
				Message:     "Select an option",
				ButtonLabel: "Select",
				OptionList: OptionListPayload{
					Sections: []OptionSection{
						{
							Title: "Section 1",
							Rows: []OptionRow{
								{ID: "row1", Title: "Row 1"},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid request with description",
			req: &SendOptionListRequest{
				Phone:       "5511999999999",
				Message:     "Select an option",
				ButtonLabel: "Select",
				OptionList: OptionListPayload{
					Sections: []OptionSection{
						{
							Title: "Section 1",
							Rows: []OptionRow{
								{ID: "row1", Title: "Row 1", Description: ptrString("Description")},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing button label",
			req: &SendOptionListRequest{
				Phone:       "5511999999999",
				Message:     "Select an option",
				ButtonLabel: "",
				OptionList: OptionListPayload{
					Sections: []OptionSection{
						{
							Title: "Section 1",
							Rows: []OptionRow{
								{ID: "row1", Title: "Row 1"},
							},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "ButtonLabel",
		},
		{
			name: "button label too long",
			req: &SendOptionListRequest{
				Phone:       "5511999999999",
				Message:     "Select an option",
				ButtonLabel: strings.Repeat("a", 21),
				OptionList: OptionListPayload{
					Sections: []OptionSection{
						{
							Title: "Section 1",
							Rows: []OptionRow{
								{ID: "row1", Title: "Row 1"},
							},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "ButtonLabel",
		},
		{
			name: "no sections",
			req: &SendOptionListRequest{
				Phone:       "5511999999999",
				Message:     "Select an option",
				ButtonLabel: "Select",
				OptionList: OptionListPayload{
					Sections: []OptionSection{},
				},
			},
			wantErr:     true,
			errContains: "Sections",
		},
		{
			name: "section without title",
			req: &SendOptionListRequest{
				Phone:       "5511999999999",
				Message:     "Select an option",
				ButtonLabel: "Select",
				OptionList: OptionListPayload{
					Sections: []OptionSection{
						{
							Title: "",
							Rows: []OptionRow{
								{ID: "row1", Title: "Row 1"},
							},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "Title",
		},
		{
			name: "section title too long",
			req: &SendOptionListRequest{
				Phone:       "5511999999999",
				Message:     "Select an option",
				ButtonLabel: "Select",
				OptionList: OptionListPayload{
					Sections: []OptionSection{
						{
							Title: strings.Repeat("a", 25),
							Rows: []OptionRow{
								{ID: "row1", Title: "Row 1"},
							},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "Title",
		},
		{
			name: "row without ID",
			req: &SendOptionListRequest{
				Phone:       "5511999999999",
				Message:     "Select an option",
				ButtonLabel: "Select",
				OptionList: OptionListPayload{
					Sections: []OptionSection{
						{
							Title: "Section 1",
							Rows: []OptionRow{
								{ID: "", Title: "Row 1"},
							},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "ID",
		},
		{
			name: "row description too long",
			req: &SendOptionListRequest{
				Phone:       "5511999999999",
				Message:     "Select an option",
				ButtonLabel: "Select",
				OptionList: OptionListPayload{
					Sections: []OptionSection{
						{
							Title: "Section 1",
							Rows: []OptionRow{
								{ID: "row1", Title: "Row 1", Description: ptrString(strings.Repeat("a", 73))},
							},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "Description",
		},
		{
			name: "too many rows total",
			req: &SendOptionListRequest{
				Phone:       "5511999999999",
				Message:     "Select an option",
				ButtonLabel: "Select",
				OptionList: OptionListPayload{
					Sections: []OptionSection{
						{
							Title: "Section 1",
							Rows: []OptionRow{
								{ID: "row1", Title: "Row 1"},
								{ID: "row2", Title: "Row 2"},
								{ID: "row3", Title: "Row 3"},
								{ID: "row4", Title: "Row 4"},
								{ID: "row5", Title: "Row 5"},
								{ID: "row6", Title: "Row 6"},
								{ID: "row7", Title: "Row 7"},
								{ID: "row8", Title: "Row 8"},
								{ID: "row9", Title: "Row 9"},
								{ID: "row10", Title: "Row 10"},
								{ID: "row11", Title: "Row 11"},
							},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "Rows",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateOptionList(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOptionList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
				}
			}
		})
	}
}

func TestValidateButtonPIX(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name        string
		req         *SendButtonPIXRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "valid CPF key",
			req: &SendButtonPIXRequest{
				Phone:  "5511999999999",
				PIXKey: "12345678901",
				Type:   PIXKeyTypeCPF,
			},
			wantErr: false,
		},
		{
			name: "valid CNPJ key",
			req: &SendButtonPIXRequest{
				Phone:  "5511999999999",
				PIXKey: "12345678901234",
				Type:   PIXKeyTypeCNPJ,
			},
			wantErr: false,
		},
		{
			name: "valid email key",
			req: &SendButtonPIXRequest{
				Phone:  "5511999999999",
				PIXKey: "email@example.com",
				Type:   PIXKeyTypeEmail,
			},
			wantErr: false,
		},
		{
			name: "valid phone key",
			req: &SendButtonPIXRequest{
				Phone:  "5511999999999",
				PIXKey: "5511999999999",
				Type:   PIXKeyTypePhone,
			},
			wantErr: false,
		},
		{
			name: "valid EVP key",
			req: &SendButtonPIXRequest{
				Phone:  "5511999999999",
				PIXKey: "12345678-1234-1234-1234-123456789012",
				Type:   PIXKeyTypeEVP,
			},
			wantErr: false,
		},
		{
			name: "valid with amount",
			req: &SendButtonPIXRequest{
				Phone:  "5511999999999",
				PIXKey: "12345678901",
				Type:   PIXKeyTypeCPF,
				Amount: ptrFloat64(100.50),
			},
			wantErr: false,
		},
		{
			name: "invalid CPF - wrong length",
			req: &SendButtonPIXRequest{
				Phone:  "5511999999999",
				PIXKey: "123456",
				Type:   PIXKeyTypeCPF,
			},
			wantErr:     true,
			errContains: "11 digits",
		},
		{
			name: "invalid CNPJ - wrong length",
			req: &SendButtonPIXRequest{
				Phone:  "5511999999999",
				PIXKey: "123456",
				Type:   PIXKeyTypeCNPJ,
			},
			wantErr:     true,
			errContains: "14 digits",
		},
		{
			name: "invalid email",
			req: &SendButtonPIXRequest{
				Phone:  "5511999999999",
				PIXKey: "notanemail",
				Type:   PIXKeyTypeEmail,
			},
			wantErr:     true,
			errContains: "email format",
		},
		{
			name: "invalid EVP - not UUID",
			req: &SendButtonPIXRequest{
				Phone:  "5511999999999",
				PIXKey: "not-a-valid-uuid",
				Type:   PIXKeyTypeEVP,
			},
			wantErr:     true,
			errContains: "valid UUID",
		},
		{
			name: "negative amount",
			req: &SendButtonPIXRequest{
				Phone:  "5511999999999",
				PIXKey: "12345678901",
				Type:   PIXKeyTypeCPF,
				Amount: ptrFloat64(-100),
			},
			wantErr:     true,
			errContains: "cannot be negative",
		},
		{
			name: "missing PIX key",
			req: &SendButtonPIXRequest{
				Phone:  "5511999999999",
				PIXKey: "",
				Type:   PIXKeyTypeCPF,
			},
			wantErr:     true,
			errContains: "PIXKey",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateButtonPIX(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateButtonPIX() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
				}
			}
		})
	}
}

func TestValidateButtonOTP(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name        string
		req         *SendButtonOTPRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "valid request",
			req: &SendButtonOTPRequest{
				Phone:   "5511999999999",
				Message: "Your verification code is below",
				Code:    "123456",
			},
			wantErr: false,
		},
		{
			name: "valid with title and footer",
			req: &SendButtonOTPRequest{
				Phone:   "5511999999999",
				Message: "Your verification code is below",
				Code:    "123456",
				Title:   ptrString("Verification"),
				Footer:  ptrString("Do not share"),
			},
			wantErr: false,
		},
		{
			name: "missing code",
			req: &SendButtonOTPRequest{
				Phone:   "5511999999999",
				Message: "Your verification code is below",
				Code:    "",
			},
			wantErr:     true,
			errContains: "Code",
		},
		{
			name: "code too long",
			req: &SendButtonOTPRequest{
				Phone:   "5511999999999",
				Message: "Your verification code is below",
				Code:    strings.Repeat("1", 21),
			},
			wantErr:     true,
			errContains: "Code",
		},
		{
			name: "missing message",
			req: &SendButtonOTPRequest{
				Phone:   "5511999999999",
				Message: "",
				Code:    "123456",
			},
			wantErr:     true,
			errContains: "Message",
		},
		{
			name: "invalid phone",
			req: &SendButtonOTPRequest{
				Phone:   "123",
				Message: "Your code",
				Code:    "123456",
			},
			wantErr:     true,
			errContains: "phone number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateButtonOTP(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateButtonOTP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
				}
			}
		})
	}
}
