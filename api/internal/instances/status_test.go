package instances

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// TestStatusJSONMarshallingMatchesZAPI ensures that the serialized payload only
// surfaces the Z-API compatible fields while keeping the internal attributes
// available in the struct for future use.
func TestStatusJSONMarshallingMatchesZAPI(t *testing.T) {
	t.Parallel()

	storeJID := "551199999999@s.whatsapp.net"
	status := Status{
		Connected:           true,
		ConnectionStatus:    "connected",
		StoreJID:            &storeJID,
		InstanceID:          uuid.MustParse("11111111-2222-3333-4444-555555555555"),
		AutoReconnect:       true,
		WorkerAssigned:      "worker-1",
		SubscriptionActive:  true,
		Error:               "",
		SmartphoneConnected: true,
	}

	payload, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("marshal status: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	expectedKeys := map[string]struct{}{
		"connected":           {},
		"error":               {},
		"smartphoneConnected": {},
	}

	if len(raw) != len(expectedKeys) {
		t.Fatalf("expected %d keys, got %d (%v)", len(expectedKeys), len(raw), raw)
	}

	for key := range raw {
		if _, ok := expectedKeys[key]; !ok {
			t.Fatalf("unexpected key in payload: %s (payload=%v)", key, raw)
		}
	}

	connected, ok := raw["connected"].(bool)
	if !ok || !connected {
		t.Fatalf("connected key missing or false: %v", raw)
	}

	errorVal, ok := raw["error"].(string)
	if !ok {
		t.Fatalf("error key missing or not a string: %v", raw)
	}
	if errorVal != "" {
		t.Fatalf("expected empty error string, got %q", errorVal)
	}

	smartphoneConnected, ok := raw["smartphoneConnected"].(bool)
	if !ok || !smartphoneConnected {
		t.Fatalf("smartphoneConnected key missing or false: %v", raw)
	}
}
