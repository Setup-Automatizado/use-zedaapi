package interactive

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/proto/waE2E"
)

// ProtoBuilder creates whatsmeow proto messages from Zé da API requests
type ProtoBuilder struct{}

// NewProtoBuilder creates a new proto builder instance
func NewProtoBuilder() *ProtoBuilder {
	return &ProtoBuilder{}
}

// BuildButtonListMessage creates InteractiveMessage with NativeFlowMessage for reply buttons
func (b *ProtoBuilder) BuildButtonListMessage(req *SendButtonListRequest) (*waE2E.Message, error) {
	// Build buttons array for NativeFlowMessage
	buttons := make([]*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton, 0, len(req.ButtonList.Buttons))

	for _, btn := range req.ButtonList.Buttons {
		paramsJSON, err := json.Marshal(map[string]interface{}{
			"id":           btn.ID,
			"display_text": btn.Label,
			"disabled":     false,
		})
		if err != nil {
			return nil, fmt.Errorf("marshal button params: %w", err)
		}

		buttons = append(buttons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
			Name:             proto.String("quick_reply"),
			ButtonParamsJSON: proto.String(string(paramsJSON)),
		})
	}

	// Build body
	body := &waE2E.InteractiveMessage_Body{
		Text: proto.String(req.Message),
	}

	// Build optional header
	var header *waE2E.InteractiveMessage_Header
	if req.Title != nil && *req.Title != "" {
		header = &waE2E.InteractiveMessage_Header{
			Title:              proto.String(*req.Title),
			HasMediaAttachment: proto.Bool(false),
		}
	}

	// Build optional footer
	var footer *waE2E.InteractiveMessage_Footer
	if req.Footer != nil && *req.Footer != "" {
		footer = &waE2E.InteractiveMessage_Footer{
			Text: proto.String(*req.Footer),
		}
	}

	// Build native flow message
	msgVersion := int32(1)
	nativeFlow := &waE2E.InteractiveMessage_NativeFlowMessage{
		MessageVersion: &msgVersion,
		Buttons:        buttons,
	}

	msg := &waE2E.Message{
		InteractiveMessage: &waE2E.InteractiveMessage{
			Header: header,
			Body:   body,
			Footer: footer,
			InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
				NativeFlowMessage: nativeFlow,
			},
		},
	}

	return msg, nil
}

// BuildButtonListMessageWithMedia creates InteractiveMessage with NativeFlowMessage for reply buttons
// with a pre-processed media header (image/video with thumbnails properly uploaded)
func (b *ProtoBuilder) BuildButtonListMessageWithMedia(req *SendButtonListRequest, mediaHeader *waProto.InteractiveMessage_Header) (*waE2E.Message, error) {
	// Build buttons array for NativeFlowMessage
	buttons := make([]*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton, 0, len(req.ButtonList.Buttons))

	for _, btn := range req.ButtonList.Buttons {
		paramsJSON, err := json.Marshal(map[string]interface{}{
			"id":           btn.ID,
			"display_text": btn.Label,
			"disabled":     false,
		})
		if err != nil {
			return nil, fmt.Errorf("marshal button params: %w", err)
		}

		buttons = append(buttons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
			Name:             proto.String("quick_reply"),
			ButtonParamsJSON: proto.String(string(paramsJSON)),
		})
	}

	// Build body
	body := &waE2E.InteractiveMessage_Body{
		Text: proto.String(req.Message),
	}

	// Use the pre-processed media header if provided, otherwise create text-only header
	var header *waE2E.InteractiveMessage_Header
	if mediaHeader != nil {
		// Convert waProto.InteractiveMessage_Header to waE2E.InteractiveMessage_Header
		header = b.convertMediaHeader(mediaHeader, req.Title)
	} else if req.Title != nil && *req.Title != "" {
		header = &waE2E.InteractiveMessage_Header{
			Title:              proto.String(*req.Title),
			HasMediaAttachment: proto.Bool(false),
		}
	}

	// Build optional footer
	var footer *waE2E.InteractiveMessage_Footer
	if req.Footer != nil && *req.Footer != "" {
		footer = &waE2E.InteractiveMessage_Footer{
			Text: proto.String(*req.Footer),
		}
	}

	// Build native flow message
	msgVersion := int32(1)
	nativeFlow := &waE2E.InteractiveMessage_NativeFlowMessage{
		MessageVersion: &msgVersion,
		Buttons:        buttons,
	}

	msg := &waE2E.Message{
		InteractiveMessage: &waE2E.InteractiveMessage{
			Header: header,
			Body:   body,
			Footer: footer,
			InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
				NativeFlowMessage: nativeFlow,
			},
		},
	}

	return msg, nil
}

