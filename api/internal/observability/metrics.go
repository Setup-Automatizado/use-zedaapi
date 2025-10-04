package observability

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	LockReacquireResultSuccess  = "success"
	LockReacquireResultFailure  = "failure"
	LockReacquireResultFallback = "fallback"
)

type Metrics struct {
	HTTPRequests                   *prometheus.CounterVec
	HTTPDuration                   *prometheus.HistogramVec
	WebhookQueue                   prometheus.Gauge
	LockAcquisitions               *prometheus.CounterVec
	LockReacquisitionAttempts      *prometheus.CounterVec
	LockReacquisitionFallbacks     *prometheus.CounterVec
	CircuitBreakerState            prometheus.Gauge
	SplitBrainDetected             prometheus.Counter
	SplitBrainInvalidLocks         *prometheus.CounterVec
	HealthChecks                   *prometheus.CounterVec
	EventsCaptured                 *prometheus.CounterVec
	EventsBuffered                 prometheus.Gauge
	EventsInserted                 *prometheus.CounterVec
	EventsProcessed                *prometheus.CounterVec
	EventProcessingDuration        *prometheus.HistogramVec
	EventRetries                   *prometheus.CounterVec
	EventsFailed                   *prometheus.CounterVec
	EventsDelivered                *prometheus.CounterVec
	EventDeliveryDuration          *prometheus.HistogramVec
	EventSequenceGaps              *prometheus.GaugeVec
	EventOutboxBacklog             *prometheus.GaugeVec
	DLQEventsTotal                 *prometheus.CounterVec
	DLQReprocessAttempts           *prometheus.CounterVec
	DLQReprocessSuccess            *prometheus.CounterVec
	DLQBacklog                     prometheus.Gauge
	MediaDownloadsTotal            *prometheus.CounterVec
	MediaDownloadDuration          *prometheus.HistogramVec
	MediaDownloadSize              *prometheus.HistogramVec
	MediaDownloadErrors            *prometheus.CounterVec
	MediaUploadsTotal              *prometheus.CounterVec
	MediaUploadDuration            *prometheus.HistogramVec
	MediaUploadAttempts            *prometheus.CounterVec
	MediaUploadErrors              *prometheus.CounterVec
	MediaUploadSizeBytes           *prometheus.CounterVec
	MediaPresignedURLGenerated     prometheus.Counter
	MediaDeleteAttempts            *prometheus.CounterVec
	MediaFailures                  *prometheus.CounterVec
	MediaBacklog                   prometheus.Gauge
	MediaFallbackAttempts          *prometheus.CounterVec
	MediaFallbackSuccess           *prometheus.CounterVec
	MediaFallbackFailure           *prometheus.CounterVec
	MediaLocalStorageSize          prometheus.Gauge
	MediaLocalStorageFiles         prometheus.Gauge
	MediaCleanupTotal              *prometheus.CounterVec
	MediaServeRequests             *prometheus.CounterVec
	MediaServeBytes                *prometheus.CounterVec
	TransportDeliveries            *prometheus.CounterVec
	TransportDuration              *prometheus.HistogramVec
	TransportErrors                *prometheus.CounterVec
	TransportRetries               *prometheus.CounterVec
	CircuitBreakerStatePerInstance *prometheus.GaugeVec
	CircuitBreakerTransitions      *prometheus.CounterVec
	WorkersActive                  *prometheus.GaugeVec
	WorkerTaskDuration             *prometheus.HistogramVec
	WorkerErrors                   *prometheus.CounterVec
}

