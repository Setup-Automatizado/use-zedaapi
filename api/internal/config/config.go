package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv string

	HTTP struct {
		Addr              string
		ReadHeaderTimeout time.Duration
		ReadTimeout       time.Duration
		WriteTimeout      time.Duration
		IdleTimeout       time.Duration
		MaxHeaderBytes    int
	}

	Log struct {
		Level string
	}

	Postgres struct {
		DSN      string
		MaxConns int32
	}

	WhatsmeowStore struct {
		DSN      string
		LogLevel string
	}

	Redis struct {
		Addr       string
		Username   string
		Password   string
		DB         int
		TLSEnabled bool
	}

	S3 struct {
		Endpoint  string
		Region    string
		Bucket    string
		AccessKey string
		SecretKey string
		UseSSL    bool
		URLExpiry time.Duration
		ACL       string // Optional: S3 object ACL (e.g., "public-read", "private"). Empty = use bucket policy (modern AWS pattern)
	}

	Media struct {
		LocalStoragePath   string
		LocalURLExpiry     time.Duration
		LocalSecretKey     string
		LocalPublicBaseURL string
	}

	Sentry struct {
		DSN         string
		Environment string
		Release     string
	}

	Workers struct {
		WebhookDispatcher int
		Media             int
	}

	Prometheus struct {
		Namespace string
	}

	Partner struct {
		AuthToken string
	}

	Events struct {
		// Event processing configuration
		BufferSize            int
		BatchSize             int
		PollInterval          time.Duration
		ProcessingTimeout     time.Duration
		ShutdownGracePeriod   time.Duration

		// Retry configuration
		MaxRetryAttempts      int
		RetryDelays           []time.Duration

		// Circuit breaker configuration
		CircuitBreakerEnabled bool
		CBMaxFailures         int
		CBTimeout             time.Duration
		CBCooldown            time.Duration

		// DLQ configuration
		DLQRetentionPeriod    time.Duration
		DLQReprocessEnabled   bool

		// Media processing configuration
		MediaBufferSize       int
		MediaBatchSize        int
		MediaMaxRetries       int
		MediaPollInterval     time.Duration
		MediaDownloadTimeout  time.Duration
		MediaUploadTimeout    time.Duration
		MediaMaxFileSize      int64
		MediaChunkSize        int64

		// Transport configuration
		WebhookTimeout        time.Duration
		WebhookMaxRetries     int
		TransportBufferSize   int

		// Cleanup configuration
		DeliveredRetention    time.Duration
		CleanupInterval       time.Duration
	}
}

