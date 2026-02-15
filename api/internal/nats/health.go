package nats

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
)

// StreamStats holds basic stats for a single stream.
type StreamStats struct {
	Name     string `json:"name"`
	Messages uint64 `json:"messages"`
	Bytes    uint64 `json:"bytes"`
	Subjects uint64 `json:"subjects"`
}

// HealthStatus represents the NATS connection health.
type HealthStatus struct {
	Connected bool          `json:"connected"`
	URL       string        `json:"url"`
	Streams   []StreamStats `json:"streams,omitempty"`
	Error     string        `json:"error,omitempty"`
}

// HealthCheck returns the current health of the NATS client.
func (c *Client) HealthCheck(ctx context.Context) HealthStatus {
	status := HealthStatus{
		URL: c.cfg.URL,
	}

	if c.conn == nil || !c.conn.IsConnected() {
		status.Error = "not connected"
		return status
	}

	status.Connected = true

	streams, err := c.AllStreamStats(ctx)
	if err != nil {
		status.Error = fmt.Sprintf("stream stats: %v", err)
		return status
	}
	status.Streams = streams

	return status
}

// StreamInfo returns info for a specific stream.
func (c *Client) StreamInfo(ctx context.Context, streamName string) (*StreamStats, error) {
	if c.js == nil {
		return nil, ErrNotConnected
	}

	stream, err := c.js.Stream(ctx, streamName)
	if err != nil {
		return nil, fmt.Errorf("get stream %s: %w", streamName, err)
	}

	info, err := stream.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("stream info %s: %w", streamName, err)
	}

	return &StreamStats{
		Name:     streamName,
		Messages: info.State.Msgs,
		Bytes:    info.State.Bytes,
		Subjects: info.State.NumSubjects,
	}, nil
}

// AllStreamStats returns stats for all configured streams.
func (c *Client) AllStreamStats(ctx context.Context) ([]StreamStats, error) {
	if c.js == nil {
		return nil, ErrNotConnected
	}

	names := []string{
		c.cfg.StreamMessageQueue,
		c.cfg.StreamWhatsAppEvents,
		c.cfg.StreamMediaProcessing,
		c.cfg.StreamDLQ,
	}

	stats := make([]StreamStats, 0, len(names))
	for _, name := range names {
		stream, err := c.js.Stream(ctx, name)
		if err != nil {
			stats = append(stats, StreamStats{
				Name: name,
			})
			continue
		}

		info, err := stream.Info(ctx)
		if err != nil {
			stats = append(stats, StreamStats{
				Name: name,
			})
			continue
		}

		stats = append(stats, StreamStats{
			Name:     name,
			Messages: info.State.Msgs,
			Bytes:    info.State.Bytes,
			Subjects: info.State.NumSubjects,
		})
	}

	return stats, nil
}

// UpdateStreamMetrics updates Prometheus gauges with current stream state.
func (c *Client) UpdateStreamMetrics(ctx context.Context) {
	if c.metrics == nil || c.js == nil {
		return
	}

	names := []string{
		c.cfg.StreamMessageQueue,
		c.cfg.StreamWhatsAppEvents,
		c.cfg.StreamMediaProcessing,
		c.cfg.StreamDLQ,
	}

	for _, name := range names {
		stream, err := c.js.Stream(ctx, name)
		if err != nil {
			continue
		}

		info, err := stream.Info(ctx)
		if err != nil {
			continue
		}

		c.metrics.StreamMessages.WithLabelValues(name).Set(float64(info.State.Msgs))
		c.metrics.StreamBytes.WithLabelValues(name).Set(float64(info.State.Bytes))
	}
}

// EnsureConsumer creates or updates a consumer on the given stream.
func (c *Client) EnsureConsumer(ctx context.Context, streamName string, cfg jetstream.ConsumerConfig) (jetstream.Consumer, error) {
	if c.js == nil {
		return nil, ErrNotConnected
	}

	consumer, err := c.js.CreateOrUpdateConsumer(ctx, streamName, cfg)
	if err != nil {
		return nil, fmt.Errorf("ensure consumer %s on %s: %w", cfg.Durable, streamName, err)
	}

	return consumer, nil
}