func NewMetrics(namespace string, reg prometheus.Registerer) *Metrics {
	labels := []string{"method", "path", "status"}
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "http_requests_total",
		Help:      "Total HTTP requests processed.",
	}, labels)
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "http_request_duration_seconds",
		Help:      "Duration of HTTP requests in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, labels)
	queue := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "webhook_outbox_backlog",
		Help:      "Number of webhook events pending delivery.",
	})

	lockAcquisitions := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "lock_acquisitions_total",
		Help:      "Total lock acquisition attempts, labeled by status (success/failure).",
	}, []string{"status"})

	circuitBreakerState := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "circuit_breaker_state",
		Help:      "Current circuit breaker state (0=CLOSED, 1=OPEN, 2=HALF_OPEN).",
	})

	splitBrainDetected := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "split_brain_detected_total",
		Help:      "Total number of split-brain conditions detected.",
	})

	lockReacquisitionAttempts := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "lock_reacquisition_attempts_total",
		Help:      "Total lock reacquisition attempts after Redis recovery, labeled by instance_id and result (success/failure/fallback).",
	}, []string{"instance_id", "result"})

	lockReacquisitionFallbacks := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "lock_reacquisition_fallbacks_total",
		Help:      "Lock reacquisition attempts that returned fallback locks, labeled by instance_id and circuit_state.",
	}, []string{"instance_id", "circuit_state"})

	splitBrainInvalidLocks := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "split_brain_invalid_locks_total",
		Help:      "Locks marked as redis mode but with empty tokens (prevents false positives), labeled by instance_id.",
	}, []string{"instance_id"})

	healthChecks := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "health_checks_total",
		Help:      "Total health check attempts, labeled by component and status.",
	}, []string{"component", "status"})

	eventsCaptured := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "events_captured_total",
		Help:      "Total events captured from WhatsApp, labeled by instance_id and event_type.",
	}, []string{"instance_id", "event_type", "source_lib"})

	eventsBuffered := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "events_buffered",
		Help:      "Number of events currently in buffer waiting to be persisted.",
	})

	eventsInserted := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "events_inserted_total",
		Help:      "Total events inserted into outbox, labeled by instance_id and event_type.",
	}, []string{"instance_id", "event_type", "status"})

	eventsProcessed := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "events_processed_total",
		Help:      "Total events processed by workers, labeled by instance_id and status.",
	}, []string{"instance_id", "event_type", "status"})

	eventProcessingDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "event_processing_duration_seconds",
		Help:      "Duration of event processing in seconds.",
		Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}, []string{"instance_id", "event_type"})

	eventRetries := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "event_retries_total",
		Help:      "Total event retry attempts, labeled by instance_id, event_type, and attempt number.",
	}, []string{"instance_id", "event_type", "attempt"})

	eventsFailed := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "events_failed_total",
		Help:      "Total events that failed permanently, labeled by instance_id, event_type, and reason.",
	}, []string{"instance_id", "event_type", "reason"})

	eventsDelivered := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "events_delivered_total",
		Help:      "Total events successfully delivered, labeled by instance_id, event_type, and transport.",
	}, []string{"instance_id", "event_type", "transport"})

	eventDeliveryDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "event_delivery_duration_seconds",
		Help:      "Duration from event creation to successful delivery in seconds.",
		Buckets:   []float64{.01, .05, .1, .5, 1, 5, 10, 30, 60, 300, 600},
	}, []string{"instance_id", "event_type", "transport"})

	eventSequenceGaps := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "event_sequence_gaps",
		Help:      "Number of sequence gaps detected per instance.",
	}, []string{"instance_id"})

	eventOutboxBacklog := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "event_outbox_backlog",
		Help:      "Number of pending events in outbox per instance.",
	}, []string{"instance_id"})

	dlqEventsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "dlq_events_total",
		Help:      "Total events moved to DLQ, labeled by instance_id, event_type, and failure_reason.",
	}, []string{"instance_id", "event_type", "failure_reason"})

	dlqReprocessAttempts := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "dlq_reprocess_attempts_total",
		Help:      "Total DLQ reprocessing attempts, labeled by instance_id and event_id.",
	}, []string{"instance_id", "event_id"})

	dlqReprocessSuccess := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "dlq_reprocess_success_total",
		Help:      "Total successful DLQ reprocessing, labeled by instance_id and event_type.",
	}, []string{"instance_id", "event_type"})

	dlqBacklog := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "dlq_backlog",
		Help:      "Number of events currently in DLQ.",
	})

	mediaDownloadsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "media_downloads_total",
		Help:      "Total media download attempts, labeled by instance_id, media_type, and status.",
	}, []string{"instance_id", "media_type", "status"})

	mediaDownloadDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "media_download_duration_seconds",
		Help:      "Duration of media downloads in seconds.",
		Buckets:   []float64{.1, .5, 1, 2, 5, 10, 30, 60, 120, 300},
	}, []string{"instance_id", "media_type"})

	mediaDownloadSize := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "media_download_size_bytes",
		Help:      "Size of downloaded media in bytes.",
		Buckets:   []float64{1024, 10240, 102400, 1048576, 10485760, 104857600, 1073741824},
	}, []string{"instance_id", "media_type"})

	mediaDownloadErrors := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "media_download_errors_total",
		Help:      "Total WhatsApp media download errors, labeled by error_type.",
	}, []string{"error_type"})

	mediaUploadsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "media_uploads_total",
		Help:      "Total media upload attempts to S3, labeled by instance_id, media_type, and status.",
	}, []string{"instance_id", "media_type", "status"})

	mediaUploadDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "media_upload_duration_seconds",
		Help:      "Duration of media uploads to S3 in seconds.",
		Buckets:   []float64{.1, .5, 1, 2, 5, 10, 30, 60, 120, 300, 600},
	}, []string{"instance_id", "media_type"})

	mediaFailures := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "media_failures_total",
		Help:      "Total media processing failures, labeled by instance_id, media_type, and stage.",
	}, []string{"instance_id", "media_type", "stage"})

	mediaBacklog := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "media_backlog",
		Help:      "Number of media items pending download.",
	})

	mediaFallbackAttempts := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "media_fallback_attempts_total",
		Help:      "Total media fallback attempts, labeled by instance_id, media_type, and fallback_type (s3/local).",
	}, []string{"instance_id", "media_type", "fallback_type"})

	mediaFallbackSuccess := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "media_fallback_success_total",
		Help:      "Total successful media fallback operations, labeled by instance_id, media_type, and storage_type.",
	}, []string{"instance_id", "media_type", "storage_type"})

	mediaFallbackFailure := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "media_fallback_failure_total",
		Help:      "Total failed media fallback operations, labeled by instance_id, media_type, and error_type.",
	}, []string{"instance_id", "media_type", "error_type"})

	mediaLocalStorageSize := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "media_local_storage_bytes",
		Help:      "Total size in bytes of local media storage.",
	})

	mediaLocalStorageFiles := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "media_local_storage_files",
		Help:      "Total number of files in local media storage.",
	})

	mediaCleanupTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "media_cleanup_total",
		Help:      "Total media cleanup operations, labeled by cleanup_type (expired_files/stale_locks/etc).",
	}, []string{"cleanup_type"})

	mediaServeRequests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "media_serve_requests_total",
		Help:      "Total media serve requests, labeled by instance_id, result (success/error), and error_type.",
	}, []string{"instance_id", "result", "error_type"})

	mediaServeBytes := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "media_serve_bytes_total",
		Help:      "Total bytes served via local media endpoints, labeled by instance_id.",
	}, []string{"instance_id"})

	mediaUploadAttempts := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "media_upload_attempts_total",
		Help:      "Total S3 upload attempts, labeled by status (success/failure).",
	}, []string{"status"})

	mediaUploadErrors := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "media_upload_errors_total",
		Help:      "Total S3 upload errors, labeled by error_type.",
	}, []string{"error_type"})

	mediaUploadSizeBytes := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "media_upload_size_bytes_total",
		Help:      "Total bytes uploaded to S3, labeled by media_type.",
	}, []string{"media_type"})

	mediaPresignedURLGenerated := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "media_presigned_url_generated_total",
		Help:      "Total presigned URLs generated for media.",
	})

	mediaDeleteAttempts := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "media_delete_attempts_total",
		Help:      "Total S3 deletion attempts, labeled by status (success/failure).",
	}, []string{"status"})

	transportDeliveries := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "transport_deliveries_total",
		Help:      "Total transport delivery attempts, labeled by instance_id, transport_type, and status.",
	}, []string{"instance_id", "transport_type", "status"})

	transportDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "transport_duration_seconds",
		Help:      "Duration of transport delivery in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"instance_id", "transport_type"})

	transportErrors := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "transport_errors_total",
		Help:      "Total transport delivery errors, labeled by instance_id, transport_type, and error_type.",
	}, []string{"instance_id", "transport_type", "error_type"})

	transportRetries := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "transport_retries_total",
		Help:      "Total transport retry attempts, labeled by instance_id, transport_type, and attempt.",
	}, []string{"instance_id", "transport_type", "attempt"})

	circuitBreakerStatePerInstance := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "circuit_breaker_state_per_instance",
		Help:      "Circuit breaker state per instance (0=CLOSED, 1=OPEN, 2=HALF_OPEN).",
	}, []string{"instance_id"})

	circuitBreakerTransitions := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "circuit_breaker_transitions_total",
		Help:      "Total circuit breaker state transitions, labeled by instance_id, from_state, and to_state.",
	}, []string{"instance_id", "from_state", "to_state"})

	workersActive := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "workers_active",
		Help:      "Number of active workers, labeled by worker_type.",
	}, []string{"worker_type"})

	workerTaskDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "worker_task_duration_seconds",
		Help:      "Duration of worker tasks in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"worker_type", "task_type"})

	workerErrors := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "worker_errors_total",
		Help:      "Total worker errors, labeled by worker_type and error_type.",
	}, []string{"worker_type", "error_type"})

	reg.MustRegister(
		requests, duration, queue,
		lockAcquisitions, lockReacquisitionAttempts, lockReacquisitionFallbacks,
		circuitBreakerState, splitBrainDetected, splitBrainInvalidLocks,
		healthChecks,
		eventsCaptured, eventsBuffered, eventsInserted, eventsProcessed,
		eventProcessingDuration, eventRetries, eventsFailed, eventsDelivered,
		eventDeliveryDuration, eventSequenceGaps, eventOutboxBacklog,
		dlqEventsTotal, dlqReprocessAttempts, dlqReprocessSuccess, dlqBacklog,
		mediaDownloadsTotal, mediaDownloadDuration, mediaDownloadSize, mediaDownloadErrors,
		mediaUploadsTotal, mediaUploadDuration, mediaUploadAttempts,
		mediaUploadErrors, mediaUploadSizeBytes, mediaPresignedURLGenerated,
		mediaDeleteAttempts, mediaFailures, mediaBacklog,
		mediaFallbackAttempts, mediaFallbackSuccess, mediaFallbackFailure,
		mediaLocalStorageSize, mediaLocalStorageFiles, mediaCleanupTotal,
		mediaServeRequests, mediaServeBytes,
		transportDeliveries, transportDuration, transportErrors, transportRetries,
		circuitBreakerStatePerInstance, circuitBreakerTransitions,
		workersActive, workerTaskDuration, workerErrors,
	)

	return &Metrics{
		HTTPRequests:                   requests,
		HTTPDuration:                   duration,
		WebhookQueue:                   queue,
		LockAcquisitions:               lockAcquisitions,
		LockReacquisitionAttempts:      lockReacquisitionAttempts,
		LockReacquisitionFallbacks:     lockReacquisitionFallbacks,
		CircuitBreakerState:            circuitBreakerState,
		SplitBrainDetected:             splitBrainDetected,
		SplitBrainInvalidLocks:         splitBrainInvalidLocks,
		HealthChecks:                   healthChecks,
		EventsCaptured:                 eventsCaptured,
		EventsBuffered:                 eventsBuffered,
		EventsInserted:                 eventsInserted,
		EventsProcessed:                eventsProcessed,
		EventProcessingDuration:        eventProcessingDuration,
		EventRetries:                   eventRetries,
		EventsFailed:                   eventsFailed,
		EventsDelivered:                eventsDelivered,
		EventDeliveryDuration:          eventDeliveryDuration,
		EventSequenceGaps:              eventSequenceGaps,
		EventOutboxBacklog:             eventOutboxBacklog,
		DLQEventsTotal:                 dlqEventsTotal,
		DLQReprocessAttempts:           dlqReprocessAttempts,
		DLQReprocessSuccess:            dlqReprocessSuccess,
		DLQBacklog:                     dlqBacklog,
		MediaDownloadsTotal:            mediaDownloadsTotal,
		MediaDownloadDuration:          mediaDownloadDuration,
		MediaDownloadSize:              mediaDownloadSize,
		MediaDownloadErrors:            mediaDownloadErrors,
		MediaUploadsTotal:              mediaUploadsTotal,
		MediaUploadDuration:            mediaUploadDuration,
		MediaUploadAttempts:            mediaUploadAttempts,
		MediaUploadErrors:              mediaUploadErrors,
		MediaUploadSizeBytes:           mediaUploadSizeBytes,
		MediaPresignedURLGenerated:     mediaPresignedURLGenerated,
		MediaDeleteAttempts:            mediaDeleteAttempts,
		MediaFailures:                  mediaFailures,
		MediaBacklog:                   mediaBacklog,
		MediaFallbackAttempts:          mediaFallbackAttempts,
		MediaFallbackSuccess:           mediaFallbackSuccess,
		MediaFallbackFailure:           mediaFallbackFailure,
		MediaLocalStorageSize:          mediaLocalStorageSize,
		MediaLocalStorageFiles:         mediaLocalStorageFiles,
		MediaCleanupTotal:              mediaCleanupTotal,
		MediaServeRequests:             mediaServeRequests,
		MediaServeBytes:                mediaServeBytes,
		TransportDeliveries:            transportDeliveries,
		TransportDuration:              transportDuration,
		TransportErrors:                transportErrors,
		TransportRetries:               transportRetries,
		CircuitBreakerStatePerInstance: circuitBreakerStatePerInstance,
		CircuitBreakerTransitions:      circuitBreakerTransitions,
		WorkersActive:                  workersActive,
		WorkerTaskDuration:             workerTaskDuration,
		WorkerErrors:                   workerErrors,
	}
}
