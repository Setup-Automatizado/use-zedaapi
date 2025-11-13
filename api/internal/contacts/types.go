package contacts

// ListParams describes the pagination configuration for listing contacts.
type ListParams struct {
	Page     int
	PageSize int
}

// Contact represents the Z-API compatible payload for a single contact.
// Maps whatsmeow's types.ContactInfo to Z-API expected format.
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