// BuildButtonActionsMessageWithMedia creates InteractiveMessage with action buttons
// with a pre-processed media header (image/video with thumbnails properly uploaded)
func (b *ProtoBuilder) BuildButtonActionsMessageWithMedia(req *SendButtonActionsRequest, mediaHeader *waProto.InteractiveMessage_Header) (*waE2E.Message, error) {
	buttons := make([]*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton, 0, len(req.ButtonActions.Buttons))

	// Track button types for mixing validation
	hasReplyButton := false
	hasActionButton := false

	for _, btn := range req.ButtonActions.Buttons {
		var paramsJSON []byte
		var buttonName string
		var err error

		normalizedType := btn.GetNormalizedType()

		switch normalizedType {
		case ButtonTypeQuickReply:
			hasReplyButton = true
			buttonName = "quick_reply"
			paramsJSON, err = json.Marshal(map[string]interface{}{
				"id":           btn.ID,
				"display_text": btn.Label,
				"disabled":     false,
			})

		case ButtonTypeCTAURL:
			hasActionButton = true
			buttonName = "cta_url"
			if btn.URL == nil || *btn.URL == "" {
				return nil, fmt.Errorf("URL is required for cta_url button")
			}
			paramsJSON, err = json.Marshal(map[string]interface{}{
				"display_text": btn.Label,
				"url":          *btn.URL,
				"disabled":     false,
			})

		case ButtonTypeCTACall:
			hasActionButton = true
			buttonName = "cta_call"
			if btn.Phone == nil || *btn.Phone == "" {
				return nil, fmt.Errorf("phone is required for cta_call button")
			}
			paramsJSON, err = json.Marshal(map[string]interface{}{
				"display_text": btn.Label,
				"phone_number": *btn.Phone,
				"disabled":     false,
			})

		case ButtonTypeCTACopy:
			hasActionButton = true
			buttonName = "cta_copy"
			if btn.CopyCode == nil || *btn.CopyCode == "" {
				return nil, fmt.Errorf("copyCode is required for cta_copy button")
			}
			paramsJSON, err = json.Marshal(map[string]interface{}{
				"display_text": btn.Label,
				"copy_code":    *btn.CopyCode,
				"disabled":     false,
			})

		case ButtonTypePaymentInfo:
			hasActionButton = true
			buttonName = "payment_info"
			pixParams, err := b.buildPIXPaymentInfoParams(&btn)
			if err != nil {
				return nil, fmt.Errorf("build payment_info params: %w", err)
			}
			paramsJSON, err = json.Marshal(pixParams)

		case ButtonTypeReviewAndPay:
			hasActionButton = true
			buttonName = "review_and_pay"
			reviewParams, err := b.buildReviewAndPayParams(&btn)
			if err != nil {
				return nil, fmt.Errorf("build review_and_pay params: %w", err)
			}
			paramsJSON, err = json.Marshal(reviewParams)

		default:
			return nil, fmt.Errorf("unsupported button type: %s (normalized: %s)", btn.Type, normalizedType)
		}

		if err != nil {
			return nil, fmt.Errorf("marshal button params: %w", err)
		}

		buttons = append(buttons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
			Name:             proto.String(buttonName),
			ButtonParamsJSON: proto.String(string(paramsJSON)),
		})
	}

	// Validate button mixing
	if hasReplyButton && hasActionButton {
		return nil, fmt.Errorf("quick_reply buttons cannot be mixed with action buttons")
	}

	// Build body
	body := &waE2E.InteractiveMessage_Body{
		Text: proto.String(req.Message),
	}

	// Use the pre-processed media header if provided
	var header *waE2E.InteractiveMessage_Header
	if mediaHeader != nil {
		header = b.convertMediaHeader(mediaHeader, req.Title)
	} else if req.Title != nil && *req.Title != "" {
		header = &waE2E.InteractiveMessage_Header{
			Title:              proto.String(*req.Title),
			HasMediaAttachment: proto.Bool(false),
		}
	}

	// Build optional footer
	var footer *waE2E.InteractiveMessage_Footer
	if req.Footer != nil && *req.Footer != "" {
		footer = &waE2E.InteractiveMessage_Footer{
			Text: proto.String(*req.Footer),
		}
	}

	// Build native flow message
	msgVersion := int32(1)

	// Check if any button is review_and_pay
	var messageParamsJSON *string
	for _, btn := range req.ButtonActions.Buttons {
		if btn.GetNormalizedType() == ButtonTypeReviewAndPay {
			empty := "{}"
			messageParamsJSON = &empty
			break
		}
	}

	msg := &waE2E.Message{
		InteractiveMessage: &waE2E.InteractiveMessage{
			Header: header,
			Body:   body,
			Footer: footer,
			InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
				NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
					MessageVersion:    &msgVersion,
					Buttons:           buttons,
					MessageParamsJSON: messageParamsJSON,
				},
			},
		},
	}

	return msg, nil
}

