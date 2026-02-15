#!/bin/sh
# Idempotent stream creation for FunnelChat NATS JetStream
# This script is run as an init container in docker-compose

set -e

NATS_URL="${NATS_URL:-nats://nats:4222}"
MAX_RETRIES=30
RETRY_INTERVAL=2

echo "Waiting for NATS server at ${NATS_URL}..."

# Wait for NATS to be ready
retries=0
until nats server check connection --server="${NATS_URL}" 2>/dev/null; do
  retries=$((retries + 1))
  if [ "$retries" -ge "$MAX_RETRIES" ]; then
    echo "ERROR: NATS server not available after ${MAX_RETRIES} attempts"
    exit 1
  fi
  echo "Waiting for NATS... (attempt ${retries}/${MAX_RETRIES})"
  sleep "$RETRY_INTERVAL"
done

echo "NATS server is ready. Creating streams..."

# MESSAGE_QUEUE: Per-instance message sending queue
# subjects: messages.{instance_id}
nats stream add MESSAGE_QUEUE \
  --server="${NATS_URL}" \
  --subjects="messages.>" \
  --retention=work \
  --max-age=72h \
  --max-bytes=10737418240 \
  --storage=file \
  --replicas=1 \
  --discard=old \
  --dupe-window=2m \
  --max-msg-size=8388608 \
  --no-deny-delete \
  --no-deny-purge \
  --defaults 2>/dev/null || \
nats stream update MESSAGE_QUEUE \
  --server="${NATS_URL}" \
  --subjects="messages.>" \
  --retention=work \
  --max-age=72h \
  --max-bytes=10737418240 \
  --discard=old \
  --dupe-window=2m \
  --max-msg-size=8388608 \
  --force 2>/dev/null || true

echo "Stream MESSAGE_QUEUE created/updated"

# WHATSAPP_EVENTS: Event capture and webhook dispatch
# subjects: events.{instance_id}.{event_type}
nats stream add WHATSAPP_EVENTS \
  --server="${NATS_URL}" \
  --subjects="events.>" \
  --retention=limits \
  --max-age=168h \
  --max-bytes=53687091200 \
  --storage=file \
  --replicas=1 \
  --discard=old \
  --dupe-window=1h \
  --max-msg-size=2097152 \
  --no-deny-delete \
  --no-deny-purge \
  --defaults 2>/dev/null || \
nats stream update WHATSAPP_EVENTS \
  --server="${NATS_URL}" \
  --subjects="events.>" \
  --retention=limits \
  --max-age=168h \
  --max-bytes=53687091200 \
  --discard=old \
  --dupe-window=1h \
  --max-msg-size=2097152 \
  --force 2>/dev/null || true

echo "Stream WHATSAPP_EVENTS created/updated"

# MEDIA_PROCESSING: Async media download/upload tasks
# subjects: media.tasks.{instance_id}, media.done.{instance_id}.{event_id}
nats stream add MEDIA_PROCESSING \
  --server="${NATS_URL}" \
  --subjects="media.tasks.>,media.done.>" \
  --retention=limits \
  --max-age=168h \
  --max-bytes=5368709120 \
  --storage=file \
  --replicas=1 \
  --discard=old \
  --dupe-window=2m \
  --max-msg-size=1048576 \
  --no-deny-delete \
  --no-deny-purge \
  --defaults 2>/dev/null || \
nats stream update MEDIA_PROCESSING \
  --server="${NATS_URL}" \
  --subjects="media.tasks.>,media.done.>" \
  --retention=limits \
  --max-age=168h \
  --max-bytes=5368709120 \
  --discard=old \
  --dupe-window=2m \
  --max-msg-size=1048576 \
  --force 2>/dev/null || true

echo "Stream MEDIA_PROCESSING created/updated"

# DLQ: Dead letter queue for failed messages/events
# subjects: dlq.{stream}.{instance_id}
nats stream add DLQ \
  --server="${NATS_URL}" \
  --subjects="dlq.>" \
  --retention=limits \
  --max-age=720h \
  --max-bytes=5368709120 \
  --storage=file \
  --replicas=1 \
  --discard=old \
  --dupe-window=2m \
  --max-msg-size=2097152 \
  --no-deny-delete \
  --no-deny-purge \
  --defaults 2>/dev/null || \
nats stream update DLQ \
  --server="${NATS_URL}" \
  --subjects="dlq.>" \
  --retention=limits \
  --max-age=720h \
  --max-bytes=5368709120 \
  --discard=old \
  --dupe-window=2m \
  --max-msg-size=2097152 \
  --force 2>/dev/null || true

echo "Stream DLQ created/updated"

echo ""
echo "All streams created successfully:"
nats stream ls --server="${NATS_URL}"
echo ""
echo "Stream setup complete."
