package groups

import "errors"

var (
	// ErrInvalidPagination is returned when page or page size are not positive numbers.
	ErrInvalidPagination = errors.New("invalid pagination parameters")

	// ErrClientNotConnected indicates the WhatsApp client is not currently available for the instance.
	ErrClientNotConnected = errors.New("whatsapp client not connected")

	// ErrInvalidGroupID denotes an invalid or malformed group identifier.
	ErrInvalidGroupID = errors.New("invalid group id")

	// ErrInvalidPhoneList indicates that the request did not provide any participant phone numbers.
	ErrInvalidPhoneList = errors.New("invalid phone list")

	// ErrInvalidGroupName is returned when the provided group name is empty or exceeds WhatsApp limits.
	ErrInvalidGroupName = errors.New("invalid group name")

	// ErrInvalidInviteURL represents an invalid or malformed invite URL.
	ErrInvalidInviteURL = errors.New("invalid invite url")

	// ErrOperationFailed is returned when WhatsApp rejects the requested operation.
	ErrOperationFailed = errors.New("group operation failed")
)
