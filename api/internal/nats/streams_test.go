package nats_test

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	natspkg "go.mau.fi/whatsmeow/api/internal/nats"
)

func TestMessageQueueStreamConfig(t *testing.T) {
	cfg := natspkg.MessageQueueStreamConfig("MESSAGE_QUEUE")

	assert.Equal(t, "MESSAGE_QUEUE", cfg.Name)
	assert.Equal(t, []string{"messages.>"}, cfg.Subjects)
	assert.Equal(t, jetstream.WorkQueuePolicy, cfg.Retention)
	assert.Equal(t, 72*time.Hour, cfg.MaxAge)
	assert.Equal(t, int64(10*1024*1024*1024), cfg.MaxBytes)
	assert.Equal(t, jetstream.FileStorage, cfg.Storage)
	assert.Equal(t, jetstream.DiscardOld, cfg.Discard)
	assert.Equal(t, 2*time.Minute, cfg.Duplicates)
	assert.Equal(t, int32(64*1024*1024), cfg.MaxMsgSize)
	assert.False(t, cfg.NoAck)
}

func TestWhatsAppEventsStreamConfig(t *testing.T) {
	cfg := natspkg.WhatsAppEventsStreamConfig("WHATSAPP_EVENTS")

	assert.Equal(t, "WHATSAPP_EVENTS", cfg.Name)
	assert.Equal(t, []string{"events.>"}, cfg.Subjects)
	assert.Equal(t, 168*time.Hour, cfg.MaxAge)
	assert.Equal(t, int64(50*1024*1024*1024), cfg.MaxBytes)
	assert.Equal(t, 1*time.Hour, cfg.Duplicates)
	assert.Equal(t, int32(64*1024*1024), cfg.MaxMsgSize)
}

func TestMediaProcessingStreamConfig(t *testing.T) {
	cfg := natspkg.MediaProcessingStreamConfig("MEDIA_PROCESSING")

	assert.Equal(t, "MEDIA_PROCESSING", cfg.Name)
	assert.Equal(t, []string{"media.tasks.>", "media.done.>"}, cfg.Subjects)
	assert.Equal(t, 168*time.Hour, cfg.MaxAge)
	assert.Equal(t, int64(5*1024*1024*1024), cfg.MaxBytes)
}

func TestDLQStreamConfig(t *testing.T) {
	cfg := natspkg.DLQStreamConfig("DLQ")

	assert.Equal(t, "DLQ", cfg.Name)
	assert.Equal(t, []string{"dlq.>"}, cfg.Subjects)
	assert.Equal(t, 720*time.Hour, cfg.MaxAge)
}

func TestMessageConsumerConfig(t *testing.T) {
	cfg := natspkg.MessageConsumerConfig("test-instance-id")

	assert.Equal(t, "msg-test-instance-id", cfg.Durable)
	assert.Equal(t, "messages.test-instance-id", cfg.FilterSubject)
	assert.Equal(t, jetstream.AckExplicitPolicy, cfg.AckPolicy)
	assert.Equal(t, 30*time.Second, cfg.AckWait)
	assert.Equal(t, 10, cfg.MaxDeliver)
	assert.Equal(t, 1, cfg.MaxAckPending) // FIFO guarantee
	assert.Len(t, cfg.BackOff, 10)
	// BackOff[0] and [1] should be short (1s) for quick multi-replica failover
	assert.Equal(t, 1*time.Second, cfg.BackOff[0])
	assert.Equal(t, 1*time.Second, cfg.BackOff[1])
}

func TestEventConsumerConfig(t *testing.T) {
	cfg := natspkg.EventConsumerConfig("test-instance-id")

	assert.Equal(t, "evt-test-instance-id", cfg.Durable)
	assert.Equal(t, "events.test-instance-id.>", cfg.FilterSubject)
	assert.Equal(t, 60*time.Second, cfg.AckWait)
	assert.Equal(t, 6, cfg.MaxDeliver)
	assert.Equal(t, 10, cfg.MaxAckPending)
	assert.Len(t, cfg.BackOff, 6)
	// BackOff[0] MUST be > 0; setting it to 0 causes immediate redelivery
	// because the worker cannot ACK within 0 seconds.
	assert.Greater(t, cfg.BackOff[0], time.Duration(0), "BackOff[0] must be > 0 to prevent immediate redelivery")
}