// BuildButtonActionsMessage creates InteractiveMessage with action buttons
func (b *ProtoBuilder) BuildButtonActionsMessage(req *SendButtonActionsRequest) (*waE2E.Message, error) {
	buttons := make([]*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton, 0, len(req.ButtonActions.Buttons))

	// Track button types for mixing validation
	hasReplyButton := false
	hasActionButton := false

	for _, btn := range req.ButtonActions.Buttons {
		var paramsJSON []byte
		var buttonName string
		var err error

		// Use normalized type to accept both Zé da API uppercase (CALL, URL) and lowercase (cta_call)
		normalizedType := btn.GetNormalizedType()

		switch normalizedType {
		case ButtonTypeQuickReply:
			hasReplyButton = true
			buttonName = "quick_reply"
			paramsJSON, err = json.Marshal(map[string]interface{}{
				"id":           btn.ID,
				"display_text": btn.Label,
				"disabled":     false,
			})

		case ButtonTypeCTAURL:
			hasActionButton = true
			buttonName = "cta_url"
			if btn.URL == nil || *btn.URL == "" {
				return nil, fmt.Errorf("URL is required for cta_url button")
			}
			paramsJSON, err = json.Marshal(map[string]interface{}{
				"display_text": btn.Label,
				"url":          *btn.URL,
				"disabled":     false,
			})

		case ButtonTypeCTACall:
			hasActionButton = true
			buttonName = "cta_call"
			if btn.Phone == nil || *btn.Phone == "" {
				return nil, fmt.Errorf("phone is required for cta_call button")
			}
			paramsJSON, err = json.Marshal(map[string]interface{}{
				"display_text": btn.Label,
				"phone_number": *btn.Phone,
				"disabled":     false,
			})

		case ButtonTypeCTACopy:
			hasActionButton = true
			buttonName = "cta_copy"
			if btn.CopyCode == nil || *btn.CopyCode == "" {
				return nil, fmt.Errorf("copyCode is required for cta_copy button")
			}
			paramsJSON, err = json.Marshal(map[string]interface{}{
				"display_text": btn.Label,
				"copy_code":    *btn.CopyCode,
				"disabled":     false,
			})

		case ButtonTypePaymentInfo:
			hasActionButton = true
			buttonName = "payment_info"
			// Build proper PIX payment_info structure matching whatsmeow pattern
			pixParams, err := b.buildPIXPaymentInfoParams(&btn)
			if err != nil {
				return nil, fmt.Errorf("build payment_info params: %w", err)
			}
			paramsJSON, err = json.Marshal(pixParams)

		case ButtonTypeReviewAndPay:
			hasActionButton = true
			buttonName = "review_and_pay"
			// Build proper review_and_pay structure matching whatsmeow pattern
			reviewParams, err := b.buildReviewAndPayParams(&btn)
			if err != nil {
				return nil, fmt.Errorf("build review_and_pay params: %w", err)
			}
			paramsJSON, err = json.Marshal(reviewParams)

		default:
			return nil, fmt.Errorf("unsupported button type: %s (normalized: %s)", btn.Type, normalizedType)
		}

		if err != nil {
			return nil, fmt.Errorf("marshal button params: %w", err)
		}

		buttons = append(buttons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
			Name:             proto.String(buttonName),
			ButtonParamsJSON: proto.String(string(paramsJSON)),
		})
	}

	// Validate button mixing: reply buttons cannot be mixed with action buttons
	if hasReplyButton && hasActionButton {
		return nil, fmt.Errorf("quick_reply buttons cannot be mixed with action buttons (cta_url, cta_call, cta_copy, payment_info, review_and_pay)")
	}

	// Build body
	body := &waE2E.InteractiveMessage_Body{
		Text: proto.String(req.Message),
	}

	// Build optional header
	var header *waE2E.InteractiveMessage_Header
	if req.Title != nil && *req.Title != "" {
		header = &waE2E.InteractiveMessage_Header{
			Title:              proto.String(*req.Title),
			HasMediaAttachment: proto.Bool(false),
		}
	}

	// Build optional footer
	var footer *waE2E.InteractiveMessage_Footer
	if req.Footer != nil && *req.Footer != "" {
		footer = &waE2E.InteractiveMessage_Footer{
			Text: proto.String(*req.Footer),
		}
	}

	// Build native flow message
	msgVersion := int32(1)

	// Check if any button is review_and_pay - requires MessageParamsJSON
	var messageParamsJSON *string
	for _, btn := range req.ButtonActions.Buttons {
		if btn.GetNormalizedType() == ButtonTypeReviewAndPay {
			empty := "{}"
			messageParamsJSON = &empty
			break
		}
	}

	msg := &waE2E.Message{
		InteractiveMessage: &waE2E.InteractiveMessage{
			Header: header,
			Body:   body,
			Footer: footer,
			InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
				NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
					MessageVersion:    &msgVersion,
					Buttons:           buttons,
					MessageParamsJSON: messageParamsJSON,
				},
			},
		},
	}

	return msg, nil
}

// BuildOptionListMessage creates ListMessage with sections
func (b *ProtoBuilder) BuildOptionListMessage(req *SendOptionListRequest) (*waE2E.Message, error) {
	sections := make([]*waE2E.ListMessage_Section, 0, len(req.OptionList.Sections))

	for _, sec := range req.OptionList.Sections {
		rows := make([]*waE2E.ListMessage_Row, 0, len(sec.Rows))

		for _, row := range sec.Rows {
			listRow := &waE2E.ListMessage_Row{
				RowID: proto.String(row.ID),
				Title: proto.String(row.Title),
			}
			if row.Description != nil && *row.Description != "" {
				listRow.Description = proto.String(*row.Description)
			}
			rows = append(rows, listRow)
		}

		sections = append(sections, &waE2E.ListMessage_Section{
			Title: proto.String(sec.Title),
			Rows:  rows,
		})
	}

	listMsg := &waE2E.ListMessage{
		Description: proto.String(req.Message),
		ButtonText:  proto.String(req.ButtonLabel),
		ListType:    waE2E.ListMessage_SINGLE_SELECT.Enum(),
		Sections:    sections,
	}

	if req.Title != nil && *req.Title != "" {
		listMsg.Title = proto.String(*req.Title)
	}

	if req.Footer != nil && *req.Footer != "" {
		listMsg.FooterText = proto.String(*req.Footer)
	}

	msg := &waE2E.Message{
		ListMessage: listMsg,
	}

	return msg, nil
}

