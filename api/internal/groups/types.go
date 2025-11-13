package groups

// ListParams describes the pagination configuration for listing groups.
type ListParams struct {
	Page     int
	PageSize int
}

// Summary represents the Z-API compatible payload for a single group item.
type Summary struct {
	IsGroup         bool    `json:"isGroup"`
	Name            string  `json:"name"`
	Phone           string  `json:"phone"`
	Unread          string  `json:"unread"`
	LastMessageTime *string `json:"lastMessageTime"`
	IsMuted         string  `json:"isMuted"`
	MuteEndTime     *string `json:"muteEndTime"`
	IsMarkedSpam    bool    `json:"isMarkedSpam"`
	Archived        bool    `json:"archived"`
	Pinned          bool    `json:"pinned"`
	MessagesUnread  string  `json:"messagesUnread"`
}

// ListResult contains the paginated result of the List operation.
type ListResult struct {
	Items []Summary
	Total int
}

// CreateParams captures the body fields required to create a group.
type CreateParams struct {
	GroupName  string
	Phones     []string
	AutoInvite bool
}

// CreateResult returns the identifiers of the newly created group.
type CreateResult struct {
	Phone          string `json:"phone"`
	InvitationLink string `json:"invitationLink"`
}

// UpdateNameParams represents the payload to rename a group.
type UpdateNameParams struct {
	GroupID   string
	GroupName string
}

// UpdatePhotoParams encapsulates the information required to update a group picture.
type UpdatePhotoParams struct {
	GroupID    string
	GroupPhoto string
}

// ModifyParticipantsParams represents participant additions or removals.
type ModifyParticipantsParams struct {
	GroupID    string
	Phones     []string
	AutoInvite bool
}

// UpdateDescriptionParams changes the group description.
type UpdateDescriptionParams struct {
	GroupID          string
	GroupDescription string
}

// UpdateSettingsParams toggles the administrative settings of a group.
type UpdateSettingsParams struct {
	Phone                string
	AdminOnlyMessage     bool
	AdminOnlySettings    bool
	RequireAdminApproval bool
	AdminOnlyAddMember   bool
}

// InvitationLinkResult represents the payload returned when fetching or redefining an invite link.
type InvitationLinkResult struct {
	Phone          string `json:"phone"`
	InvitationLink string `json:"invitationLink"`
}

// ValueResult mirrors the generic Z-API boolean response.
type ValueResult struct {
	Value bool `json:"value"`
}

// AcceptInviteParams wraps the invite URL for acceptance.
type AcceptInviteParams struct {
	URL string
}

// AcceptInviteResult conveys whether the invite acceptance succeeded.
type AcceptInviteResult struct {
	Success bool `json:"success"`
}

// InvitationMetadata summarises group details resolved from an invite link.
type InvitationMetadata struct {
	Phone             string               `json:"phone"`
	Owner             string               `json:"owner"`
	Subject           string               `json:"subject"`
	Description       string               `json:"description"`
	Creation          int64                `json:"creation"`
	InvitationLink    string               `json:"invitationLink"`
	ContactsCount     int                  `json:"contactsCount"`
	ParticipantsCount int                  `json:"participantsCount"`
	Participants      []ParticipantSummary `json:"participants"`
	SubjectTime       int64                `json:"subjectTime"`
	SubjectOwner      string               `json:"subjectOwner"`
}

// Metadata captures the full group metadata including invite link, while LightMetadata omits it.
type Metadata struct {
	Phone                string               `json:"phone"`
	Description          string               `json:"description"`
	Owner                string               `json:"owner"`
	Subject              string               `json:"subject"`
	Creation             int64                `json:"creation"`
	InvitationLink       *string              `json:"invitationLink"`
	InvitationLinkError  *string              `json:"invitationLinkError"`
	CommunityID          *string              `json:"communityId"`
	AdminOnlyMessage     bool                 `json:"adminOnlyMessage"`
	AdminOnlySettings    bool                 `json:"adminOnlySettings"`
	RequireAdminApproval bool                 `json:"requireAdminApproval"`
	IsGroupAnnouncement  bool                 `json:"isGroupAnnouncement"`
	Participants         []ParticipantSummary `json:"participants"`
	SubjectTime          int64                `json:"subjectTime"`
	SubjectOwner         string               `json:"subjectOwner"`
}

// ParticipantSummary adapts WhatsApp participant data to the Z-API schema.
type ParticipantSummary struct {
	Phone        string  `json:"phone"`
	LID          string  `json:"lid,omitempty"`
	IsAdmin      bool    `json:"isAdmin"`
	IsSuperAdmin bool    `json:"isSuperAdmin"`
	Short        *string `json:"short,omitempty"`
	Name         *string `json:"name,omitempty"`
}
