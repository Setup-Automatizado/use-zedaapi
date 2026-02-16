package communities

// ListParams defines the pagination parameters for listing communities.
type ListParams struct {
	Page     int
	PageSize int
}

// Summary captures the minimal payload for a community entry.
type Summary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListResult represents the outcome of a paginated list operation.
type ListResult struct {
	Items []Summary
	Total int
}

// CreateParams holds the values required to create a new community.
type CreateParams struct {
	Name        string
	Description string
}

// SubGroup represents a linked group inside a community.
type SubGroup struct {
	Name                string `json:"name"`
	Phone               string `json:"phone"`
	IsGroupAnnouncement bool   `json:"isGroupAnnouncement"`
}

// CreateResult is returned after a community is created successfully.
type CreateResult struct {
	ID                  string     `json:"id"`
	InvitationLink      *string    `json:"invitationLink,omitempty"`
	AnnouncementGroupID string     `json:"announcementGroupId,omitempty"`
	SubGroups           []SubGroup `json:"subGroups"`
}

// LinkParams models the request to link or unlink groups to a community.
type LinkParams struct {
	CommunityID string
	GroupIDs    []string
}

// Metadata captures the detailed information of a community.
type Metadata struct {
	ID                  string     `json:"id"`
	Name                string     `json:"name"`
	Description         *string    `json:"description,omitempty"`
	AnnouncementGroupID string     `json:"announcementGroupId,omitempty"`
	ParticipantsCount   int        `json:"participantsCount"`
	SubGroups           []SubGroup `json:"subGroups"`
}

// InvitationResult wraps an invitation link response.
type InvitationResult struct {
	InvitationLink string `json:"invitationLink"`
}

// OperationResult adapts boolean acknowledgements to either `value` or `success`.
type OperationResult struct {
	Value   *bool `json:"value,omitempty"`
	Success *bool `json:"success,omitempty"`
}

// NewValueResult builds an OperationResult with the Zé da API `value` field.
func NewValueResult(ok bool) OperationResult {
	return OperationResult{Value: boolPtr(ok)}
}

// NewSuccessResult builds an OperationResult with the Zé da API `success` field.
func NewSuccessResult(ok bool) OperationResult {
	return OperationResult{Success: boolPtr(ok)}
}

// ValueBool returns the value flag and whether it is present.
func (r OperationResult) ValueBool() (bool, bool) {
	if r.Value == nil {
		return false, false
	}
	return *r.Value, true
}

// SuccessBool returns the success flag and whether it is present.
func (r OperationResult) SuccessBool() (bool, bool) {
	if r.Success == nil {
		return false, false
	}
	return *r.Success, true
}

// SettingsParams defines the payload accepted by the communities settings endpoint.
type SettingsParams struct {
	CommunityID        string
	WhoCanAddNewGroups string
}

// ParticipantsParams encapsulates participant mutations for communities.
type ParticipantsParams struct {
	CommunityID string
	Phones      []string
	AutoInvite  bool
}

// UpdateDescriptionParams captures community description updates.
type UpdateDescriptionParams struct {
	CommunityID string
	Description string
}

func boolPtr(v bool) *bool {
	return &v
}