// BuildPIXButtonMessage creates InteractiveMessage with PIX payment button
// Uses the proper nested structure matching whatsmeow patterns
func (b *ProtoBuilder) BuildPIXButtonMessage(req *SendButtonPIXRequest) (*waE2E.Message, error) {
	// Validate merchant name is required
	merchantName := ""
	if req.Name != nil && *req.Name != "" {
		merchantName = *req.Name
	}
	if merchantName == "" {
		return nil, fmt.Errorf("merchant name (name) is required for PIX payment")
	}

	// Build PIX payment settings with proper nested structure
	pixStaticCode := &PIXStaticCode{
		Key:          req.PIXKey,
		MerchantName: merchantName,
		KeyType:      string(req.Type),
	}

	paymentSettings := []PIXPaymentSettings{
		{
			Type:          "pix_static_code",
			PixStaticCode: pixStaticCode,
			Cards:         &PIXCards{Enabled: false},
		},
	}

	// Calculate amount (value and offset for decimal places)
	amountValue := 0
	amountOffset := 100 // Default 2 decimal places (divide by 100)
	if req.Amount != nil && *req.Amount > 0 {
		// Convert float to integer cents (multiply by 100)
		amountValue = int(*req.Amount * 100)
	}

	// Generate reference ID
	referenceID := ""
	if req.TransactionID != nil && *req.TransactionID != "" {
		referenceID = *req.TransactionID
	} else {
		referenceID = b.generateReferenceID()
	}

	// Build order with items
	order := PIXOrder{
		Status: "payment_requested",
		Items: []PIXItem{
			{
				Amount:     PIXTotalAmount{Value: amountValue, Offset: amountOffset},
				Name:       "",
				RetailerID: fmt.Sprintf("custom-item-%d", time.Now().UnixNano()),
				Quantity:   0,
			},
		},
		Subtotal:  PIXTotalAmount{Value: amountValue, Offset: amountOffset},
		OrderType: proto.String("ORDER_WITHOUT_AMOUNT"),
	}

	// Build complete PIX button params matching whatsmeow structure
	pixParams := PIXButtonParamsJSON{
		Currency:        "BRL",
		Type:            "physical-goods",
		TotalAmount:     PIXTotalAmount{Value: amountValue, Offset: amountOffset},
		ReferenceID:     referenceID,
		PaymentSettings: paymentSettings,
		Order:           order,
	}

	paramsJSON, err := json.Marshal(pixParams)
	if err != nil {
		return nil, fmt.Errorf("marshal PIX params: %w", err)
	}

	buttons := []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
		{
			Name:             proto.String("payment_info"),
			ButtonParamsJSON: proto.String(string(paramsJSON)),
		},
	}

	// Build body
	bodyText := "Pagamento via PIX"
	if req.Message != nil && *req.Message != "" {
		bodyText = *req.Message
	}

	body := &waE2E.InteractiveMessage_Body{
		Text: proto.String(bodyText),
	}

	// Build optional header
	var header *waE2E.InteractiveMessage_Header
	if req.Title != nil && *req.Title != "" {
		header = &waE2E.InteractiveMessage_Header{
			Title:              proto.String(*req.Title),
			HasMediaAttachment: proto.Bool(false),
		}
	}

	// Build optional footer
	var footer *waE2E.InteractiveMessage_Footer
	if req.Footer != nil && *req.Footer != "" {
		footer = &waE2E.InteractiveMessage_Footer{
			Text: proto.String(*req.Footer),
		}
	}

	// Build native flow message
	msgVersion := int32(1)
	msg := &waE2E.Message{
		InteractiveMessage: &waE2E.InteractiveMessage{
			Header: header,
			Body:   body,
			Footer: footer,
			InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
				NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
					MessageVersion: &msgVersion,
					Buttons:        buttons,
				},
			},
		},
	}

	return msg, nil
}

// BuildOTPButtonMessage creates InteractiveMessage with copy code button for OTP
func (b *ProtoBuilder) BuildOTPButtonMessage(req *SendButtonOTPRequest) (*waE2E.Message, error) {
	paramsJSON, err := json.Marshal(map[string]interface{}{
		"display_text": "Copiar Codigo",
		"copy_code":    req.Code,
		"disabled":     false,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal OTP params: %w", err)
	}

	buttons := []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
		{
			Name:             proto.String("cta_copy"),
			ButtonParamsJSON: proto.String(string(paramsJSON)),
		},
	}

	// Build body
	body := &waE2E.InteractiveMessage_Body{
		Text: proto.String(req.Message),
	}

	// Build optional header
	var header *waE2E.InteractiveMessage_Header
	if req.Title != nil && *req.Title != "" {
		header = &waE2E.InteractiveMessage_Header{
			Title:              proto.String(*req.Title),
			HasMediaAttachment: proto.Bool(false),
		}
	}

	// Build optional footer
	var footer *waE2E.InteractiveMessage_Footer
	if req.Footer != nil && *req.Footer != "" {
		footer = &waE2E.InteractiveMessage_Footer{
			Text: proto.String(*req.Footer),
		}
	}

	// Build native flow message
	msgVersion := int32(1)
	msg := &waE2E.Message{
		InteractiveMessage: &waE2E.InteractiveMessage{
			Header: header,
			Body:   body,
			Footer: footer,
			InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
				NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
					MessageVersion: &msgVersion,
					Buttons:        buttons,
				},
			},
		},
	}

	return msg, nil
}

// AddContextInfo adds reply context to an interactive message
func (b *ProtoBuilder) AddContextInfo(msg *waE2E.Message, replyToMessageID, participant string) {
	if replyToMessageID == "" {
		return
	}

	contextInfo := &waE2E.ContextInfo{
		StanzaID:      proto.String(replyToMessageID),
		Participant:   proto.String(participant),
		QuotedMessage: &waE2E.Message{},
	}

	if msg.InteractiveMessage != nil {
		msg.InteractiveMessage.ContextInfo = contextInfo
	} else if msg.ListMessage != nil {
		msg.ListMessage.ContextInfo = contextInfo
	}
}

// BuildCarouselMessage creates CarouselMessage with multiple interactive cards
func (b *ProtoBuilder) BuildCarouselMessage(req *SendCarouselRequest) (*waE2E.Message, error) {
	if len(req.Cards) == 0 {
		return nil, fmt.Errorf("carousel requires at least one card")
	}
	if len(req.Cards) > 10 {
		return nil, fmt.Errorf("carousel supports maximum 10 cards")
	}

	cards := make([]*waE2E.InteractiveMessage, 0, len(req.Cards))

	for i, card := range req.Cards {
		interactiveCard, err := b.buildCarouselCard(&card, i)
		if err != nil {
			return nil, fmt.Errorf("card %d: %w", i, err)
		}
		cards = append(cards, interactiveCard)
	}

	// Determine carousel card type
	cardType := waE2E.InteractiveMessage_CarouselMessage_HSCROLL_CARDS
	if req.CardType == CarouselCardTypeAlbum {
		cardType = waE2E.InteractiveMessage_CarouselMessage_ALBUM_IMAGE
	}

	carousel := &waE2E.InteractiveMessage_CarouselMessage{
		Cards:            cards,
		MessageVersion:   proto.Int32(1),
		CarouselCardType: &cardType,
	}

	// Wrap in InteractiveMessage for proper delivery
	// Include Body at the root InteractiveMessage level for the carousel body text
	interactiveMsg := &waE2E.InteractiveMessage{
		InteractiveMessage: &waE2E.InteractiveMessage_CarouselMessage_{
			CarouselMessage: carousel,
		},
	}

	// Add body text if provided (displayed above carousel cards)
	if req.Message != "" {
		interactiveMsg.Body = &waE2E.InteractiveMessage_Body{
			Text: proto.String(req.Message),
		}
	}

	msg := &waE2E.Message{
		InteractiveMessage: interactiveMsg,
	}

	return msg, nil
}

