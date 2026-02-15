package nats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

// MediaResultEntry stores a media processing result in the KV store.
type MediaResultEntry struct {
	Success  bool   `json:"success"`
	MediaURL string `json:"media_url,omitempty"`
	Error    string `json:"error,omitempty"`
}

// MediaResultKV provides a NATS Key-Value backed store for media processing results.
// The CompletionHandler writes results here after media processing completes.
// The NATSDispatchWorker reads results to inject media URLs into webhook payloads.
type MediaResultKV struct {
	bucket jetstream.KeyValue
}

// EnsureMediaResultsBucket creates or gets the MEDIA_RESULTS KV bucket.
// Uses memory storage with 1-hour TTL since results are ephemeral.
func EnsureMediaResultsBucket(ctx context.Context, js jetstream.JetStream) (*MediaResultKV, error) {
	bucket, err := js.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:      "MEDIA_RESULTS",
		TTL:         1 * time.Hour,
		Storage:     jetstream.MemoryStorage,
		Description: "Media processing results for webhook URL injection",
	})
	if err != nil {
		return nil, fmt.Errorf("ensure MEDIA_RESULTS bucket: %w", err)
	}
	return &MediaResultKV{bucket: bucket}, nil
}

// Put stores a media processing result keyed by event ID.
func (kv *MediaResultKV) Put(ctx context.Context, eventID string, success bool, mediaURL string, errMsg string) error {
	entry := MediaResultEntry{
		Success:  success,
		MediaURL: mediaURL,
		Error:    errMsg,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal media result: %w", err)
	}
	_, err = kv.bucket.Put(ctx, eventID, data)
	if err != nil {
		return fmt.Errorf("put media result for %s: %w", eventID, err)
	}
	return nil
}

// LookupMediaURL checks if a media result exists for the given event.
// Returns (url, true, nil) if the result exists (success or failure).
// Returns ("", false, nil) if no result exists yet.
func (kv *MediaResultKV) LookupMediaURL(ctx context.Context, eventID string) (string, bool, error) {
	entry, err := kv.bucket.Get(ctx, eventID)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("get media result for %s: %w", eventID, err)
	}

	var result MediaResultEntry
	if err := json.Unmarshal(entry.Value(), &result); err != nil {
		return "", false, fmt.Errorf("unmarshal media result for %s: %w", eventID, err)
	}

	return result.MediaURL, true, nil
}
