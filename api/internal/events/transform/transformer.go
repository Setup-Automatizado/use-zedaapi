package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"go.mau.fi/whatsmeow/api/internal/events/types"
)

// SourceTransformer converts raw events from a source library (e.g., whatsmeow)
// to the unified InternalEvent format.
//
// Implementations must be thread-safe as they may be called concurrently.
type SourceTransformer interface {
	// Transform converts a raw event to InternalEvent format.
	// The rawEvent parameter should be a whatsmeow event (e.g., *events.Message).
	//
	// Returns:
	// - *types.InternalEvent: The transformed event ready for persistence
	// - error: Transformation error or ErrUnsupportedEvent if event type not supported
	Transform(ctx context.Context, rawEvent interface{}) (*types.InternalEvent, error)

	// SupportsEvent returns true if this transformer can handle the given event type.
	// This allows the transformation pipeline to route events to the correct transformer.
	SupportsEvent(eventType reflect.Type) bool

	// SourceLib returns the source library identifier (e.g., "whatsmeow", "cloud_api")
	SourceLib() types.SourceLib
}

// TargetTransformer converts unified InternalEvent to a specific target format (e.g., Z-API webhook).
//
// Implementations must be thread-safe as they may be called concurrently.
type TargetTransformer interface {
	// Transform converts an InternalEvent to the target webhook format.
	//
	// Returns:
	// - json.RawMessage: The serialized webhook payload ready for delivery
	// - error: Transformation error or ErrUnsupportedEvent if event type not supported
	Transform(ctx context.Context, event *types.InternalEvent) (json.RawMessage, error)

	// TargetSchema returns the target schema identifier (e.g., "zapi", "webhook", "rabbitmq")
	TargetSchema() string

	// SupportsEventType returns true if this transformer can handle the given event type.
	SupportsEventType(eventType string) bool
}

// Pipeline chains a SourceTransformer and TargetTransformer to convert
// raw events to target webhook payloads in a single operation.
type Pipeline struct {
	source SourceTransformer
	target TargetTransformer
}

// NewPipeline creates a transformation pipeline with the given source and target transformers.
func NewPipeline(source SourceTransformer, target TargetTransformer) *Pipeline {
	return &Pipeline{
		source: source,
		target: target,
	}
}

// Transform executes the full transformation pipeline:
// rawEvent → InternalEvent → target format (e.g., Z-API webhook)
//
// Returns:
// - json.RawMessage: The final webhook payload ready for delivery
// - *types.InternalEvent: The intermediate unified event (useful for logging/debugging)
// - error: Any transformation error
func (p *Pipeline) Transform(ctx context.Context, rawEvent interface{}) (json.RawMessage, *types.InternalEvent, error) {
	// Step 1: Source transformation (rawEvent → InternalEvent)
	internalEvent, err := p.source.Transform(ctx, rawEvent)
	if err != nil {
		return nil, nil, fmt.Errorf("source transformation failed: %w", err)
	}

	// Step 2: Target transformation (InternalEvent → webhook payload)
	payload, err := p.target.Transform(ctx, internalEvent)
	if err != nil {
		return nil, internalEvent, fmt.Errorf("target transformation failed: %w", err)
	}

	return payload, internalEvent, nil
}

// TransformToInternal performs only source transformation (rawEvent → InternalEvent).
// Useful when you need the unified format but not the final webhook payload.
func (p *Pipeline) TransformToInternal(ctx context.Context, rawEvent interface{}) (*types.InternalEvent, error) {
	return p.source.Transform(ctx, rawEvent)
}

// TransformFromInternal performs only target transformation (InternalEvent → webhook).
// Useful when you already have an InternalEvent from persistence.
func (p *Pipeline) TransformFromInternal(ctx context.Context, event *types.InternalEvent) (json.RawMessage, error) {
	return p.target.Transform(ctx, event)
}

// Common errors
var (
	// ErrUnsupportedEvent indicates the transformer doesn't support this event type
	ErrUnsupportedEvent = fmt.Errorf("unsupported event type")

	// ErrInvalidEvent indicates the event data is malformed or invalid
	ErrInvalidEvent = fmt.Errorf("invalid event data")

	// ErrMissingRequiredField indicates a required field is missing in the event
	ErrMissingRequiredField = fmt.Errorf("missing required field")

	// ErrTransformationFailed indicates a general transformation failure
	ErrTransformationFailed = fmt.Errorf("transformation failed")
)