// buildCarouselCard creates a single card for carousel message
func (b *ProtoBuilder) buildCarouselCard(card *CarouselCard, index int) (*waE2E.InteractiveMessage, error) {
	// Build buttons using NativeFlowMessage
	buttons := make([]*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton, 0, len(card.Buttons))

	for _, btn := range card.Buttons {
		var paramsJSON []byte
		var buttonName string
		var err error

		normalizedType := btn.GetNormalizedType()

		switch normalizedType {
		case ButtonTypeQuickReply:
			buttonName = "quick_reply"
			paramsJSON, err = json.Marshal(map[string]interface{}{
				"id":           btn.ID,
				"display_text": btn.Label,
				"disabled":     false,
			})

		case ButtonTypeCTAURL:
			buttonName = "cta_url"
			if btn.URL == nil || *btn.URL == "" {
				return nil, fmt.Errorf("URL is required for cta_url button")
			}
			paramsJSON, err = json.Marshal(map[string]interface{}{
				"display_text": btn.Label,
				"url":          *btn.URL,
				"disabled":     false,
			})

		case ButtonTypeCTACall:
			buttonName = "cta_call"
			if btn.Phone == nil || *btn.Phone == "" {
				return nil, fmt.Errorf("phone is required for cta_call button")
			}
			paramsJSON, err = json.Marshal(map[string]interface{}{
				"display_text": btn.Label,
				"phone_number": *btn.Phone,
				"disabled":     false,
			})

		case ButtonTypeCTACopy:
			buttonName = "cta_copy"
			if btn.CopyCode == nil || *btn.CopyCode == "" {
				return nil, fmt.Errorf("copyCode is required for cta_copy button")
			}
			paramsJSON, err = json.Marshal(map[string]interface{}{
				"display_text": btn.Label,
				"copy_code":    *btn.CopyCode,
				"disabled":     false,
			})

		default:
			return nil, fmt.Errorf("unsupported button type for carousel: %s", btn.Type)
		}

		if err != nil {
			return nil, fmt.Errorf("marshal button params: %w", err)
		}

		buttons = append(buttons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
			Name:             proto.String(buttonName),
			ButtonParamsJSON: proto.String(string(paramsJSON)),
		})
	}

	// Build body
	body := &waE2E.InteractiveMessage_Body{
		Text: proto.String(card.Body.Text),
	}

	// Build optional header
	var header *waE2E.InteractiveMessage_Header
	if card.Header != nil && card.Header.Text != "" {
		header = &waE2E.InteractiveMessage_Header{
			Title:              proto.String(card.Header.Text),
			HasMediaAttachment: proto.Bool(false),
		}
	}

	// Build optional footer
	var footer *waE2E.InteractiveMessage_Footer
	if card.Footer != nil && card.Footer.Text != "" {
		footer = &waE2E.InteractiveMessage_Footer{
			Text: proto.String(card.Footer.Text),
		}
	}

	// Build native flow message
	msgVersion := int32(1)
	nativeFlow := &waE2E.InteractiveMessage_NativeFlowMessage{
		MessageVersion: &msgVersion,
		Buttons:        buttons,
	}

	interactiveCard := &waE2E.InteractiveMessage{
		Header: header,
		Body:   body,
		Footer: footer,
		InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
			NativeFlowMessage: nativeFlow,
		},
	}

	return interactiveCard, nil
}

// buildActionButtons is a helper to build NativeFlowMessage buttons from ActionButton slice
func (b *ProtoBuilder) buildActionButtons(buttons []ActionButton) ([]*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton, error) {
	nativeButtons := make([]*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton, 0, len(buttons))

	for _, btn := range buttons {
		var paramsJSON []byte
		var buttonName string
		var err error

		normalizedType := btn.GetNormalizedType()

		switch normalizedType {
		case ButtonTypeQuickReply:
			buttonName = "quick_reply"
			paramsJSON, err = json.Marshal(map[string]string{
				"id":           btn.ID,
				"display_text": btn.Label,
			})

		case ButtonTypeCTAURL:
			buttonName = "cta_url"
			if btn.URL == nil || *btn.URL == "" {
				return nil, fmt.Errorf("URL is required for cta_url button")
			}
			paramsJSON, err = json.Marshal(map[string]string{
				"display_text": btn.Label,
				"url":          *btn.URL,
			})

		case ButtonTypeCTACall:
			buttonName = "cta_call"
			if btn.Phone == nil || *btn.Phone == "" {
				return nil, fmt.Errorf("phone is required for cta_call button")
			}
			paramsJSON, err = json.Marshal(map[string]string{
				"display_text": btn.Label,
				"phone_number": *btn.Phone,
			})

		case ButtonTypeCTACopy:
			buttonName = "cta_copy"
			if btn.CopyCode == nil || *btn.CopyCode == "" {
				return nil, fmt.Errorf("copyCode is required for cta_copy button")
			}
			paramsJSON, err = json.Marshal(map[string]string{
				"display_text": btn.Label,
				"copy_code":    *btn.CopyCode,
			})

		default:
			return nil, fmt.Errorf("unsupported button type: %s", btn.Type)
		}

		if err != nil {
			return nil, fmt.Errorf("marshal button params: %w", err)
		}

		nativeButtons = append(nativeButtons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
			Name:             proto.String(buttonName),
			ButtonParamsJSON: proto.String(string(paramsJSON)),
		})
	}

	return nativeButtons, nil
}