func Load() (Config, error) {
	var cfg Config

	cfg.AppEnv = getEnv("APP_ENV", "development")

	httpReadHeaderTimeout, err := parseDuration(getEnv("HTTP_READ_HEADER_TIMEOUT", "5s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid HTTP_READ_HEADER_TIMEOUT: %w", err)
	}
	httpReadTimeout, err := parseDuration(getEnv("HTTP_READ_TIMEOUT", "15s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid HTTP_READ_TIMEOUT: %w", err)
	}
	httpWriteTimeout, err := parseDuration(getEnv("HTTP_WRITE_TIMEOUT", "30s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid HTTP_WRITE_TIMEOUT: %w", err)
	}
	httpIdleTimeout, err := parseDuration(getEnv("HTTP_IDLE_TIMEOUT", "120s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid HTTP_IDLE_TIMEOUT: %w", err)
	}
	maxHeaderBytes, err := parseInt(getEnv("HTTP_MAX_HEADER_BYTES", "1048576"))
	if err != nil {
		return cfg, fmt.Errorf("invalid HTTP_MAX_HEADER_BYTES: %w", err)
	}

	cfg.HTTP = struct {
		Addr              string
		ReadHeaderTimeout time.Duration
		ReadTimeout       time.Duration
		WriteTimeout      time.Duration
		IdleTimeout       time.Duration
		MaxHeaderBytes    int
	}{
		Addr:              getEnv("HTTP_ADDR", "0.0.0.0:8080"),
		ReadHeaderTimeout: httpReadHeaderTimeout,
		ReadTimeout:       httpReadTimeout,
		WriteTimeout:      httpWriteTimeout,
		IdleTimeout:       httpIdleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
	}

	cfg.Log.Level = getEnv("LOG_LEVEL", "INFO")

	maxConns, err := parseInt32(getEnv("POSTGRES_MAX_CONNS", "32"))
	if err != nil {
		return cfg, fmt.Errorf("invalid POSTGRES_MAX_CONNS: %w", err)
	}
	cfg.Postgres = struct {
		DSN      string
		MaxConns int32
	}{
		DSN:      getEnv("POSTGRES_DSN", "postgres://funnelchat:funnelchat@localhost:5432/funnelchat_api?sslmode=disable"),
		MaxConns: maxConns,
	}

	cfg.WhatsmeowStore = struct {
		DSN      string
		LogLevel string
	}{
		DSN:      getEnv("WAMEOW_POSTGRES_DSN", "postgres://funnelchat:funnelchat@localhost:5432/funnelchat_store?sslmode=disable"),
		LogLevel: getEnv("WAMEOW_LOG_LEVEL", "INFO"),
	}

	redisDB, err := parseInt(getEnv("REDIS_DB", "0"))
	if err != nil {
		return cfg, fmt.Errorf("invalid REDIS_DB: %w", err)
	}
	cfg.Redis = struct {
		Addr       string
		Username   string
		Password   string
		DB         int
		TLSEnabled bool
	}{
		Addr:       getEnv("REDIS_ADDR", "localhost:6379"),
		Username:   os.Getenv("REDIS_USERNAME"),
		Password:   os.Getenv("REDIS_PASSWORD"),
		DB:         redisDB,
		TLSEnabled: parseBool(getEnv("REDIS_TLS_ENABLED", "false")),
	}

	cfg.S3 = struct {
		Endpoint  string
		Region    string
		Bucket    string
		AccessKey string
		SecretKey string
		UseSSL    bool
		URLExpiry time.Duration
		ACL       string
	}{
		Endpoint:  getEnv("S3_ENDPOINT", "http://localhost:9000"),
		Region:    getEnv("S3_REGION", "us-east-1"),
		Bucket:    getEnv("S3_BUCKET", "funnelchat-media"),
		AccessKey: os.Getenv("S3_ACCESS_KEY"),
		SecretKey: os.Getenv("S3_SECRET_KEY"),
		UseSSL:    parseBool(getEnv("S3_USE_SSL", "false")),
		ACL:       getEnv("S3_ACL", ""), // Empty = use bucket policy (modern pattern)
	}
	expiry, err := parseDuration(getEnv("S3_URL_EXPIRATION", "30d"))
	if err != nil {
		return cfg, fmt.Errorf("invalid S3_URL_EXPIRATION: %w", err)
	}
	cfg.S3.URLExpiry = expiry

	// Media configuration
	cfg.Media = struct {
		LocalStoragePath   string
		LocalURLExpiry     time.Duration
		LocalSecretKey     string
		LocalPublicBaseURL string
	}{
		LocalStoragePath:   getEnv("MEDIA_LOCAL_STORAGE_PATH", "/var/whatsmeow/media"),
		LocalSecretKey:     os.Getenv("MEDIA_LOCAL_SECRET_KEY"),
		LocalPublicBaseURL: os.Getenv("MEDIA_LOCAL_PUBLIC_BASE_URL"),
	}
	mediaExpiry, err := parseDuration(getEnv("MEDIA_LOCAL_URL_EXPIRY", "720h"))
	if err != nil {
		return cfg, fmt.Errorf("invalid MEDIA_LOCAL_URL_EXPIRY: %w", err)
	}
	cfg.Media.LocalURLExpiry = mediaExpiry

	cfg.Sentry = struct {
		DSN         string
		Environment string
		Release     string
	}{
		DSN:         os.Getenv("SENTRY_DSN"),
		Environment: getEnv("SENTRY_ENVIRONMENT", cfg.AppEnv),
		Release:     getEnv("SENTRY_RELEASE", "dev"),
	}

	cfg.Workers = struct {
		WebhookDispatcher int
		Media             int
	}{
		WebhookDispatcher: mustParsePositiveInt(getEnv("WEBHOOK_DISPATCHER_CONCURRENCY", "8")),
		Media:             mustParsePositiveInt(getEnv("MEDIA_WORKER_CONCURRENCY", "4")),
	}

	cfg.Prometheus.Namespace = getEnv("PROMETHEUS_NAMESPACE", "funnelchat_api")

	cfg.Partner.AuthToken = os.Getenv("PARTNER_AUTH_TOKEN")

	// Event system configuration
	eventBufferSize := mustParsePositiveInt(getEnv("EVENT_BUFFER_SIZE", "1000"))
	eventBatchSize := mustParsePositiveInt(getEnv("EVENT_BATCH_SIZE", "10"))
	eventPollInterval, err := parseDuration(getEnv("EVENT_POLL_INTERVAL", "100ms"))
	if err != nil {
		return cfg, fmt.Errorf("invalid EVENT_POLL_INTERVAL: %w", err)
	}
	eventProcessingTimeout, err := parseDuration(getEnv("EVENT_PROCESSING_TIMEOUT", "30s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid EVENT_PROCESSING_TIMEOUT: %w", err)
	}
	eventShutdownGracePeriod, err := parseDuration(getEnv("EVENT_SHUTDOWN_GRACE_PERIOD", "30s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid EVENT_SHUTDOWN_GRACE_PERIOD: %w", err)
	}

	// Retry configuration
	maxRetryAttempts := mustParsePositiveInt(getEnv("EVENT_MAX_RETRY_ATTEMPTS", "6"))
	retryDelaysStr := getEnv("EVENT_RETRY_DELAYS", "0s,10s,30s,2m,5m,15m")
	retryDelays, err := parseRetryDelays(retryDelaysStr)
	if err != nil {
		return cfg, fmt.Errorf("invalid EVENT_RETRY_DELAYS: %w", err)
	}

	// Circuit breaker configuration
	cbMaxFailures := mustParsePositiveInt(getEnv("CB_MAX_FAILURES", "5"))
	cbTimeout, err := parseDuration(getEnv("CB_TIMEOUT", "60s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid CB_TIMEOUT: %w", err)
	}
	cbCooldown, err := parseDuration(getEnv("CB_COOLDOWN", "30s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid CB_COOLDOWN: %w", err)
	}

	// DLQ configuration
	dlqRetention, err := parseDuration(getEnv("DLQ_RETENTION_PERIOD", "7d"))
	if err != nil {
		return cfg, fmt.Errorf("invalid DLQ_RETENTION_PERIOD: %w", err)
	}

	// Media processing configuration
	mediaBufferSize := mustParsePositiveInt(getEnv("MEDIA_BUFFER_SIZE", "500"))
	mediaBatchSize := mustParsePositiveInt(getEnv("MEDIA_BATCH_SIZE", "5"))
	mediaMaxRetries := mustParsePositiveInt(getEnv("MEDIA_MAX_RETRIES", "3"))
	mediaPollInterval, err := parseDuration(getEnv("MEDIA_POLL_INTERVAL", "1s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid MEDIA_POLL_INTERVAL: %w", err)
	}
	mediaDownloadTimeout, err := parseDuration(getEnv("MEDIA_DOWNLOAD_TIMEOUT", "5m"))
	if err != nil {
		return cfg, fmt.Errorf("invalid MEDIA_DOWNLOAD_TIMEOUT: %w", err)
	}
	mediaUploadTimeout, err := parseDuration(getEnv("MEDIA_UPLOAD_TIMEOUT", "10m"))
	if err != nil {
		return cfg, fmt.Errorf("invalid MEDIA_UPLOAD_TIMEOUT: %w", err)
	}
	mediaMaxFileSize, err := parseInt64(getEnv("MEDIA_MAX_FILE_SIZE", "104857600")) // 100MB default
	if err != nil {
		return cfg, fmt.Errorf("invalid MEDIA_MAX_FILE_SIZE: %w", err)
	}
	mediaChunkSize, err := parseInt64(getEnv("MEDIA_CHUNK_SIZE", "5242880")) // 5MB default
	if err != nil {
		return cfg, fmt.Errorf("invalid MEDIA_CHUNK_SIZE: %w", err)
	}

	// Transport configuration
	webhookTimeout, err := parseDuration(getEnv("WEBHOOK_TIMEOUT", "30s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid WEBHOOK_TIMEOUT: %w", err)
	}
	webhookMaxRetries := mustParsePositiveInt(getEnv("WEBHOOK_MAX_RETRIES", "3"))
	transportBufferSize := mustParsePositiveInt(getEnv("TRANSPORT_BUFFER_SIZE", "100"))

	// Cleanup configuration
	deliveredRetention, err := parseDuration(getEnv("DELIVERED_RETENTION_PERIOD", "1d"))
	if err != nil {
		return cfg, fmt.Errorf("invalid DELIVERED_RETENTION_PERIOD: %w", err)
	}
	cleanupInterval, err := parseDuration(getEnv("CLEANUP_INTERVAL", "1h"))
	if err != nil {
		return cfg, fmt.Errorf("invalid CLEANUP_INTERVAL: %w", err)
	}

	cfg.Events = struct {
		// Event processing configuration
		BufferSize            int
		BatchSize             int
		PollInterval          time.Duration
		ProcessingTimeout     time.Duration
		ShutdownGracePeriod   time.Duration

		// Retry configuration
		MaxRetryAttempts      int
		RetryDelays           []time.Duration

		// Circuit breaker configuration
		CircuitBreakerEnabled bool
		CBMaxFailures         int
		CBTimeout             time.Duration
		CBCooldown            time.Duration

		// DLQ configuration
		DLQRetentionPeriod    time.Duration
		DLQReprocessEnabled   bool

		// Media processing configuration
		MediaBufferSize       int
		MediaBatchSize        int
		MediaMaxRetries       int
		MediaPollInterval     time.Duration
		MediaDownloadTimeout  time.Duration
		MediaUploadTimeout    time.Duration
		MediaMaxFileSize      int64
		MediaChunkSize        int64

		// Transport configuration
		WebhookTimeout        time.Duration
		WebhookMaxRetries     int
		TransportBufferSize   int

		// Cleanup configuration
		DeliveredRetention    time.Duration
		CleanupInterval       time.Duration
	}{
		BufferSize:            eventBufferSize,
		BatchSize:             eventBatchSize,
		PollInterval:          eventPollInterval,
		ProcessingTimeout:     eventProcessingTimeout,
		ShutdownGracePeriod:   eventShutdownGracePeriod,
		MaxRetryAttempts:      maxRetryAttempts,
		RetryDelays:           retryDelays,
		CircuitBreakerEnabled: parseBool(getEnv("CB_ENABLED", "true")),
		CBMaxFailures:         cbMaxFailures,
		CBTimeout:             cbTimeout,
		CBCooldown:            cbCooldown,
		DLQRetentionPeriod:    dlqRetention,
		DLQReprocessEnabled:   parseBool(getEnv("DLQ_REPROCESS_ENABLED", "true")),
		MediaBufferSize:       mediaBufferSize,
		MediaBatchSize:        mediaBatchSize,
		MediaMaxRetries:       mediaMaxRetries,
		MediaPollInterval:     mediaPollInterval,
		MediaDownloadTimeout:  mediaDownloadTimeout,
		MediaUploadTimeout:    mediaUploadTimeout,
		MediaMaxFileSize:      mediaMaxFileSize,
		MediaChunkSize:        mediaChunkSize,
		WebhookTimeout:        webhookTimeout,
		WebhookMaxRetries:     webhookMaxRetries,
		TransportBufferSize:   transportBufferSize,
		DeliveredRetention:    deliveredRetention,
		CleanupInterval:       cleanupInterval,
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		return val
	}
	return fallback
}