func TestMediaConsumerConfig(t *testing.T) {
	cfg := natspkg.MediaConsumerConfig("test-instance-id")

	assert.Equal(t, "media-test-instance-id", cfg.Durable)
	assert.Equal(t, "media.tasks.test-instance-id", cfg.FilterSubject)
	assert.Equal(t, 120*time.Second, cfg.AckWait)
	assert.Equal(t, 5, cfg.MaxDeliver)
	assert.Equal(t, 3, cfg.MaxAckPending)
	assert.Len(t, cfg.BackOff, 5)
}

func TestEnsureAllStreams(t *testing.T) {
	srv := startEmbeddedNATS(t)
	cfg := testConfig(srv)

	client := natspkg.NewClient(cfg, testLogger(), testMetrics(t))
	require.NoError(t, client.Connect(context.Background()))
	defer client.Close()

	// First call creates streams
	err := natspkg.EnsureAllStreams(context.Background(), client.JetStream(), cfg, testLogger())
	require.NoError(t, err)

	// Second call is idempotent (update)
	err = natspkg.EnsureAllStreams(context.Background(), client.JetStream(), cfg, testLogger())
	require.NoError(t, err)

	// Verify all 4 streams exist
	js := client.JetStream()

	streams := []string{
		cfg.StreamMessageQueue,
		cfg.StreamWhatsAppEvents,
		cfg.StreamMediaProcessing,
		cfg.StreamDLQ,
	}
	for _, name := range streams {
		stream, err := js.Stream(context.Background(), name)
		require.NoError(t, err, "stream %s should exist", name)

		info, err := stream.Info(context.Background())
		require.NoError(t, err)
		assert.Equal(t, name, info.Config.Name)
	}
}

func TestEnsureAllStreams_PublishToEachStream(t *testing.T) {
	srv := startEmbeddedNATS(t)
	cfg := testConfig(srv)

	client := natspkg.NewClient(cfg, testLogger(), testMetrics(t))
	require.NoError(t, client.Connect(context.Background()))
	defer client.Close()

	require.NoError(t, natspkg.EnsureAllStreams(context.Background(), client.JetStream(), cfg, testLogger()))

	tests := []struct {
		name    string
		subject string
		stream  string
	}{
		{"message queue", "messages.test-instance", cfg.StreamMessageQueue},
		{"events", "events.test-instance.message", cfg.StreamWhatsAppEvents},
		{"media tasks", "media.tasks.test-instance", cfg.StreamMediaProcessing},
		{"media done", "media.done.test-instance.evt1", cfg.StreamMediaProcessing},
		{"dlq messages", "dlq.messages.test-instance", cfg.StreamDLQ},
		{"dlq events", "dlq.events.test-instance", cfg.StreamDLQ},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ack, err := client.Publish(context.Background(), tt.subject, []byte(`{"test":true}`))
			require.NoError(t, err)
			assert.Equal(t, tt.stream, ack.Stream, "subject %s should route to stream %s", tt.subject, tt.stream)
		})
	}
}

func TestUpdateStreamMetrics(t *testing.T) {
	srv := startEmbeddedNATS(t)
	cfg := testConfig(srv)
	metrics := testMetrics(t)

	client := natspkg.NewClient(cfg, testLogger(), metrics)
	require.NoError(t, client.Connect(context.Background()))
	defer client.Close()

	require.NoError(t, natspkg.EnsureAllStreams(context.Background(), client.JetStream(), cfg, testLogger()))

	// Publish a message to have data
	_, err := client.Publish(context.Background(), "messages.test", []byte(`{"test":true}`))
	require.NoError(t, err)

	// Update metrics
	client.UpdateStreamMetrics(context.Background())

	// No panic, metrics updated (hard to read prometheus gauges in test without full collector)
}
