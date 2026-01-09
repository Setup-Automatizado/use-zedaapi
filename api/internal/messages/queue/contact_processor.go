package queue

import (
	"context"
	"fmt"
	"log/slog"

	wameow "go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

// ContactProcessor handles contact message sending via WhatsApp
type ContactProcessor struct {
	log            *slog.Logger
	presenceHelper *PresenceHelper
}

// NewContactProcessor creates a new contact message processor
func NewContactProcessor(log *slog.Logger) *ContactProcessor {
	return &ContactProcessor{
		log:            log.With(slog.String("processor", "contact")),
		presenceHelper: NewPresenceHelper(),
	}
}

// Process sends a contact message via WhatsApp
func (p *ContactProcessor) Process(ctx context.Context, client *wameow.Client, args *SendMessageArgs) error {
	if args.ContactContent == nil {
		return fmt.Errorf("contact_content is required for contact messages")
	}

	// Parse recipient JID
	recipientJID, err := types.ParseJID(args.Phone)
	if err != nil {
		return fmt.Errorf("invalid phone number: %w", err)
	}

	// Simulate typing indicator if DelayTyping is set
	if args.DelayTyping > 0 {
		if err := p.presenceHelper.SimulateTyping(client, recipientJID, args.DelayTyping); err != nil {
			p.log.Warn("failed to send typing indicator",
				slog.String("error", err.Error()),
				slog.String("phone", args.Phone))
		}
	}

	// Build vCard (contact card)
	// If VCard is provided directly (Z-API format), use it
	// Otherwise, build from individual fields
	var vcard string
	var displayName string

	if args.ContactContent.VCard != nil && *args.ContactContent.VCard != "" {
		// Use pre-formatted vCard
		vcard = *args.ContactContent.VCard
		// Extract display name from vCard (FN: field)
		displayName = p.extractDisplayNameFromVCard(vcard)
	} else {
		// Build vCard from individual fields
		vcard = p.buildVCard(args.ContactContent)
		displayName = args.ContactContent.FullName
	}

	// Build ContextInfo using helper
	// This provides support for: mentions, reply-to, ephemeral messages, private answer
	contextBuilder := NewContextInfoBuilder(client, recipientJID, args, p.log)
	contextInfo, err := contextBuilder.Build(ctx)
	if err != nil {
		p.log.Warn("failed to build context info, sending without it",
			slog.String("error", err.Error()),
			slog.String("phone", args.Phone))
	}

	// Build contact message
	msg := &waProto.Message{
		ContactMessage: &waProto.ContactMessage{
			DisplayName: &displayName,
			Vcard:       &vcard,
			ContextInfo: contextInfo, // Can be nil if no context features
		},
	}

	// Send message
	resp, err := client.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return fmt.Errorf("send contact message: %w", err)
	}

	p.log.Info("contact message sent successfully",
		slog.String("zaap_id", args.ZaapID),
		slog.String("phone", args.Phone),
		slog.String("whatsapp_message_id", resp.ID),
		slog.String("contact_name", args.ContactContent.FullName),
		slog.String("contact_phone", args.ContactContent.PhoneNumber),
		// Professional fields
		slog.Bool("has_organization", args.ContactContent.Organization != nil && *args.ContactContent.Organization != ""),
		slog.Bool("has_job_title", args.ContactContent.JobTitle != nil && *args.ContactContent.JobTitle != ""),
		slog.Bool("has_email", args.ContactContent.Email != nil && *args.ContactContent.Email != ""),
		slog.Bool("has_url", args.ContactContent.URL != nil && *args.ContactContent.URL != ""),
		// Name fields
		slog.Bool("has_structured_name", args.ContactContent.FirstName != nil || args.ContactContent.LastName != nil),
		slog.Bool("has_nickname", args.ContactContent.Nickname != nil && *args.ContactContent.Nickname != ""),
		// Address field
		slog.Bool("has_address", args.ContactContent.Address != nil),
		// Personal fields
		slog.Bool("has_birthday", args.ContactContent.Birthday != nil && *args.ContactContent.Birthday != ""),
		slog.Bool("has_note", args.ContactContent.Note != nil && *args.ContactContent.Note != ""),
		// Context info
		slog.Bool("has_context", contextInfo != nil),
		slog.Time("timestamp", resp.Timestamp))

	args.WhatsAppMessageID = resp.ID

	return nil
}