// generateReferenceID generates a random reference ID for PIX transactions
func (b *ProtoBuilder) generateReferenceID() string {
	bytes := make([]byte, 10)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("REF%d", time.Now().UnixNano())
	}
	return strings.ToUpper(hex.EncodeToString(bytes))
}

// buildPIXPaymentInfoParams builds the payment_info button params using proper whatsmeow structure
func (b *ProtoBuilder) buildPIXPaymentInfoParams(btn *ActionButton) (*PIXButtonParamsJSON, error) {
	// Validate required fields
	if btn.PIXKey == nil || *btn.PIXKey == "" {
		return nil, fmt.Errorf("PIX key is required for payment_info button")
	}
	if btn.MerchantName == nil || *btn.MerchantName == "" {
		return nil, fmt.Errorf("merchant name is required for payment_info button")
	}
	if btn.PIXKeyType == nil || *btn.PIXKeyType == "" {
		return nil, fmt.Errorf("PIX key type is required for payment_info button")
	}

	// Build PIX static code
	pixStaticCode := &PIXStaticCode{
		Key:          *btn.PIXKey,
		MerchantName: *btn.MerchantName,
		KeyType:      *btn.PIXKeyType,
	}

	paymentSettings := []PIXPaymentSettings{
		{
			Type:          "pix_static_code",
			PixStaticCode: pixStaticCode,
			Cards:         &PIXCards{Enabled: false},
		},
	}

	// Calculate amount
	amountValue := 0
	amountOffset := 100 // 2 decimal places
	if btn.Amount != nil && *btn.Amount > 0 {
		amountValue = int(*btn.Amount)
	}

	// Generate or use provided reference ID
	referenceID := b.generateReferenceID()
	if btn.ReferenceID != nil && *btn.ReferenceID != "" {
		referenceID = *btn.ReferenceID
	}

	// Build order
	order := PIXOrder{
		Status: "payment_requested",
		Items: []PIXItem{
			{
				Amount:     PIXTotalAmount{Value: amountValue, Offset: amountOffset},
				Name:       "",
				RetailerID: fmt.Sprintf("custom-item-%d", time.Now().UnixNano()),
				Quantity:   0,
			},
		},
		Subtotal:  PIXTotalAmount{Value: amountValue, Offset: amountOffset},
		OrderType: proto.String("ORDER_WITHOUT_AMOUNT"),
	}

	// Set currency
	currency := "BRL"
	if btn.Currency != nil && *btn.Currency != "" {
		currency = *btn.Currency
	}

	return &PIXButtonParamsJSON{
		Currency:        currency,
		Type:            "physical-goods",
		TotalAmount:     PIXTotalAmount{Value: amountValue, Offset: amountOffset},
		ReferenceID:     referenceID,
		PaymentSettings: paymentSettings,
		Order:           order,
	}, nil
}

// buildReviewAndPayParams builds the review_and_pay button params using proper whatsmeow structure
func (b *ProtoBuilder) buildReviewAndPayParams(btn *ActionButton) (*ReviewAndPayParamsJSON, error) {
	// Validate required fields
	if btn.PIXKey == nil || *btn.PIXKey == "" {
		return nil, fmt.Errorf("PIX key is required for review_and_pay button")
	}
	if btn.MerchantName == nil || *btn.MerchantName == "" {
		return nil, fmt.Errorf("merchant name is required for review_and_pay button")
	}
	if btn.PIXKeyType == nil || *btn.PIXKeyType == "" {
		return nil, fmt.Errorf("PIX key type is required for review_and_pay button")
	}

	// Build PIX static code
	pixStaticCode := &PIXStaticCode{
		Key:          *btn.PIXKey,
		MerchantName: *btn.MerchantName,
		KeyType:      *btn.PIXKeyType,
	}

	paymentSettings := []PIXPaymentSettings{
		{
			Type:          "pix_static_code",
			PixStaticCode: pixStaticCode,
			Cards:         &PIXCards{Enabled: false},
		},
	}

	// Calculate total amount
	totalValue := 0
	totalOffset := 100 // 2 decimal places
	if btn.TotalAmount != nil && *btn.TotalAmount > 0 {
		totalValue = int(*btn.TotalAmount)
	} else if btn.Amount != nil && *btn.Amount > 0 {
		totalValue = int(*btn.Amount)
	}

	// Build order items
	var items []PIXItem
	subtotalValue := 0
	if btn.Order != nil && len(btn.Order.Items) > 0 {
		for _, item := range btn.Order.Items {
			itemAmount := int(item.Price)
			items = append(items, PIXItem{
				Amount:     PIXTotalAmount{Value: itemAmount, Offset: totalOffset},
				Name:       item.Name,
				RetailerID: fmt.Sprintf("item-%d", time.Now().UnixNano()),
				Quantity:   item.Quantity,
			})
			subtotalValue += itemAmount * max(1, item.Quantity)
		}
	} else {
		// Default item
		items = []PIXItem{
			{
				Amount:     PIXTotalAmount{Value: totalValue, Offset: totalOffset},
				Name:       "",
				RetailerID: fmt.Sprintf("custom-item-%d", time.Now().UnixNano()),
				Quantity:   0,
			},
		}
		subtotalValue = totalValue
	}

	// If total not provided, use subtotal
	if totalValue == 0 {
		totalValue = subtotalValue
	}

	// Generate or use provided reference ID
	referenceID := b.generateReferenceID()
	if btn.ReferenceID != nil && *btn.ReferenceID != "" {
		referenceID = *btn.ReferenceID
	}

	// Build order
	order := PIXOrder{
		Status:   "pending",
		Items:    items,
		Subtotal: PIXTotalAmount{Value: subtotalValue, Offset: totalOffset},
	}

	// Set currency
	currency := "BRL"
	if btn.Currency != nil && *btn.Currency != "" {
		currency = *btn.Currency
	}

	return &ReviewAndPayParamsJSON{
		ReferenceID:          referenceID,
		Type:                 "physical-goods",
		PaymentType:          "br",
		PaymentConfiguration: "merchant_categorization_code",
		PaymentSettings:      paymentSettings,
		Currency:             currency,
		TotalAmount:          PIXTotalAmount{Value: totalValue, Offset: totalOffset},
		Order:                order,
		SharePaymentStatus:   false,
		Referral:             "chat_attachment",
	}, nil
}

