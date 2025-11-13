package newsletters

import "errors"

var (
	// ErrInvalidPagination indicates the request did not pass valid pagination parameters.
	ErrInvalidPagination = errors.New("invalid pagination parameters")

	// ErrClientNotConnected signals that the WhatsApp client is unavailable.
	ErrClientNotConnected = errors.New("whatsapp client not connected")

	// ErrInvalidNewsletterID indicates a malformed or empty newsletter identifier.
	ErrInvalidNewsletterID = errors.New("invalid newsletter id")

	// ErrInvalidName denotes an empty or otherwise invalid newsletter name.
	ErrInvalidName = errors.New("invalid newsletter name")

	// ErrInvalidPicture means the picture payload could not be processed.
	ErrInvalidPicture = errors.New("invalid newsletter picture")

	// ErrInvalidReactionCodes represents an unsupported reaction setting value.
	ErrInvalidReactionCodes = errors.New("invalid reaction codes")

	// ErrInvalidPhone indicates the provided phone value is invalid.
	ErrInvalidPhone = errors.New("invalid phone")
)
