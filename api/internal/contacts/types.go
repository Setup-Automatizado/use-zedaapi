package contacts

// ListParams describes the pagination configuration for listing contacts.
type ListParams struct {
	Page     int
	PageSize int
}

// Contact represents the payload for a single contact.
// Maps whatsmeow's types.ContactInfo to Zé da API expected format.
type Contact struct {
	Phone  string  `json:"phone"`  // Phone number
	Name   *string `json:"name"`   // Full name (only if in contacts)
	Short  *string `json:"short"`  // First name (only if in contacts)
	Notify *string `json:"notify"` // Name from WhatsApp settings (push name)
	Vname  *string `json:"vname"`  // Contact name from vcard (full name)
}

// ListResult contains the paginated result of the List operation.
type ListResult struct {
	Items []Contact
	Total int
}

// PhoneExistsResponse represents the response for phone validation.
// Returns whether a phone number is registered on WhatsApp along with phone and LID.
type PhoneExistsResponse struct {
	Exists bool    `json:"exists"` // Whether the phone is registered on WhatsApp
	Phone  *string `json:"phone"`  // Phone number if exists, null otherwise
	LID    *string `json:"lid"`    // LID (Linked ID) if available, null otherwise
}

// PhoneExistsBatchRequest represents the request for batch phone validation.
type PhoneExistsBatchRequest struct {
	Phones []string `json:"phones"`
}

// PhoneExistsBatchResponse represents the response for batch phone validation.
// Contains the validation result for each phone number in the batch.
type PhoneExistsBatchResponse struct {
	Exists      bool   `json:"exists"`      // true if the phone has WhatsApp
	InputPhone  string `json:"inputPhone"`  // Phone number as sent in the request
	OutputPhone string `json:"outputPhone"` // Formatted phone number from WhatsApp response
}