// max returns the larger of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// convertMediaHeader converts waProto.InteractiveMessage_Header to waE2E.InteractiveMessage_Header
// This is needed because whatsmeow uses different proto packages for different purposes
func (b *ProtoBuilder) convertMediaHeader(mediaHeader *waProto.InteractiveMessage_Header, title *string) *waE2E.InteractiveMessage_Header {
	if mediaHeader == nil {
		return nil
	}

	header := &waE2E.InteractiveMessage_Header{
		HasMediaAttachment: proto.Bool(mediaHeader.GetHasMediaAttachment()),
	}

	// Set title if provided
	if title != nil && *title != "" {
		header.Title = proto.String(*title)
	} else if mediaHeader.GetTitle() != "" {
		header.Title = proto.String(mediaHeader.GetTitle())
	}

	// Convert media based on type
	switch m := mediaHeader.Media.(type) {
	case *waProto.InteractiveMessage_Header_ImageMessage:
		if m.ImageMessage != nil {
			header.Media = &waE2E.InteractiveMessage_Header_ImageMessage{
				ImageMessage: b.convertImageMessage(m.ImageMessage),
			}
		}
	case *waProto.InteractiveMessage_Header_VideoMessage:
		if m.VideoMessage != nil {
			header.Media = &waE2E.InteractiveMessage_Header_VideoMessage{
				VideoMessage: b.convertVideoMessage(m.VideoMessage),
			}
		}
	case *waProto.InteractiveMessage_Header_DocumentMessage:
		if m.DocumentMessage != nil {
			header.Media = &waE2E.InteractiveMessage_Header_DocumentMessage{
				DocumentMessage: b.convertDocumentMessage(m.DocumentMessage),
			}
		}
	case *waProto.InteractiveMessage_Header_JpegThumbnail:
		if len(m.JPEGThumbnail) > 0 {
			header.Media = &waE2E.InteractiveMessage_Header_JPEGThumbnail{
				JPEGThumbnail: m.JPEGThumbnail,
			}
		}
	}

	return header
}

// convertImageMessage converts waProto.ImageMessage to waE2E.ImageMessage
func (b *ProtoBuilder) convertImageMessage(img *waProto.ImageMessage) *waE2E.ImageMessage {
	if img == nil {
		return nil
	}

	result := &waE2E.ImageMessage{
		URL:           img.URL,
		DirectPath:    img.DirectPath,
		MediaKey:      img.MediaKey,
		Mimetype:      img.Mimetype,
		FileEncSHA256: img.FileEncSHA256,
		FileSHA256:    img.FileSHA256,
		FileLength:    img.FileLength,
		Width:         img.Width,
		Height:        img.Height,
		JPEGThumbnail: img.JPEGThumbnail,
	}

	if img.ThumbnailDirectPath != nil {
		result.ThumbnailDirectPath = img.ThumbnailDirectPath
	}
	if img.ThumbnailSHA256 != nil {
		result.ThumbnailSHA256 = img.ThumbnailSHA256
	}
	if img.ThumbnailEncSHA256 != nil {
		result.ThumbnailEncSHA256 = img.ThumbnailEncSHA256
	}

	return result
}

// convertVideoMessage converts waProto.VideoMessage to waE2E.VideoMessage
func (b *ProtoBuilder) convertVideoMessage(vid *waProto.VideoMessage) *waE2E.VideoMessage {
	if vid == nil {
		return nil
	}

	result := &waE2E.VideoMessage{
		URL:           vid.URL,
		DirectPath:    vid.DirectPath,
		MediaKey:      vid.MediaKey,
		Mimetype:      vid.Mimetype,
		FileEncSHA256: vid.FileEncSHA256,
		FileSHA256:    vid.FileSHA256,
		FileLength:    vid.FileLength,
		Width:         vid.Width,
		Height:        vid.Height,
		JPEGThumbnail: vid.JPEGThumbnail,
	}

	if vid.ThumbnailDirectPath != nil {
		result.ThumbnailDirectPath = vid.ThumbnailDirectPath
	}
	if vid.ThumbnailSHA256 != nil {
		result.ThumbnailSHA256 = vid.ThumbnailSHA256
	}
	if vid.ThumbnailEncSHA256 != nil {
		result.ThumbnailEncSHA256 = vid.ThumbnailEncSHA256
	}

	return result
}

// convertDocumentMessage converts waProto.DocumentMessage to waE2E.DocumentMessage
func (b *ProtoBuilder) convertDocumentMessage(doc *waProto.DocumentMessage) *waE2E.DocumentMessage {
	if doc == nil {
		return nil
	}

	result := &waE2E.DocumentMessage{
		URL:           doc.URL,
		DirectPath:    doc.DirectPath,
		MediaKey:      doc.MediaKey,
		Mimetype:      doc.Mimetype,
		FileEncSHA256: doc.FileEncSHA256,
		FileSHA256:    doc.FileSHA256,
		FileLength:    doc.FileLength,
		FileName:      doc.FileName,
		PageCount:     doc.PageCount,
		JPEGThumbnail: doc.JPEGThumbnail,
	}

	if doc.ThumbnailDirectPath != nil {
		result.ThumbnailDirectPath = doc.ThumbnailDirectPath
	}
	if doc.ThumbnailSHA256 != nil {
		result.ThumbnailSHA256 = doc.ThumbnailSHA256
	}
	if doc.ThumbnailEncSHA256 != nil {
		result.ThumbnailEncSHA256 = doc.ThumbnailEncSHA256
	}

	return result
}