func parseDuration(val string) (time.Duration, error) {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return 0, nil
	}
	if strings.HasSuffix(trimmed, "d") {
		daysStr := strings.TrimSuffix(trimmed, "d")
		days, err := strconv.ParseFloat(daysStr, 64)
		if err != nil {
			return 0, err
		}
		d := time.Duration(days * 24 * float64(time.Hour))
		return d, nil
	}
	if strings.HasSuffix(trimmed, "w") {
		weeksStr := strings.TrimSuffix(trimmed, "w")
		weeks, err := strconv.ParseFloat(weeksStr, 64)
		if err != nil {
			return 0, err
		}
		d := time.Duration(weeks * 7 * 24 * float64(time.Hour))
		return d, nil
	}
	return time.ParseDuration(trimmed)
}

func parseInt(val string) (int, error) {
	i, err := strconv.Atoi(strings.TrimSpace(val))
	if err != nil {
		return 0, err
	}
	return i, nil
}

func parseInt32(val string) (int32, error) {
	parsed, err := parseInt(val)
	if err != nil {
		return 0, err
	}
	return int32(parsed), nil
}

func parseBool(val string) bool {
	b, err := strconv.ParseBool(strings.TrimSpace(val))
	if err != nil {
		return false
	}
	return b
}

func mustParsePositiveInt(val string) int {
	parsed, err := parseInt(val)
	if err != nil {
		return 1
	}
	if parsed <= 0 {
		return 1
	}
	return parsed
}

func parseInt64(val string) (int64, error) {
	i, err := strconv.ParseInt(strings.TrimSpace(val), 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func parseRetryDelays(val string) ([]time.Duration, error) {
	parts := strings.Split(val, ",")
	delays := make([]time.Duration, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		d, err := parseDuration(trimmed)
		if err != nil {
			return nil, fmt.Errorf("invalid duration %q: %w", trimmed, err)
		}
		delays = append(delays, d)
	}

	if len(delays) == 0 {
		return []time.Duration{0}, nil
	}

	return delays, nil
}
