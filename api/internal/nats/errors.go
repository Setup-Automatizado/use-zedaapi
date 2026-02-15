package nats

import "errors"

// Sentinel errors for NATS operations.
var (
	ErrNotConnected    = errors.New("nats: not connected")
	ErrStreamNotFound  = errors.New("nats: stream not found")
	ErrPublishFailed   = errors.New("nats: publish failed")
	ErrPublishTimeout  = errors.New("nats: publish timeout")
	ErrConsumerFailed  = errors.New("nats: consumer creation failed")
	ErrDrainTimeout    = errors.New("nats: drain timeout")
	ErrInvalidConfig   = errors.New("nats: invalid configuration")
	ErrStreamExists    = errors.New("nats: stream already exists")
	ErrConsumerExists  = errors.New("nats: consumer already exists")
	ErrMessageTooLarge = errors.New("nats: message exceeds max payload size")
	ErrNoResponders    = errors.New("nats: no responders for subject")
	ErrAckFailed       = errors.New("nats: acknowledgment failed")
)