// buildVCard builds a complete vCard string from contact information
// Follows RFC 6350 (vCard 3.0) with WhatsApp-specific extensions
// Supports all standard vCard fields: FN, N, NICKNAME, TEL, EMAIL, ORG, TITLE, URL, ADR, BDAY, NOTE
func (p *ContactProcessor) buildVCard(contact *ContactMessage) string {
	vcard := "BEGIN:VCARD\n"
	vcard += "VERSION:3.0\n"

	// FN: Full Name (required)
	vcard += fmt.Sprintf("FN:%s\n", escapeVCardValue(contact.FullName))

	// N: Structured Name (LastName;FirstName;MiddleName;Prefix;Suffix)
	// If any name component is provided, build the N: field
	if contact.LastName != nil || contact.FirstName != nil || contact.MiddleName != nil ||
		contact.NamePrefix != nil || contact.NameSuffix != nil {
		lastName := safeDeref(contact.LastName)
		firstName := safeDeref(contact.FirstName)
		middleName := safeDeref(contact.MiddleName)
		prefix := safeDeref(contact.NamePrefix)
		suffix := safeDeref(contact.NameSuffix)
		vcard += fmt.Sprintf("N:%s;%s;%s;%s;%s\n",
			escapeVCardValue(lastName),
			escapeVCardValue(firstName),
			escapeVCardValue(middleName),
			escapeVCardValue(prefix),
			escapeVCardValue(suffix))
	}

	// NICKNAME: Nickname or alias
	if contact.Nickname != nil && *contact.Nickname != "" {
		vcard += fmt.Sprintf("NICKNAME:%s\n", escapeVCardValue(*contact.Nickname))
	}

	// TEL: Phone number (required, with WhatsApp WAID format)
	vcard += fmt.Sprintf("TEL;type=CELL;waid=%s:+%s\n",
		contact.PhoneNumber, contact.PhoneNumber)

	// EMAIL: Email address
	if contact.Email != nil && *contact.Email != "" {
		vcard += fmt.Sprintf("EMAIL:%s\n", escapeVCardValue(*contact.Email))
	}

	// ORG: Organization/company name
	if contact.Organization != nil && *contact.Organization != "" {
		vcard += fmt.Sprintf("ORG:%s\n", escapeVCardValue(*contact.Organization))
	}

	// TITLE: Job title or position
	if contact.JobTitle != nil && *contact.JobTitle != "" {
		vcard += fmt.Sprintf("TITLE:%s\n", escapeVCardValue(*contact.JobTitle))
	}

	// URL: Website or social media URL
	if contact.URL != nil && *contact.URL != "" {
		vcard += fmt.Sprintf("URL:%s\n", escapeVCardValue(*contact.URL))
	}

	// ADR: Structured address (PostBox;Extended;Street;City;Region;PostalCode;Country)
	if contact.Address != nil {
		addrType := "HOME"
		if contact.Address.Type != nil && *contact.Address.Type != "" {
			addrType = *contact.Address.Type
		}
		vcard += fmt.Sprintf("ADR;TYPE=%s:%s;%s;%s;%s;%s;%s;%s\n",
			addrType,
			escapeVCardValue(safeDeref(contact.Address.PostBox)),
			escapeVCardValue(safeDeref(contact.Address.Extended)),
			escapeVCardValue(safeDeref(contact.Address.Street)),
			escapeVCardValue(safeDeref(contact.Address.City)),
			escapeVCardValue(safeDeref(contact.Address.Region)),
			escapeVCardValue(safeDeref(contact.Address.PostalCode)),
			escapeVCardValue(safeDeref(contact.Address.Country)))
	}

	// BDAY: Birthday (format: YYYY-MM-DD)
	if contact.Birthday != nil && *contact.Birthday != "" {
		vcard += fmt.Sprintf("BDAY:%s\n", *contact.Birthday)
	}

	// NOTE: Additional notes or comments
	if contact.Note != nil && *contact.Note != "" {
		vcard += fmt.Sprintf("NOTE:%s\n", escapeVCardValue(*contact.Note))
	}

	vcard += "END:VCARD"

	return vcard
}

