package communities

import "errors"

var (
	// ErrInvalidPagination indicates the request did not provide valid pagination parameters.
	ErrInvalidPagination = errors.New("invalid pagination parameters")

	// ErrClientNotConnected signals that the WhatsApp client is not currently connected.
	ErrClientNotConnected = errors.New("whatsapp client not connected")

	// ErrInvalidCommunityID denotes an invalid or malformed community identifier.
	ErrInvalidCommunityID = errors.New("invalid community id")

	// ErrInvalidGroupList indicates that the request did not provide any valid group identifiers.
	ErrInvalidGroupList = errors.New("invalid groups list")

	// ErrInvalidPhoneList indicates that the request did not provide any participant phone numbers.
	ErrInvalidPhoneList = errors.New("invalid phone list")

	// ErrOperationFailed is returned when WhatsApp rejects the requested operation.
	ErrOperationFailed = errors.New("community operation failed")

	// ErrInvalidCommunityName is returned when the provided community name is empty or exceeds WhatsApp limits.
	ErrInvalidCommunityName = errors.New("invalid community name")

	// ErrInvalidCommunityDescription denotes an empty or malformed community description.
	ErrInvalidCommunityDescription = errors.New("invalid community description")
)
