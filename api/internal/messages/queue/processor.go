package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	wameow "go.mau.fi/whatsmeow"

	"go.mau.fi/whatsmeow/api/internal/events/echo"
)

// WhatsAppMessageProcessor orchestrates message processing by delegating to type-specific processors
// This follows the Strategy pattern for clean separation of concerns
type WhatsAppMessageProcessor struct {
	textProcessor            *TextProcessor
	imageProcessor           *ImageProcessor
	audioProcessor           *AudioProcessor
	videoProcessor           *VideoProcessor
	documentProcessor        *DocumentProcessor
	stickerProcessor         *StickerProcessor
	ptvProcessor             *PTVProcessor
	locationProcessor        *LocationProcessor
	contactProcessor         *ContactProcessor
	interactiveProcessor     *InteractiveProcessor
	interactiveZAPIProcessor *InteractiveZAPIProcessor
	pollProcessor            *PollProcessor
	eventProcessor           *EventProcessor
	statusProcessor          *StatusProcessor
	echoEmitter              *echo.Emitter
	log                      *slog.Logger
}

// NewWhatsAppMessageProcessor creates a new message processor with all type-specific processors
// echoEmitter can be nil if API echo is disabled
func NewWhatsAppMessageProcessor(log *slog.Logger, echoEmitter *echo.Emitter) *WhatsAppMessageProcessor {
	return &WhatsAppMessageProcessor{
		textProcessor:            NewTextProcessor(log, echoEmitter),
		imageProcessor:           NewImageProcessor(log, echoEmitter),
		audioProcessor:           NewAudioProcessor(log, echoEmitter),
		videoProcessor:           NewVideoProcessor(log, echoEmitter),
		documentProcessor:        NewDocumentProcessor(log, echoEmitter),
		stickerProcessor:         NewStickerProcessor(log, echoEmitter),
		ptvProcessor:             NewPTVProcessor(log, echoEmitter),
		locationProcessor:        NewLocationProcessor(log, echoEmitter),
		contactProcessor:         NewContactProcessor(log, echoEmitter),
		interactiveProcessor:     NewInteractiveProcessor(log, echoEmitter),
		interactiveZAPIProcessor: NewInteractiveZAPIProcessor(log, echoEmitter),
		pollProcessor:            NewPollProcessor(log, echoEmitter),
		eventProcessor:           NewEventProcessor(log, echoEmitter),
		statusProcessor:          NewStatusProcessor(log, echoEmitter),
		echoEmitter:              echoEmitter,
		log:                      log,
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
	case MessageTypeSticker:
		return p.stickerProcessor.Process(ctx, client, &args)
	case MessageTypePTV:
		return p.ptvProcessor.Process(ctx, client, &args)
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

	// interactive message types
	case MessageTypeButtonList:
		return p.interactiveZAPIProcessor.ProcessButtonList(ctx, client, &args)
	case MessageTypeButtonActions:
		return p.interactiveZAPIProcessor.ProcessButtonActions(ctx, client, &args)
	case MessageTypeOptionList:
		return p.interactiveZAPIProcessor.ProcessOptionList(ctx, client, &args)
	case MessageTypeButtonPIX:
		return p.interactiveZAPIProcessor.ProcessButtonPIX(ctx, client, &args)
	case MessageTypeButtonOTP:
		return p.interactiveZAPIProcessor.ProcessButtonOTP(ctx, client, &args)
	case MessageTypeCarousel:
		return p.interactiveZAPIProcessor.ProcessCarousel(ctx, client, &args)

	// Status/Stories message types (sent to status@broadcast)
	case MessageTypeTextStatus:
		return p.statusProcessor.ProcessText(ctx, client, &args)
	case MessageTypeImageStatus:
		return p.statusProcessor.ProcessImage(ctx, client, &args)
	case MessageTypeAudioStatus:
		return p.statusProcessor.ProcessAudio(ctx, client, &args)
	case MessageTypeVideoStatus:
		return p.statusProcessor.ProcessVideo(ctx, client, &args)

	default:
		return fmt.Errorf("unsupported message type: %s", args.MessageType)
	}
}