// safeDeref safely dereferences a string pointer, returning empty string if nil
func safeDeref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// escapeVCardValue escapes special characters in vCard values
// According to RFC 6350, the following characters should be escaped:
// - Backslash (\) -> \\
// - Comma (,) -> \,
// - Semicolon (;) -> \;
// - Newline (\n) -> \n
func escapeVCardValue(value string) string {
	if value == "" {
		return ""
	}

	// Replace in order: \ -> \\, then special chars
	value = replaceAll(value, "\\", "\\\\")
	value = replaceAll(value, ";", "\\;")
	value = replaceAll(value, ",", "\\,")
	value = replaceAll(value, "\n", "\\n")
	value = replaceAll(value, "\r", "")

	return value
}

// replaceAll replaces all occurrences of old with new in string s
func replaceAll(s, old, new string) string {
	result := ""
	for len(s) > 0 {
		idx := indexString(s, old)
		if idx == -1 {
			result += s
			break
		}
		result += s[:idx] + new
		s = s[idx+len(old):]
	}
	return result
}

// indexString finds the index of substring in s, returns -1 if not found
func indexString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// extractDisplayNameFromVCard extracts the display name from a vCard string
// Tries FN: field first, falls back to N: field if not found
// Handles case-insensitive matching and escaped values
func (p *ContactProcessor) extractDisplayNameFromVCard(vcard string) string {
	lines := splitVCardLines(vcard)

	// First, try to find FN: field (Full Name)
	for _, line := range lines {
		line = trimSpace(line)
		if len(line) > 3 {
			// Case-insensitive match for FN:
			if line[:3] == "FN:" || line[:3] == "fn:" {
				name := line[3:]
				// Unescape vCard value
				name = unescapeVCardValue(name)
				return trimSpace(name)
			}
		}
	}

	// Fallback: try to extract from N: field (Structured Name)
	// Format: N:LastName;FirstName;MiddleName;Prefix;Suffix
	for _, line := range lines {
		line = trimSpace(line)
		if len(line) > 2 {
			if line[:2] == "N:" || line[:2] == "n:" {
				parts := splitVCardField(line[2:])
				// Build name from available parts: Prefix FirstName MiddleName LastName Suffix
				var nameParts []string
				if len(parts) > 3 && parts[3] != "" {
					nameParts = append(nameParts, parts[3]) // Prefix
				}
				if len(parts) > 1 && parts[1] != "" {
					nameParts = append(nameParts, parts[1]) // FirstName
				}
				if len(parts) > 2 && parts[2] != "" {
					nameParts = append(nameParts, parts[2]) // MiddleName
				}
				if len(parts) > 0 && parts[0] != "" {
					nameParts = append(nameParts, parts[0]) // LastName
				}
				if len(parts) > 4 && parts[4] != "" {
					nameParts = append(nameParts, parts[4]) // Suffix
				}

				if len(nameParts) > 0 {
					return joinStrings(nameParts, " ")
				}
			}
		}
	}

	// If no name found, return empty string
	return ""
}

// unescapeVCardValue unescapes special characters in vCard values
// Reverses escaping done by escapeVCardValue
func unescapeVCardValue(value string) string {
	if value == "" {
		return ""
	}

	value = replaceAll(value, "\\n", "\n")
	value = replaceAll(value, "\\,", ",")
	value = replaceAll(value, "\\;", ";")
	value = replaceAll(value, "\\\\", "\\")

	return value
}

// splitVCardField splits a vCard field value by semicolon, handling escaped semicolons
func splitVCardField(value string) []string {
	var parts []string
	current := ""
	escaped := false

	for i := 0; i < len(value); i++ {
		if escaped {
			current += string(value[i])
			escaped = false
		} else if value[i] == '\\' {
			current += string(value[i])
			escaped = true
		} else if value[i] == ';' {
			parts = append(parts, unescapeVCardValue(current))
			current = ""
		} else {
			current += string(value[i])
		}
	}

	// Add last part
	if current != "" || len(parts) > 0 {
		parts = append(parts, unescapeVCardValue(current))
	}

	return parts
}

