package newsletters

// ListParams defines pagination options for listing newsletters.
type ListParams struct {
	Page     int
	PageSize int
}

// CreateParams holds the information required to create a newsletter.
type CreateParams struct {
	Name        string
	Description *string
	Picture     string
}

// CreateResult represents the response returned after creating a channel.
type CreateResult struct {
	ID string `json:"id"`
}

// UpdatePictureParams describes the payload for updating a channel picture.
type UpdatePictureParams struct {
	ID      string
	Picture string
}

// UpdateNameParams contains the fields to rename a channel.
type UpdateNameParams struct {
	ID   string
	Name string
}

// UpdateDescriptionParams updates the textual description of a channel.
type UpdateDescriptionParams struct {
	ID          string
	Description string
}

// SettingsParams carries configuration changes for a channel.
type SettingsParams struct {
	ID            string
	ReactionCodes string
}

// AdminActionParams encapsulates newsletter administrator mutations.
type AdminActionParams struct {
	ID    string
	Phone string
}

// TransferOwnershipParams represents the body of the ownership transfer operation.
type TransferOwnershipParams struct {
	ID        string
	Phone     string
	QuitAdmin bool
}

// IDParams is used by operations that only require a channel identifier.
type IDParams struct {
	ID string
}

// Summary represents the newsletter payload.
type Summary struct {
	ID               string        `json:"id"`
	CreationTime     string        `json:"creationTime"`
	State            string        `json:"state"`
	Name             string        `json:"name"`
	Description      string        `json:"description"`
	SubscribersCount string        `json:"subscribersCount"`
	InviteLink       string        `json:"inviteLink"`
	Verification     string        `json:"verification"`
	Picture          *string       `json:"picture"`
	Preview          *string       `json:"preview"`
	ViewMetadata     *ViewMetadata `json:"viewMetadata,omitempty"`
}

// ViewMetadata mirrors the nested object returned by Zé da API for newsletters.
type ViewMetadata struct {
	Mute string `json:"mute"`
	Role string `json:"role"`
}

// ListResult contains the paginated outcome of the List operation.
type ListResult struct {
	Items []Summary
	Total int
}

// MetadataResult contains the metadata for a single newsletter.
type MetadataResult Summary

// SearchParams aggregates filters for the search endpoint.
type SearchParams struct {
	Limit        int
	View         string
	CountryCodes []string
	SearchText   *string
}

// SearchItem represents a single entry returned by the search endpoint.
type SearchItem struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Description      *string `json:"description,omitempty"`
	SubscribersCount string  `json:"subscribersCount"`
	Picture          *string `json:"picture,omitempty"`
}

// SearchResult wraps search outcomes together with an optional cursor.
type SearchResult struct {
	Cursor *string      `json:"cursor,omitempty"`
	Data   []SearchItem `json:"data"`
}

// OperationResult mirrors the boolean success envelope used across Zé da API mutations.
type OperationResult struct {
	Value   bool    `json:"value"`
	Message *string `json:"message,omitempty"`
}
