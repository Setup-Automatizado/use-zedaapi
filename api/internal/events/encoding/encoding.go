package encoding

import (
    "bytes"
    "encoding/base64"
    "encoding/gob"
    "fmt"

    "github.com/google/uuid"

    "go.mau.fi/whatsmeow/api/internal/events/types"
    whatsmeowevents "go.mau.fi/whatsmeow/types/events"
)

func init() {
    // Register core whatsmeow event types so gob can encode/decode them when stored via interface{}
    gob.Register(&whatsmeowevents.Message{})
    gob.Register(&whatsmeowevents.Receipt{})
    gob.Register(&whatsmeowevents.ChatPresence{})
    gob.Register(&whatsmeowevents.Presence{})
    gob.Register(&whatsmeowevents.Connected{})
    gob.Register(&whatsmeowevents.Disconnected{})
    gob.Register(&whatsmeowevents.JoinedGroup{})
    gob.Register(&whatsmeowevents.GroupInfo{})
    gob.Register(&whatsmeowevents.Picture{})
    // Also register supporting types that may appear in metadata payloads
    gob.Register(uuid.UUID{})
}

// EncodeInternalEvent serialises the provided InternalEvent using gob and returns a base64
// encoded string representation that is safe to persist in JSON columns.
func EncodeInternalEvent(event *types.InternalEvent) (string, error) {
    if event == nil {
        return "", fmt.Errorf("internal event is nil")
    }

    var buf bytes.Buffer
    encoder := gob.NewEncoder(&buf)
    if err := encoder.Encode(event); err != nil {
        return "", fmt.Errorf("encode internal event: %w", err)
    }

    return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// DecodeInternalEvent decodes a base64 gob string back into an InternalEvent instance.
func DecodeInternalEvent(encoded string) (*types.InternalEvent, error) {
    if encoded == "" {
        return nil, fmt.Errorf("encoded internal event payload is empty")
    }

    raw, err := base64.StdEncoding.DecodeString(encoded)
    if err != nil {
        return nil, fmt.Errorf("base64 decode internal event: %w", err)
    }

    decoder := gob.NewDecoder(bytes.NewReader(raw))
    var event types.InternalEvent
    if err := decoder.Decode(&event); err != nil {
        return nil, fmt.Errorf("decode internal event: %w", err)
    }

    return &event, nil
}
