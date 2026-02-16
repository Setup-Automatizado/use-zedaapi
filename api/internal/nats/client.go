package nats

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	natsgo "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Client wraps a NATS connection with JetStream support, reconnect handling,
// publish helpers, and graceful drain/close.
type Client struct {
	cfg     Config
	conn    *natsgo.Conn
	js      jetstream.JetStream
	log     *slog.Logger
	metrics *NATSMetrics

	mu     sync.RWMutex
	closed bool
}

// NewClient creates a new NATS client but does not connect.
// Call Connect() to establish the connection.
func NewClient(cfg Config, log *slog.Logger, metrics *NATSMetrics) *Client {
	return &Client{
		cfg:     cfg,
		log:     log.With(slog.String("component", "nats_client")),
		metrics: metrics,
	}
}

// Connect establishes the NATS connection and initializes JetStream.
func (c *Client) Connect(ctx context.Context) error {
	if err := c.cfg.Validate(); err != nil {
		return fmt.Errorf("nats config: %w", err)
	}

	opts := []natsgo.Option{
		natsgo.Name("zedaapi"),
		natsgo.Timeout(c.cfg.ConnectTimeout),
		natsgo.ReconnectWait(c.cfg.ReconnectWait),
		natsgo.MaxReconnects(c.cfg.MaxReconnects),
		natsgo.DisconnectErrHandler(c.onDisconnect),
		natsgo.ReconnectHandler(c.onReconnect),
		natsgo.ClosedHandler(c.onClosed),
		natsgo.ErrorHandler(c.onError),
	}

	if c.cfg.Token != "" {
		opts = append(opts, natsgo.Token(c.cfg.Token))
	}

	conn, err := natsgo.Connect(c.cfg.URL, opts...)
	if err != nil {
		return fmt.Errorf("nats connect to %s: %w", c.cfg.URL, err)
	}

	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return fmt.Errorf("jetstream init: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.js = js
	c.mu.Unlock()

	if c.metrics != nil {
		c.metrics.ConnectionStatus.Set(1)
	}

	c.log.Info("connected to NATS",
		slog.String("url", c.cfg.URL),
		slog.String("server_id", conn.ConnectedServerId()),
	)

	return nil
}

// Publish publishes a message to the given subject via JetStream.
func (c *Client) Publish(ctx context.Context, subject string, data []byte, opts ...jetstream.PublishOpt) (*jetstream.PubAck, error) {
	c.mu.RLock()
	js := c.js
	c.mu.RUnlock()

	if js == nil {
		return nil, ErrNotConnected
	}

	start := time.Now()
	ack, err := js.Publish(ctx, subject, data, opts...)
	duration := time.Since(start)

	streamName := c.streamForSubject(subject)

	if c.metrics != nil {
		c.metrics.PublishDuration.WithLabelValues(streamName).Observe(duration.Seconds())
		if err != nil {
			c.metrics.PublishTotal.WithLabelValues(streamName, "error").Inc()
			c.metrics.PublishErrors.WithLabelValues(streamName, "publish").Inc()
		} else {
			c.metrics.PublishTotal.WithLabelValues(streamName, "success").Inc()
		}
	}

	if err != nil {
		return nil, fmt.Errorf("publish to %s: %w", subject, err)
	}

	return ack, nil
}

// PublishMsg publishes a nats.Msg via JetStream with full header control.
func (c *Client) PublishMsg(ctx context.Context, msg *natsgo.Msg, opts ...jetstream.PublishOpt) (*jetstream.PubAck, error) {
	c.mu.RLock()
	js := c.js
	c.mu.RUnlock()

	if js == nil {
		return nil, ErrNotConnected
	}

	start := time.Now()
	ack, err := js.PublishMsg(ctx, msg, opts...)
	duration := time.Since(start)

	streamName := c.streamForSubject(msg.Subject)

	if c.metrics != nil {
		c.metrics.PublishDuration.WithLabelValues(streamName).Observe(duration.Seconds())
		if err != nil {
			c.metrics.PublishTotal.WithLabelValues(streamName, "error").Inc()
			c.metrics.PublishErrors.WithLabelValues(streamName, "publish").Inc()
		} else {
			c.metrics.PublishTotal.WithLabelValues(streamName, "success").Inc()
		}
	}

	if err != nil {
		return nil, fmt.Errorf("publish msg to %s: %w", msg.Subject, err)
	}

	return ack, nil
}

// JetStream returns the underlying JetStream context for advanced operations.
func (c *Client) JetStream() jetstream.JetStream {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.js
}

// Conn returns the underlying NATS connection.
func (c *Client) Conn() *natsgo.Conn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

// IsConnected returns true if the NATS connection is active.
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil && c.conn.IsConnected()
}

// Drain initiates a graceful drain of the connection, waiting for in-flight
// messages to complete before closing.
func (c *Client) Drain(timeout time.Duration) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return nil
	}

	c.log.Info("draining NATS connection", slog.Duration("timeout", timeout))

	if err := conn.Drain(); err != nil {
		return fmt.Errorf("nats drain: %w", err)
	}

	// Wait for drain to complete or timeout
	deadline := time.After(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			c.log.Warn("NATS drain timeout exceeded, forcing close")
			conn.Close()
			return ErrDrainTimeout
		case <-ticker.C:
			if conn.IsClosed() {
				c.log.Info("NATS drain completed")
				return nil
			}
		}
	}
}

// Close immediately closes the NATS connection.
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}
	c.closed = true

	if c.conn != nil {
		c.conn.Close()
	}

	if c.metrics != nil {
		c.metrics.ConnectionStatus.Set(0)
	}

	c.log.Info("NATS connection closed")
}

// streamForSubject maps a subject to a stream name for metrics labeling.
func (c *Client) streamForSubject(subject string) string {
	if len(subject) >= 8 && subject[:8] == "messages" {
		return c.cfg.StreamMessageQueue
	}
	if len(subject) >= 6 && subject[:6] == "events" {
		return c.cfg.StreamWhatsAppEvents
	}
	if len(subject) >= 5 && subject[:5] == "media" {
		return c.cfg.StreamMediaProcessing
	}
	if len(subject) >= 3 && subject[:3] == "dlq" {
		return c.cfg.StreamDLQ
	}
	return "unknown"
}

// Connection event handlers

func (c *Client) onDisconnect(conn *natsgo.Conn, err error) {
	if c.metrics != nil {
		c.metrics.ConnectionStatus.Set(0)
		c.metrics.DisconnectionTotal.Inc()
	}
	if err != nil {
		c.log.Warn("NATS disconnected", slog.String("error", err.Error()))
	} else {
		c.log.Warn("NATS disconnected")
	}
}

func (c *Client) onReconnect(conn *natsgo.Conn) {
	if c.metrics != nil {
		c.metrics.ConnectionStatus.Set(1)
		c.metrics.ReconnectionTotal.Inc()
	}
	c.log.Info("NATS reconnected",
		slog.String("url", conn.ConnectedUrl()),
		slog.String("server_id", conn.ConnectedServerId()),
	)
}

func (c *Client) onClosed(conn *natsgo.Conn) {
	if c.metrics != nil {
		c.metrics.ConnectionStatus.Set(0)
	}
	c.log.Info("NATS connection closed")
}

func (c *Client) onError(conn *natsgo.Conn, sub *natsgo.Subscription, err error) {
	if c.metrics != nil {
		c.metrics.ConnectionErrorTotal.Inc()
	}
	fields := []any{slog.String("error", err.Error())}
	if sub != nil {
		fields = append(fields, slog.String("subject", sub.Subject))
	}
	c.log.Error("NATS async error", fields...)
}
