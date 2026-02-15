package nats_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	natsgo "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	natspkg "go.mau.fi/whatsmeow/api/internal/nats"
)

// startEmbeddedNATS starts an embedded NATS server with JetStream for testing.
func startEmbeddedNATS(t *testing.T) *natsserver.Server {
	t.Helper()

	dir := t.TempDir()
	opts := &natsserver.Options{
		Host:      "127.0.0.1",
		Port:      -1, // random port
		JetStream: true,
		StoreDir:  dir,
		NoLog:     true,
		NoSigs:    true,
	}

	srv, err := natsserver.NewServer(opts)
	require.NoError(t, err, "failed to create NATS server")

	srv.Start()
	if !srv.ReadyForConnections(5 * time.Second) {
		t.Fatal("NATS server not ready for connections")
	}

	t.Cleanup(func() {
		srv.Shutdown()
		srv.WaitForShutdown()
	})

	return srv
}

func testConfig(srv *natsserver.Server) natspkg.Config {
	cfg := natspkg.DefaultConfig()
	cfg.URL = srv.ClientURL()
	return cfg
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
}

func testMetrics(t *testing.T) *natspkg.NATSMetrics {
	t.Helper()
	reg := prometheus.NewRegistry()
	return natspkg.NewNATSMetrics("test", reg)
}

func TestClient_Connect(t *testing.T) {
	srv := startEmbeddedNATS(t)
	cfg := testConfig(srv)
	log := testLogger()
	metrics := testMetrics(t)

	client := natspkg.NewClient(cfg, log, metrics)

	err := client.Connect(context.Background())
	require.NoError(t, err)
	defer client.Close()

	assert.True(t, client.IsConnected())
	assert.NotNil(t, client.JetStream())
	assert.NotNil(t, client.Conn())
}

func TestClient_ConnectInvalidURL(t *testing.T) {
	cfg := natspkg.DefaultConfig()
	cfg.URL = "nats://invalid:9999"
	cfg.ConnectTimeout = 1 * time.Second

	client := natspkg.NewClient(cfg, testLogger(), nil)

	err := client.Connect(context.Background())
	assert.Error(t, err)
	assert.False(t, client.IsConnected())
}

func TestClient_ConnectInvalidConfig(t *testing.T) {
	cfg := natspkg.Config{URL: ""}
	client := natspkg.NewClient(cfg, testLogger(), nil)

	err := client.Connect(context.Background())
	assert.Error(t, err)
}

func TestClient_Publish(t *testing.T) {
	srv := startEmbeddedNATS(t)
	cfg := testConfig(srv)
	metrics := testMetrics(t)

	client := natspkg.NewClient(cfg, testLogger(), metrics)
	require.NoError(t, client.Connect(context.Background()))
	defer client.Close()

	// Create a stream to publish to
	js := client.JetStream()
	_, err := js.CreateStream(context.Background(), jetstream.StreamConfig{
		Name:     "TEST",
		Subjects: []string{"test.>"},
	})
	require.NoError(t, err)

	// Publish a message
	ack, err := client.Publish(context.Background(), "test.hello", []byte(`{"msg":"hello"}`))
	require.NoError(t, err)
	assert.NotNil(t, ack)
	assert.Equal(t, "TEST", ack.Stream)
	assert.Equal(t, uint64(1), ack.Sequence)
}

func TestClient_PublishMsg(t *testing.T) {
	srv := startEmbeddedNATS(t)
	cfg := testConfig(srv)

	client := natspkg.NewClient(cfg, testLogger(), testMetrics(t))
	require.NoError(t, client.Connect(context.Background()))
	defer client.Close()

	js := client.JetStream()
	_, err := js.CreateStream(context.Background(), jetstream.StreamConfig{
		Name:     "TEST",
		Subjects: []string{"test.>"},
	})
	require.NoError(t, err)

	msg := &natsgo.Msg{
		Subject: "test.with-headers",
		Data:    []byte(`{"msg":"with headers"}`),
		Header:  natsgo.Header{},
	}
	msg.Header.Set("X-Instance-ID", "abc-123")

	ack, err := client.PublishMsg(context.Background(), msg)
	require.NoError(t, err)
	assert.NotNil(t, ack)
}

func TestClient_PublishNotConnected(t *testing.T) {
	cfg := natspkg.DefaultConfig()
	client := natspkg.NewClient(cfg, testLogger(), nil)

	_, err := client.Publish(context.Background(), "test.hello", []byte("data"))
	assert.ErrorIs(t, err, natspkg.ErrNotConnected)
}

func TestClient_PublishMsgNotConnected(t *testing.T) {
	cfg := natspkg.DefaultConfig()
	client := natspkg.NewClient(cfg, testLogger(), nil)

	msg := &natsgo.Msg{Subject: "test.hello", Data: []byte("data")}
	_, err := client.PublishMsg(context.Background(), msg)
	assert.ErrorIs(t, err, natspkg.ErrNotConnected)
}

func TestClient_Drain(t *testing.T) {
	srv := startEmbeddedNATS(t)
	cfg := testConfig(srv)

	client := natspkg.NewClient(cfg, testLogger(), testMetrics(t))
	require.NoError(t, client.Connect(context.Background()))

	err := client.Drain(5 * time.Second)
	require.NoError(t, err)
	assert.False(t, client.IsConnected())
}

