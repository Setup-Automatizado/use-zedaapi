package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"go.mau.fi/whatsmeow/api/internal/events/types"
)

type SourceTransformer interface {
	Transform(ctx context.Context, rawEvent interface{}) (*types.InternalEvent, error)
	SupportsEvent(eventType reflect.Type) bool
	SourceLib() types.SourceLib
}

type TargetTransformer interface {
	Transform(ctx context.Context, event *types.InternalEvent) (json.RawMessage, error)
	TargetSchema() string
	SupportsEventType(eventType string) bool
}

type Pipeline struct {
	source SourceTransformer
	target TargetTransformer
}

func NewPipeline(source SourceTransformer, target TargetTransformer) *Pipeline {
	return &Pipeline{
		source: source,
		target: target,
	}
}

func (p *Pipeline) Transform(ctx context.Context, rawEvent interface{}) (json.RawMessage, *types.InternalEvent, error) {
	internalEvent, err := p.source.Transform(ctx, rawEvent)
	if err != nil {
		return nil, nil, fmt.Errorf("source transformation failed: %w", err)
	}

	payload, err := p.target.Transform(ctx, internalEvent)
	if err != nil {
		return nil, internalEvent, fmt.Errorf("target transformation failed: %w", err)
	}

	return payload, internalEvent, nil
}

func (p *Pipeline) TransformToInternal(ctx context.Context, rawEvent interface{}) (*types.InternalEvent, error) {
	return p.source.Transform(ctx, rawEvent)
}

func (p *Pipeline) TransformFromInternal(ctx context.Context, event *types.InternalEvent) (json.RawMessage, error) {
	return p.target.Transform(ctx, event)
}

var (
	ErrUnsupportedEvent     = fmt.Errorf("unsupported event type")
	ErrInvalidEvent         = fmt.Errorf("invalid event data")
	ErrMissingRequiredField = fmt.Errorf("missing required field")
	ErrTransformationFailed = fmt.Errorf("transformation failed")
)
