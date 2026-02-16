package interactive

import (
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	// Phone validation regex - accepts digits only (10-15 chars)
	phoneDigitsRegex = regexp.MustCompile(`^\d{10,15}$`)

	// CPF validation regex - 11 digits
	cpfRegex = regexp.MustCompile(`^\d{11}$`)

	// CNPJ validation regex - 14 digits
	cnpjRegex = regexp.MustCompile(`^\d{14}$`)

	// UUID/EVP validation regex - 36 chars with hyphens
	uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

// Validator for interactive message requests
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new validator instance with custom validation rules
func NewValidator() *Validator {
	v := validator.New()

	// Register custom validation functions
	_ = v.RegisterValidation("phone", validatePhone)

	return &Validator{validate: v}
}

// validatePhone is a custom validator for phone numbers
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	normalized := NormalizePhone(phone)
	return len(normalized) >= 10 && len(normalized) <= 15
}

// NormalizePhone removes all non-digit characters from a phone number
func NormalizePhone(phone string) string {
	var normalized strings.Builder
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			normalized.WriteRune(c)
		}
	}
	return normalized.String()
}

// isValidURL checks if a string is a valid HTTP/HTTPS URL
func isValidURL(str string) bool {
	if str == "" {
		return false
	}
	u, err := url.Parse(str)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

// isValidEmail performs RFC 5322 compliant email validation
func isValidEmail(email string) bool {
	if len(email) < 5 {
		return false
	}
	_, err := mail.ParseAddress(email)
	return err == nil
}

// ValidateButtonList validates a SendButtonListRequest
func (v *Validator) ValidateButtonList(req *SendButtonListRequest) error {
	if err := v.validate.Struct(req); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Validate phone
	phone := NormalizePhone(req.Phone)
	if len(phone) < 10 || len(phone) > 15 {
		return fmt.Errorf("invalid phone number: must be 10-15 digits")
	}

	// Validate message
	if req.Message == "" {
		return fmt.Errorf("message is required")
	}
	if len(req.Message) > 4096 {
		return fmt.Errorf("message too long: max 4096 characters")
	}

	// Validate buttons
	if len(req.ButtonList.Buttons) == 0 {
		return fmt.Errorf("at least one button is required")
	}
	if len(req.ButtonList.Buttons) > 3 {
		return fmt.Errorf("maximum 3 buttons allowed, got %d", len(req.ButtonList.Buttons))
	}

	// Validate each button
	for i, btn := range req.ButtonList.Buttons {
		if btn.ID == "" {
			return fmt.Errorf("button %d: id is required", i)
		}
		if len(btn.ID) > 256 {
			return fmt.Errorf("button %d: id too long, max 256 characters", i)
		}
		if btn.Label == "" {
			return fmt.Errorf("button %d: label is required", i)
		}
		if len(btn.Label) > 20 {
			return fmt.Errorf("button %d: label too long, max 20 characters", i)
		}
	}

	// Validate optional title
	if req.Title != nil && len(*req.Title) > 60 {
		return fmt.Errorf("title too long: max 60 characters")
	}

	// Validate optional footer
	if req.Footer != nil && len(*req.Footer) > 60 {
		return fmt.Errorf("footer too long: max 60 characters")
	}

	// Validate optional image URL
	if req.Image != nil && *req.Image != "" && !isValidURL(*req.Image) {
		// Could be base64, allow it for now
		if len(*req.Image) < 100 {
			return fmt.Errorf("image must be a valid URL or base64 encoded")
		}
	}

	// Validate optional video URL
	if req.Video != nil && *req.Video != "" && !isValidURL(*req.Video) {
		// Could be base64, allow it for now
		if len(*req.Video) < 100 {
			return fmt.Errorf("video must be a valid URL or base64 encoded")
		}
	}

	return nil
}

// ValidateButtonActions validates a SendButtonActionsRequest
func (v *Validator) ValidateButtonActions(req *SendButtonActionsRequest) error {
	if err := v.validate.Struct(req); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Validate phone
	phone := NormalizePhone(req.Phone)
	if len(phone) < 10 || len(phone) > 15 {
		return fmt.Errorf("invalid phone number: must be 10-15 digits")
	}

	// Validate message
	if req.Message == "" {
		return fmt.Errorf("message is required")
	}
	if len(req.Message) > 4096 {
		return fmt.Errorf("message too long: max 4096 characters")
	}

	// Validate buttons
	if len(req.ButtonActions.Buttons) == 0 {
		return fmt.Errorf("at least one button is required")
	}
	if len(req.ButtonActions.Buttons) > 3 {
		return fmt.Errorf("maximum 3 buttons allowed, got %d", len(req.ButtonActions.Buttons))
	}

	// Validate each button based on type
	for i, btn := range req.ButtonActions.Buttons {
		if btn.ID == "" {
			return fmt.Errorf("button %d: id is required", i)
		}
		if len(btn.ID) > 256 {
			return fmt.Errorf("button %d: id too long, max 256 characters", i)
		}
		if btn.Label == "" {
			return fmt.Errorf("button %d: label is required", i)
		}
		if len(btn.Label) > 20 {
			return fmt.Errorf("button %d: label too long, max 20 characters", i)
		}

		// Validate type-specific fields using normalized type
		// Accepts both Zé da API uppercase (CALL, URL) and lowercase (cta_call, cta_url)
		normalizedType := btn.GetNormalizedType()
		switch normalizedType {
		case ButtonTypeQuickReply:
			// No additional fields required
		case ButtonTypeCTAURL:
			if btn.URL == nil || *btn.URL == "" {
				return fmt.Errorf("button %d: url is required for cta_url type", i)
			}
			if !isValidURL(*btn.URL) {
				return fmt.Errorf("button %d: invalid url format", i)
			}
		case ButtonTypeCTACall:
			if btn.Phone == nil || *btn.Phone == "" {
				return fmt.Errorf("button %d: phone is required for cta_call type", i)
			}
			callPhone := NormalizePhone(*btn.Phone)
			if len(callPhone) < 10 || len(callPhone) > 15 {
				return fmt.Errorf("button %d: invalid phone number format", i)
			}
		case ButtonTypeCTACopy:
			if btn.CopyCode == nil || *btn.CopyCode == "" {
				return fmt.Errorf("button %d: copyCode is required for cta_copy type", i)
			}
		case ButtonTypePaymentInfo:
			// Payment info buttons have optional fields
		case ButtonTypeReviewAndPay:
			// Review and pay buttons have optional fields
		default:
			return fmt.Errorf("button %d: unsupported button type: %s", i, btn.Type)
		}
	}

	// Validate optional title
	if req.Title != nil && len(*req.Title) > 60 {
		return fmt.Errorf("title too long: max 60 characters")
	}

	// Validate optional footer
	if req.Footer != nil && len(*req.Footer) > 60 {
		return fmt.Errorf("footer too long: max 60 characters")
	}

	return nil
}

// ValidateOptionList validates a SendOptionListRequest
func (v *Validator) ValidateOptionList(req *SendOptionListRequest) error {
	if err := v.validate.Struct(req); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Validate phone
	phone := NormalizePhone(req.Phone)
	if len(phone) < 10 || len(phone) > 15 {
		return fmt.Errorf("invalid phone number: must be 10-15 digits")
	}

	// Validate message
	if req.Message == "" {
		return fmt.Errorf("message is required")
	}
	if len(req.Message) > 4096 {
		return fmt.Errorf("message too long: max 4096 characters")
	}

	// Validate button label
	if req.ButtonLabel == "" {
		return fmt.Errorf("buttonLabel is required")
	}
	if len(req.ButtonLabel) > 20 {
		return fmt.Errorf("buttonLabel too long: max 20 characters")
	}

	// Validate sections
	if len(req.OptionList.Sections) == 0 {
		return fmt.Errorf("at least one section is required")
	}
	if len(req.OptionList.Sections) > 10 {
		return fmt.Errorf("maximum 10 sections allowed, got %d", len(req.OptionList.Sections))
	}

	// Count total rows
	totalRows := 0
	for i, sec := range req.OptionList.Sections {
		if sec.Title == "" {
			return fmt.Errorf("section %d: title is required", i)
		}
		if len(sec.Title) > 24 {
			return fmt.Errorf("section %d: title too long, max 24 characters", i)
		}

		if len(sec.Rows) == 0 {
			return fmt.Errorf("section %d: at least one row is required", i)
		}
		if len(sec.Rows) > 10 {
			return fmt.Errorf("section %d: maximum 10 rows per section, got %d", i, len(sec.Rows))
		}

		for j, row := range sec.Rows {
			if row.ID == "" {
				return fmt.Errorf("section %d row %d: id is required", i, j)
			}
			if len(row.ID) > 200 {
				return fmt.Errorf("section %d row %d: id too long, max 200 characters", i, j)
			}
			if row.Title == "" {
				return fmt.Errorf("section %d row %d: title is required", i, j)
			}
			if len(row.Title) > 24 {
				return fmt.Errorf("section %d row %d: title too long, max 24 characters", i, j)
			}
			if row.Description != nil && len(*row.Description) > 72 {
				return fmt.Errorf("section %d row %d: description too long, max 72 characters", i, j)
			}
			totalRows++
		}
	}

	if totalRows > 10 {
		return fmt.Errorf("maximum 10 total rows allowed across all sections, got %d", totalRows)
	}

	// Validate optional title
	if req.Title != nil && len(*req.Title) > 60 {
		return fmt.Errorf("title too long: max 60 characters")
	}

	// Validate optional footer
	if req.Footer != nil && len(*req.Footer) > 60 {
		return fmt.Errorf("footer too long: max 60 characters")
	}

	return nil
}

// ValidateButtonPIX validates a SendButtonPIXRequest
func (v *Validator) ValidateButtonPIX(req *SendButtonPIXRequest) error {
	if err := v.validate.Struct(req); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Validate phone
	phone := NormalizePhone(req.Phone)
	if len(phone) < 10 || len(phone) > 15 {
		return fmt.Errorf("invalid phone number: must be 10-15 digits")
	}

	// Validate PIX key
	if req.PIXKey == "" {
		return fmt.Errorf("pixKey is required")
	}

	// Validate PIX key format based on type
	switch req.Type {
	case PIXKeyTypeCPF:
		normalized := NormalizePhone(req.PIXKey) // Reuse to strip non-digits
		if !cpfRegex.MatchString(normalized) {
			return fmt.Errorf("CPF key must have exactly 11 digits")
		}
	case PIXKeyTypeCNPJ:
		normalized := NormalizePhone(req.PIXKey)
		if !cnpjRegex.MatchString(normalized) {
			return fmt.Errorf("CNPJ key must have exactly 14 digits")
		}
	case PIXKeyTypeEmail:
		if !isValidEmail(req.PIXKey) {
			return fmt.Errorf("invalid email format for PIX key")
		}
	case PIXKeyTypePhone:
		normalized := NormalizePhone(req.PIXKey)
		if len(normalized) < 10 || len(normalized) > 15 {
			return fmt.Errorf("invalid phone format for PIX key")
		}
	case PIXKeyTypeEVP:
		if !uuidRegex.MatchString(req.PIXKey) {
			return fmt.Errorf("EVP key must be a valid UUID (36 characters with hyphens)")
		}
	default:
		return fmt.Errorf("unsupported PIX key type: %s", req.Type)
	}

	// Validate optional message
	if req.Message != nil && len(*req.Message) > 4096 {
		return fmt.Errorf("message too long: max 4096 characters")
	}

	// Validate optional amount
	if req.Amount != nil && *req.Amount < 0 {
		return fmt.Errorf("amount cannot be negative")
	}

	return nil
}

// ValidateButtonOTP validates a SendButtonOTPRequest
func (v *Validator) ValidateButtonOTP(req *SendButtonOTPRequest) error {
	if err := v.validate.Struct(req); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Validate phone
	phone := NormalizePhone(req.Phone)
	if len(phone) < 10 || len(phone) > 15 {
		return fmt.Errorf("invalid phone number: must be 10-15 digits")
	}

	// Validate message
	if req.Message == "" {
		return fmt.Errorf("message is required")
	}
	if len(req.Message) > 4096 {
		return fmt.Errorf("message too long: max 4096 characters")
	}

	// Validate code
	if req.Code == "" {
		return fmt.Errorf("code is required")
	}
	if len(req.Code) > 20 {
		return fmt.Errorf("code too long: max 20 characters")
	}

	// Validate optional title
	if req.Title != nil && len(*req.Title) > 60 {
		return fmt.Errorf("title too long: max 60 characters")
	}

	// Validate optional footer
	if req.Footer != nil && len(*req.Footer) > 60 {
		return fmt.Errorf("footer too long: max 60 characters")
	}

	return nil
}

// ValidateCarousel validates a SendCarouselRequest
func (v *Validator) ValidateCarousel(req *SendCarouselRequest) error {
	if err := v.validate.Struct(req); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Validate phone
	phone := NormalizePhone(req.Phone)
	if len(phone) < 10 || len(phone) > 15 {
		return fmt.Errorf("invalid phone number: must be 10-15 digits")
	}

	// Validate cards count
	if len(req.Cards) == 0 {
		return fmt.Errorf("at least one card is required")
	}
	if len(req.Cards) > 10 {
		return fmt.Errorf("maximum 10 cards allowed, got %d", len(req.Cards))
	}

	// Validate each card
	for i, card := range req.Cards {
		// Validate body (required)
		if card.Body.Text == "" {
			return fmt.Errorf("card %d: body text is required", i)
		}
		if len(card.Body.Text) > 1024 {
			return fmt.Errorf("card %d: body text too long, max 1024 characters", i)
		}

		// Validate optional header
		if card.Header != nil && len(card.Header.Text) > 60 {
			return fmt.Errorf("card %d: header text too long, max 60 characters", i)
		}

		// Validate optional footer
		if card.Footer != nil && len(card.Footer.Text) > 60 {
			return fmt.Errorf("card %d: footer text too long, max 60 characters", i)
		}

		// Validate buttons
		if len(card.Buttons) == 0 {
			return fmt.Errorf("card %d: at least one button is required", i)
		}
		if len(card.Buttons) > 3 {
			return fmt.Errorf("card %d: maximum 3 buttons per card, got %d", i, len(card.Buttons))
		}

		// Validate each button in the card
		for j, btn := range card.Buttons {
			if btn.ID == "" {
				return fmt.Errorf("card %d button %d: id is required", i, j)
			}
			if len(btn.ID) > 256 {
				return fmt.Errorf("card %d button %d: id too long, max 256 characters", i, j)
			}
			if btn.Label == "" {
				return fmt.Errorf("card %d button %d: label is required", i, j)
			}
			if len(btn.Label) > 20 {
				return fmt.Errorf("card %d button %d: label too long, max 20 characters", i, j)
			}

			// Validate type-specific fields using normalized type
			normalizedType := btn.GetNormalizedType()
			switch normalizedType {
			case ButtonTypeQuickReply:
				// No additional fields required
			case ButtonTypeCTAURL:
				if btn.URL == nil || *btn.URL == "" {
					return fmt.Errorf("card %d button %d: url is required for cta_url type", i, j)
				}
				if !isValidURL(*btn.URL) {
					return fmt.Errorf("card %d button %d: invalid url format", i, j)
				}
			case ButtonTypeCTACall:
				if btn.Phone == nil || *btn.Phone == "" {
					return fmt.Errorf("card %d button %d: phone is required for cta_call type", i, j)
				}
				callPhone := NormalizePhone(*btn.Phone)
				if len(callPhone) < 10 || len(callPhone) > 15 {
					return fmt.Errorf("card %d button %d: invalid phone number format", i, j)
				}
			case ButtonTypeCTACopy:
				if btn.CopyCode == nil || *btn.CopyCode == "" {
					return fmt.Errorf("card %d button %d: copyCode is required for cta_copy type", i, j)
				}
			default:
				return fmt.Errorf("card %d button %d: unsupported button type for carousel: %s", i, j, btn.Type)
			}
		}

		// Validate optional media
		if card.MediaURL != nil && *card.MediaURL != "" {
			if !isValidURL(*card.MediaURL) {
				// Could be base64, allow if long enough
				if len(*card.MediaURL) < 100 {
					return fmt.Errorf("card %d: mediaUrl must be a valid URL or base64 encoded", i)
				}
			}
		}
	}

	// Validate card type if provided
	if req.CardType != "" && req.CardType != CarouselCardTypeHScroll && req.CardType != CarouselCardTypeAlbum {
		return fmt.Errorf("invalid cardType: must be HSCROLL_CARDS or ALBUM_IMAGE")
	}

	return nil
}