// BuildCarouselMessageWithMedia creates CarouselMessage with multiple interactive cards
// using pre-processed media headers for each card
func (b *ProtoBuilder) BuildCarouselMessageWithMedia(req *SendCarouselRequest, cardMediaHeaders []*waProto.InteractiveMessage_Header) (*waE2E.Message, error) {
	if len(req.Cards) == 0 {
		return nil, fmt.Errorf("carousel requires at least one card")
	}
	if len(req.Cards) > 10 {
		return nil, fmt.Errorf("carousel supports maximum 10 cards")
	}

	cards := make([]*waE2E.InteractiveMessage, 0, len(req.Cards))

	for i, card := range req.Cards {
		// Get media header for this card (if available)
		var mediaHeader *waProto.InteractiveMessage_Header
		if i < len(cardMediaHeaders) && cardMediaHeaders[i] != nil {
			mediaHeader = cardMediaHeaders[i]
		}

		interactiveCard, err := b.buildCarouselCardWithMedia(&card, i, mediaHeader)
		if err != nil {
			return nil, fmt.Errorf("card %d: %w", i, err)
		}
		cards = append(cards, interactiveCard)
	}

	// Determine carousel card type
	cardType := waE2E.InteractiveMessage_CarouselMessage_HSCROLL_CARDS
	if req.CardType == CarouselCardTypeAlbum {
		cardType = waE2E.InteractiveMessage_CarouselMessage_ALBUM_IMAGE
	}

	carousel := &waE2E.InteractiveMessage_CarouselMessage{
		Cards:            cards,
		MessageVersion:   proto.Int32(1),
		CarouselCardType: &cardType,
	}

	// Wrap in InteractiveMessage for proper delivery
	// Include Body at the root InteractiveMessage level for the carousel body text
	interactiveMsg := &waE2E.InteractiveMessage{
		InteractiveMessage: &waE2E.InteractiveMessage_CarouselMessage_{
			CarouselMessage: carousel,
		},
	}

	// Add body text if provided (displayed above carousel cards)
	if req.Message != "" {
		interactiveMsg.Body = &waE2E.InteractiveMessage_Body{
			Text: proto.String(req.Message),
		}
	}

	msg := &waE2E.Message{
		InteractiveMessage: interactiveMsg,
	}

	return msg, nil
}

// buildCarouselCardWithMedia creates a single card for carousel message with pre-processed media
func (b *ProtoBuilder) buildCarouselCardWithMedia(card *CarouselCard, index int, mediaHeader *waProto.InteractiveMessage_Header) (*waE2E.InteractiveMessage, error) {
	// Build buttons using NativeFlowMessage
	buttons := make([]*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton, 0, len(card.Buttons))

	for _, btn := range card.Buttons {
		var paramsJSON []byte
		var buttonName string
		var err error

		normalizedType := btn.GetNormalizedType()

		switch normalizedType {
		case ButtonTypeQuickReply:
			buttonName = "quick_reply"
			paramsJSON, err = json.Marshal(map[string]interface{}{
				"id":           btn.ID,
				"display_text": btn.Label,
				"disabled":     false,
			})

		case ButtonTypeCTAURL:
			buttonName = "cta_url"
			if btn.URL == nil || *btn.URL == "" {
				return nil, fmt.Errorf("URL is required for cta_url button")
			}
			paramsJSON, err = json.Marshal(map[string]interface{}{
				"display_text": btn.Label,
				"url":          *btn.URL,
				"disabled":     false,
			})

		case ButtonTypeCTACall:
			buttonName = "cta_call"
			if btn.Phone == nil || *btn.Phone == "" {
				return nil, fmt.Errorf("phone is required for cta_call button")
			}
			paramsJSON, err = json.Marshal(map[string]interface{}{
				"display_text": btn.Label,
				"phone_number": *btn.Phone,
				"disabled":     false,
			})

		case ButtonTypeCTACopy:
			buttonName = "cta_copy"
			if btn.CopyCode == nil || *btn.CopyCode == "" {
				return nil, fmt.Errorf("copyCode is required for cta_copy button")
			}
			paramsJSON, err = json.Marshal(map[string]interface{}{
				"display_text": btn.Label,
				"copy_code":    *btn.CopyCode,
				"disabled":     false,
			})

		default:
			return nil, fmt.Errorf("unsupported button type for carousel: %s", btn.Type)
		}

		if err != nil {
			return nil, fmt.Errorf("marshal button params: %w", err)
		}

		buttons = append(buttons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
			Name:             proto.String(buttonName),
			ButtonParamsJSON: proto.String(string(paramsJSON)),
		})
	}

	// Build body
	body := &waE2E.InteractiveMessage_Body{
		Text: proto.String(card.Body.Text),
	}

	// Build header with media if available
	var header *waE2E.InteractiveMessage_Header
	if mediaHeader != nil {
		var title *string
		if card.Header != nil && card.Header.Text != "" {
			title = &card.Header.Text
		}
		header = b.convertMediaHeader(mediaHeader, title)
	} else if card.Header != nil && card.Header.Text != "" {
		header = &waE2E.InteractiveMessage_Header{
			Title:              proto.String(card.Header.Text),
			HasMediaAttachment: proto.Bool(false),
		}
	}

	// Build optional footer
	var footer *waE2E.InteractiveMessage_Footer
	if card.Footer != nil && card.Footer.Text != "" {
		footer = &waE2E.InteractiveMessage_Footer{
			Text: proto.String(card.Footer.Text),
		}
	}

	// Build native flow message
	msgVersion := int32(1)
	nativeFlow := &waE2E.InteractiveMessage_NativeFlowMessage{
		MessageVersion: &msgVersion,
		Buttons:        buttons,
	}

	interactiveCard := &waE2E.InteractiveMessage{
		Header: header,
		Body:   body,
		Footer: footer,
		InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
			NativeFlowMessage: nativeFlow,
		},
	}

	return interactiveCard, nil
}