// trimSpace removes leading and trailing whitespace from string
func trimSpace(s string) string {
	start := 0
	end := len(s)

	// Trim leading spaces
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r' || s[start] == '\n') {
		start++
	}

	// Trim trailing spaces
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r' || s[end-1] == '\n') {
		end--
	}

	return s[start:end]
}

// joinStrings joins string slice with separator
func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}

	return result
}

// splitVCardLines splits vCard content by newlines, handling both \n and \r\n
func splitVCardLines(vcard string) []string {
	var lines []string
	currentLine := ""

	for i := 0; i < len(vcard); i++ {
		if vcard[i] == '\r' && i+1 < len(vcard) && vcard[i+1] == '\n' {
			// Windows line ending \r\n
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = ""
			}
			i++ // Skip the \n
		} else if vcard[i] == '\n' {
			// Unix line ending \n
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = ""
			}
		} else {
			currentLine += string(vcard[i])
		}
	}

	// Add last line if any
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// ProcessMultiple sends multiple contacts in a single ContactsArrayMessage via WhatsApp
// This is the correct way to send multiple contacts together (not as separate messages)
// The WhatsApp UI will show them as "Contact Name e outros X contatos" with a "Ver todos" button
func (p *ContactProcessor) ProcessMultiple(ctx context.Context, client *wameow.Client, args *SendMessageArgs, contacts []*ContactMessage) error {
	if len(contacts) == 0 {
		return fmt.Errorf("at least one contact is required for contacts array message")
	}

	// Parse recipient JID
	recipientJID, err := types.ParseJID(args.Phone)
	if err != nil {
		return fmt.Errorf("invalid phone number: %w", err)
	}

	// Simulate typing indicator if DelayTyping is set
	if args.DelayTyping > 0 {
		if err := p.presenceHelper.SimulateTyping(client, recipientJID, args.DelayTyping); err != nil {
			p.log.Warn("failed to send typing indicator",
				slog.String("error", err.Error()),
				slog.String("phone", args.Phone))
		}
	}

	// Build ContactMessage array for whatsmeow proto
	var contactMessages []*waProto.ContactMessage
	for _, contact := range contacts {
		// Build vCard for each contact
		var vcard string
		var displayName string

		if contact.VCard != nil && *contact.VCard != "" {
			// Use pre-formatted vCard
			vcard = *contact.VCard
			displayName = p.extractDisplayNameFromVCard(vcard)
		} else {
			// Build vCard from individual fields
			vcard = p.buildVCard(contact)
			displayName = contact.FullName
		}

		// Create ContactMessage for this contact
		contactMessages = append(contactMessages, &waProto.ContactMessage{
			DisplayName: &displayName,
			Vcard:       &vcard,
		})
	}

	// Build display name for the contacts array
	// Format: "FirstContactName e outros X contatos" (like WhatsApp does)
	arrayDisplayName := contacts[0].FullName
	if len(contacts) > 1 {
		arrayDisplayName = fmt.Sprintf("%s e outros %d contatos", contacts[0].FullName, len(contacts)-1)
	}

	// Build ContextInfo using helper
	// This provides support for: mentions, reply-to, ephemeral messages, private answer
	contextBuilder := NewContextInfoBuilder(client, recipientJID, args, p.log)
	contextInfo, err := contextBuilder.Build(ctx)
	if err != nil {
		p.log.Warn("failed to build context info, sending without it",
			slog.String("error", err.Error()),
			slog.String("phone", args.Phone))
	}

	// Build ContactsArrayMessage (for multiple contacts in one message)
	msg := &waProto.Message{
		ContactsArrayMessage: &waProto.ContactsArrayMessage{
			DisplayName: &arrayDisplayName,
			Contacts:    contactMessages,
			ContextInfo: contextInfo, // Can be nil if no context features
		},
	}

	// Send message
	resp, err := client.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return fmt.Errorf("send contacts array message: %w", err)
	}

	p.log.Info("contacts array message sent successfully",
		slog.String("zaap_id", args.ZaapID),
		slog.String("phone", args.Phone),
		slog.String("whatsapp_message_id", resp.ID),
		slog.Int("contact_count", len(contacts)),
		slog.String("display_name", arrayDisplayName),
		slog.Bool("has_context", contextInfo != nil),
		slog.Time("timestamp", resp.Timestamp))

	args.WhatsAppMessageID = resp.ID

	return nil
}
