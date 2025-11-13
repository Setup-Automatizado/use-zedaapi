package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	wameow "go.mau.fi/whatsmeow"
)

// WhatsAppMessageProcessor orchestrates message processing by delegating to type-specific processors
// This follows the Strategy pattern for clean separation of concerns
type WhatsAppMessageProcessor struct {
	textProcessor        *TextProcessor
	imageProcessor       *ImageProcessor
	audioProcessor       *AudioProcessor
	videoProcessor       *VideoProcessor
	documentProcessor    *DocumentProcessor
	locationProcessor    *LocationProcessor
	contactProcessor     *ContactProcessor
	interactiveProcessor *InteractiveProcessor
	pollProcessor        *PollProcessor
	eventProcessor       *EventProcessor
	log                  *slog.Logger
}

// NewWhatsAppMessageProcessor creates a new message processor with all type-specific processors
func NewWhatsAppMessageProcessor(log *slog.Logger) *WhatsAppMessageProcessor {
	return &WhatsAppMessageProcessor{
		textProcessor:        NewTextProcessor(log),
		imageProcessor:       NewImageProcessor(log),
		audioProcessor:       NewAudioProcessor(log),
		videoProcessor:       NewVideoProcessor(log),
		documentProcessor:    NewDocumentProcessor(log),
		locationProcessor:    NewLocationProcessor(log),
		contactProcessor:     NewContactProcessor(log),
		interactiveProcessor: NewInteractiveProcessor(log),
		pollProcessor:        NewPollProcessor(log),
		eventProcessor:       NewEventProcessor(log),
		log:                  log,
	}
}

// Process implements MessageProcessor interface
// Routes the message to the appropriate type-specific processor
func (p *WhatsAppMessageProcessor) Process(ctx context.Context, client *wameow.Client, payload []byte) error {
	// Parse payload to extract message details
	var args SendMessageArgs
	if err := json.Unmarshal(payload, &args); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	// Validate message arguments
	if err := args.Validate(); err != nil {
		return fmt.Errorf("invalid message args: %w", err)
	}

	p.log.Debug("processing message",
		slog.String("zaap_id", args.ZaapID),
		slog.String("phone", args.Phone),
		slog.String("type", string(args.MessageType)),
		slog.Int64("delay_typing", args.DelayTyping))

	// Route to appropriate processor based on message type
	switch args.MessageType {
	case MessageTypeText:
		return p.textProcessor.Process(ctx, client, &args)
	case MessageTypeImage:
		return p.imageProcessor.Process(ctx, client, &args)
	case MessageTypeAudio:
		return p.audioProcessor.Process(ctx, client, &args)
	case MessageTypeVideo:
		return p.videoProcessor.Process(ctx, client, &args)
	case MessageTypeDocument:
		return p.documentProcessor.Process(ctx, client, &args)
	case MessageTypeLocation:
		return p.locationProcessor.Process(ctx, client, &args)
	case MessageTypeContact:
		return p.contactProcessor.Process(ctx, client, &args)
	case MessageTypeInteractive:
		return p.interactiveProcessor.Process(ctx, client, &args)
	case MessageTypePoll:
		return p.pollProcessor.Process(ctx, client, &args)
	case MessageTypeEvent:
		return p.eventProcessor.Process(ctx, client, &args)
	default:
		return fmt.Errorf("unsupported message type: %s", args.MessageType)
	}
}