func TestClient_DrainNotConnected(t *testing.T) {
	cfg := natspkg.DefaultConfig()
	client := natspkg.NewClient(cfg, testLogger(), nil)

	err := client.Drain(5 * time.Second)
	assert.NoError(t, err) // no-op when not connected
}

func TestClient_Close(t *testing.T) {
	srv := startEmbeddedNATS(t)
	cfg := testConfig(srv)

	client := natspkg.NewClient(cfg, testLogger(), testMetrics(t))
	require.NoError(t, client.Connect(context.Background()))

	client.Close()
	assert.False(t, client.IsConnected())

	// Double close is safe
	client.Close()
}

func TestClient_EnsureConsumer(t *testing.T) {
	srv := startEmbeddedNATS(t)
	cfg := testConfig(srv)

	client := natspkg.NewClient(cfg, testLogger(), testMetrics(t))
	require.NoError(t, client.Connect(context.Background()))
	defer client.Close()

	// Create stream first
	js := client.JetStream()
	_, err := js.CreateStream(context.Background(), jetstream.StreamConfig{
		Name:     "TEST_CONSUMER",
		Subjects: []string{"tc.>"},
	})
	require.NoError(t, err)

	// Create consumer
	consumer, err := client.EnsureConsumer(context.Background(), "TEST_CONSUMER", jetstream.ConsumerConfig{
		Durable:       "test-consumer",
		FilterSubject: "tc.test",
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxAckPending: 1,
	})
	require.NoError(t, err)
	assert.NotNil(t, consumer)

	// Update consumer (idempotent)
	consumer2, err := client.EnsureConsumer(context.Background(), "TEST_CONSUMER", jetstream.ConsumerConfig{
		Durable:       "test-consumer",
		FilterSubject: "tc.test",
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxAckPending: 5,
	})
	require.NoError(t, err)
	assert.NotNil(t, consumer2)
}

func TestClient_EnsureConsumerNotConnected(t *testing.T) {
	cfg := natspkg.DefaultConfig()
	client := natspkg.NewClient(cfg, testLogger(), nil)

	_, err := client.EnsureConsumer(context.Background(), "STREAM", jetstream.ConsumerConfig{
		Durable: "test",
	})
	assert.ErrorIs(t, err, natspkg.ErrNotConnected)
}

func TestClient_HealthCheck(t *testing.T) {
	srv := startEmbeddedNATS(t)
	cfg := testConfig(srv)

	client := natspkg.NewClient(cfg, testLogger(), testMetrics(t))
	require.NoError(t, client.Connect(context.Background()))
	defer client.Close()

	// Ensure streams so health check has data
	require.NoError(t, natspkg.EnsureAllStreams(context.Background(), client.JetStream(), cfg, testLogger()))

	status := client.HealthCheck(context.Background())
	assert.True(t, status.Connected)
	assert.Empty(t, status.Error)
	assert.Len(t, status.Streams, 4)
}

func TestClient_HealthCheckNotConnected(t *testing.T) {
	cfg := natspkg.DefaultConfig()
	client := natspkg.NewClient(cfg, testLogger(), nil)

	status := client.HealthCheck(context.Background())
	assert.False(t, status.Connected)
	assert.NotEmpty(t, status.Error)
}

func TestClient_StreamInfo(t *testing.T) {
	srv := startEmbeddedNATS(t)
	cfg := testConfig(srv)

	client := natspkg.NewClient(cfg, testLogger(), testMetrics(t))
	require.NoError(t, client.Connect(context.Background()))
	defer client.Close()

	require.NoError(t, natspkg.EnsureAllStreams(context.Background(), client.JetStream(), cfg, testLogger()))

	stats, err := client.StreamInfo(context.Background(), cfg.StreamMessageQueue)
	require.NoError(t, err)
	assert.Equal(t, cfg.StreamMessageQueue, stats.Name)
}

func TestClient_PublishDuplicateDetection(t *testing.T) {
	srv := startEmbeddedNATS(t)
	cfg := testConfig(srv)

	client := natspkg.NewClient(cfg, testLogger(), testMetrics(t))
	require.NoError(t, client.Connect(context.Background()))
	defer client.Close()

	require.NoError(t, natspkg.EnsureAllStreams(context.Background(), client.JetStream(), cfg, testLogger()))

	msgID := fmt.Sprintf("test-dedup-%d", time.Now().UnixNano())

	// First publish should succeed
	ack1, err := client.Publish(context.Background(), "messages.test-instance",
		[]byte(`{"msg":"first"}`),
		jetstream.WithMsgID(msgID),
	)
	require.NoError(t, err)
	assert.False(t, ack1.Duplicate)

	// Second publish with same MsgID should be detected as duplicate
	ack2, err := client.Publish(context.Background(), "messages.test-instance",
		[]byte(`{"msg":"duplicate"}`),
		jetstream.WithMsgID(msgID),
	)
	require.NoError(t, err)
	assert.True(t, ack2.Duplicate)
}
