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
		BaseURL           string
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

	RedisLock struct {
		KeyPrefix       string
		TTL             time.Duration
		RefreshInterval time.Duration
	}

	S3 struct {
		Endpoint         string
		Region           string
		Bucket           string
		AccessKey        string
		SecretKey        string
		UseSSL           bool
		UsePresignedURLs bool
		PublicBaseURL    string
		URLExpiry        time.Duration
		ACL              string // Optional: S3 object ACL (e.g., "public-read", "private"). Empty = use bucket policy (modern AWS pattern)
	}

	Media struct {
		LocalStoragePath   string
		LocalURLExpiry     time.Duration
		LocalSecretKey     string
		LocalPublicBaseURL string
		CleanupInterval    time.Duration
		CleanupBatchSize   int
		S3Retention        time.Duration
		LocalRetention     time.Duration
	}

	Document struct {
		MuPDFVersion string
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

	WorkerRegistry struct {
		HeartbeatInterval time.Duration
		Expiry            time.Duration
		RebalanceInterval time.Duration
	}

	Prometheus struct {
		Namespace string
	}

	Partner struct {
		AuthToken string
	}

	Client struct {
		AuthToken string
	}

	ContactMetadata struct {
		CacheCapacity   int
		NameTTL         time.Duration
		PhotoTTL        time.Duration
		ErrorTTL        time.Duration
		PrefetchWorkers int
		FetchQueueSize  int
	}

	Events struct {
		BufferSize          int
		BatchSize           int
		PollInterval        time.Duration
		ProcessingTimeout   time.Duration
		HandlerTimeout      time.Duration
		ShutdownGracePeriod time.Duration

		MaxRetryAttempts int
		RetryDelays      []time.Duration

		CircuitBreakerEnabled bool
		CBMaxFailures         int
		CBTimeout             time.Duration
		CBCooldown            time.Duration

		DLQRetentionPeriod  time.Duration
		DLQReprocessEnabled bool

		MediaBufferSize      int
		MediaBatchSize       int
		MediaMaxRetries      int
		MediaPollInterval    time.Duration
		MediaDownloadTimeout time.Duration
		MediaUploadTimeout   time.Duration
		MediaMaxFileSize     int64
		MediaChunkSize       int64

		WebhookTimeout      time.Duration
		WebhookMaxRetries   int
		TransportBufferSize int

		DeliveredRetention time.Duration
		CleanupInterval    time.Duration

		DebugRawPayload bool
		DebugDumpDir    string
	}

	MessageQueue struct {
		Enabled              bool
		PollInterval         time.Duration
		MaxAttempts          int
		InitialBackoff       time.Duration
		MaxBackoff           time.Duration
		BackoffMultiplier    float64
		DisconnectRetryDelay time.Duration
		CompletedRetention   time.Duration
		FailedRetention      time.Duration
		CleanupInterval      time.Duration
		CleanupTimeout       time.Duration
	}

	Reconciliation struct {
		Enabled  bool
		Interval time.Duration
	}

	Connection struct {
		AutoConnectPaired bool
		MaxAttempts       int
		InitialBackoff    time.Duration
		MaxBackoff        time.Duration
	}

	Shutdown struct {
		OverallTimeout          time.Duration
		HTTPTimeout             time.Duration
		QueueDrainTimeout       time.Duration
		EventFlushTimeout       time.Duration
		ClientDisconnectTimeout time.Duration
		LockReleaseTimeout      time.Duration
	}

	StatusCache struct {
		Enabled          bool
		TTL              time.Duration
		Types            []string // read, delivered, played, sent
		Scope            []string // groups, direct
		SuppressWebhooks bool
		FlushBatchSize   int
		CleanupInterval  time.Duration
		OperationTimeout time.Duration
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
		BaseURL           string
		ReadHeaderTimeout time.Duration
		ReadTimeout       time.Duration
		WriteTimeout      time.Duration
		IdleTimeout       time.Duration
		MaxHeaderBytes    int
	}{
		Addr:              getEnv("HTTP_ADDR", "0.0.0.0:8080"),
		BaseURL:           getEnv("API_BASE_URL", "http://localhost:8080"),
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

	lockTTL, err := parseDuration(getEnv("REDIS_LOCK_TTL", "30s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid REDIS_LOCK_TTL: %w", err)
	}
	if lockTTL <= 0 {
		lockTTL = 30 * time.Second
	}
	lockRefresh, err := parseDuration(getEnv("REDIS_LOCK_REFRESH_INTERVAL", "10s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid REDIS_LOCK_REFRESH_INTERVAL: %w", err)
	}
	if lockRefresh <= 0 || lockRefresh >= lockTTL {
		lockRefresh = lockTTL / 2
	}
	cfg.RedisLock = struct {
		KeyPrefix       string
		TTL             time.Duration
		RefreshInterval time.Duration
	}{
		KeyPrefix:       getEnv("REDIS_LOCK_KEY_PREFIX", "funnelchat"),
		TTL:             lockTTL,
		RefreshInterval: lockRefresh,
	}

	cfg.S3 = struct {
		Endpoint         string
		Region           string
		Bucket           string
		AccessKey        string
		SecretKey        string
		UseSSL           bool
		UsePresignedURLs bool
		PublicBaseURL    string
		URLExpiry        time.Duration
		ACL              string
	}{
		Endpoint:         getEnv("S3_ENDPOINT", "http://localhost:9000"),
		Region:           getEnv("S3_REGION", "us-east-1"),
		Bucket:           getEnv("S3_BUCKET", "funnelchat-media"),
		AccessKey:        os.Getenv("S3_ACCESS_KEY"),
		SecretKey:        os.Getenv("S3_SECRET_KEY"),
		UseSSL:           parseBool(getEnv("S3_USE_SSL", "false")),
		UsePresignedURLs: parseBool(getEnv("S3_USE_PRESIGNED_URLS", "true")),
		PublicBaseURL:    os.Getenv("S3_PUBLIC_BASE_URL"),
		ACL:              getEnv("S3_ACL", ""),
	}
	expiry, err := parseDuration(getEnv("S3_URL_EXPIRATION", "6d"))
	if err != nil {
		return cfg, fmt.Errorf("invalid S3_URL_EXPIRATION: %w", err)
	}
	cfg.S3.URLExpiry = expiry

	cfg.Media.LocalStoragePath = getEnv("MEDIA_LOCAL_STORAGE_PATH", "/var/whatsmeow/media")
	cfg.Media.LocalSecretKey = os.Getenv("MEDIA_LOCAL_SECRET_KEY")
	cfg.Media.LocalPublicBaseURL = os.Getenv("MEDIA_LOCAL_PUBLIC_BASE_URL")
	mediaExpiry, err := parseDuration(getEnv("MEDIA_LOCAL_URL_EXPIRY", "720h"))
	if err != nil {
		return cfg, fmt.Errorf("invalid MEDIA_LOCAL_URL_EXPIRY: %w", err)
	}
	cfg.Media.LocalURLExpiry = mediaExpiry
	mediaCleanupInterval, err := parseDuration(getEnv("MEDIA_CLEANUP_INTERVAL", "168h"))
	if err != nil {
		return cfg, fmt.Errorf("invalid MEDIA_CLEANUP_INTERVAL: %w", err)
	}
	cfg.Media.CleanupInterval = mediaCleanupInterval
	cleanupBatchSize, err := parseInt(getEnv("MEDIA_CLEANUP_BATCH_SIZE", "200"))
	if err != nil {
		return cfg, fmt.Errorf("invalid MEDIA_CLEANUP_BATCH_SIZE: %w", err)
	}
	cfg.Media.CleanupBatchSize = cleanupBatchSize
	if cfg.Media.CleanupBatchSize <= 0 {
		cfg.Media.CleanupBatchSize = 200
	}
	s3Retention, err := parseDuration(getEnv("S3_MEDIA_RETENTION", "720h"))
	if err != nil {
		return cfg, fmt.Errorf("invalid S3_MEDIA_RETENTION: %w", err)
	}
	cfg.Media.S3Retention = s3Retention
	localRetention, err := parseDuration(getEnv("LOCAL_MEDIA_RETENTION", "720h"))
	if err != nil {
		return cfg, fmt.Errorf("invalid LOCAL_MEDIA_RETENTION: %w", err)
	}
	cfg.Media.LocalRetention = localRetention

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

	heartbeatInterval, err := parseDuration(getEnv("WORKER_HEARTBEAT_INTERVAL", "5s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid WORKER_HEARTBEAT_INTERVAL: %w", err)
	}
	workerExpiry, err := parseDuration(getEnv("WORKER_HEARTBEAT_EXPIRY", "20s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid WORKER_HEARTBEAT_EXPIRY: %w", err)
	}
	if workerExpiry <= heartbeatInterval {
		workerExpiry = heartbeatInterval * 2
	}
	rebalanceInterval, err := parseDuration(getEnv("WORKER_REBALANCE_INTERVAL", "30s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid WORKER_REBALANCE_INTERVAL: %w", err)
	}
	if rebalanceInterval <= 0 {
		rebalanceInterval = 30 * time.Second
	}
	cfg.WorkerRegistry = struct {
		HeartbeatInterval time.Duration
		Expiry            time.Duration
		RebalanceInterval time.Duration
	}{
		HeartbeatInterval: heartbeatInterval,
		Expiry:            workerExpiry,
		RebalanceInterval: rebalanceInterval,
	}

	cfg.Prometheus.Namespace = getEnv("PROMETHEUS_NAMESPACE", "funnelchat_api")

	cfg.Partner.AuthToken = strings.TrimSpace(os.Getenv("PARTNER_AUTH_TOKEN"))
	cfg.Client.AuthToken = strings.TrimSpace(os.Getenv("CLIENT_AUTH_TOKEN"))

	if cfg.Client.AuthToken == "" {
		return cfg, fmt.Errorf("CLIENT_AUTH_TOKEN must be configured")
	}
	if len(cfg.Client.AuthToken) < 16 {
		return cfg, fmt.Errorf("CLIENT_AUTH_TOKEN must be at least 16 characters")
	}

	contactCacheCapacity := mustParsePositiveInt(getEnv("CONTACT_METADATA_CACHE_CAPACITY", "50000"))
	contactNameTTL, err := parseDuration(getEnv("CONTACT_METADATA_NAME_TTL", "24h"))
	if err != nil {
		return cfg, fmt.Errorf("invalid CONTACT_METADATA_NAME_TTL: %w", err)
	}
	contactPhotoTTL, err := parseDuration(getEnv("CONTACT_METADATA_PHOTO_TTL", "24h"))
	if err != nil {
		return cfg, fmt.Errorf("invalid CONTACT_METADATA_PHOTO_TTL: %w", err)
	}
	contactErrorTTL, err := parseDuration(getEnv("CONTACT_METADATA_ERROR_TTL", "24h"))
	if err != nil {
		return cfg, fmt.Errorf("invalid CONTACT_METADATA_ERROR_TTL: %w", err)
	}
	contactPrefetchWorkers := mustParsePositiveInt(getEnv("CONTACT_METADATA_PREFETCH_WORKERS", "4"))
	contactFetchQueueSize := mustParsePositiveInt(getEnv("CONTACT_METADATA_FETCH_QUEUE_SIZE", "1024"))

	cfg.ContactMetadata = struct {
		CacheCapacity   int
		NameTTL         time.Duration
		PhotoTTL        time.Duration
		ErrorTTL        time.Duration
		PrefetchWorkers int
		FetchQueueSize  int
	}{
		CacheCapacity:   contactCacheCapacity,
		NameTTL:         contactNameTTL,
		PhotoTTL:        contactPhotoTTL,
		ErrorTTL:        contactErrorTTL,
		PrefetchWorkers: contactPrefetchWorkers,
		FetchQueueSize:  contactFetchQueueSize,
	}

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
	eventHandlerTimeout, err := parseDuration(getEnv("EVENT_HANDLER_TIMEOUT", "60s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid EVENT_HANDLER_TIMEOUT: %w", err)
	}
	eventShutdownGracePeriod, err := parseDuration(getEnv("EVENT_SHUTDOWN_GRACE_PERIOD", "30s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid EVENT_SHUTDOWN_GRACE_PERIOD: %w", err)
	}

	maxRetryAttempts := mustParsePositiveInt(getEnv("EVENT_MAX_RETRY_ATTEMPTS", "6"))
	retryDelaysStr := getEnv("EVENT_RETRY_DELAYS", "0s,10s,30s,2m,5m,15m")
	retryDelays, err := parseRetryDelays(retryDelaysStr)
	if err != nil {
		return cfg, fmt.Errorf("invalid EVENT_RETRY_DELAYS: %w", err)
	}

	cbMaxFailures := mustParsePositiveInt(getEnv("CB_MAX_FAILURES", "5"))
	cbTimeout, err := parseDuration(getEnv("CB_TIMEOUT", "60s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid CB_TIMEOUT: %w", err)
	}
	cbCooldown, err := parseDuration(getEnv("CB_COOLDOWN", "30s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid CB_COOLDOWN: %w", err)
	}

	dlqRetention, err := parseDuration(getEnv("DLQ_RETENTION_PERIOD", "7d"))
	if err != nil {
		return cfg, fmt.Errorf("invalid DLQ_RETENTION_PERIOD: %w", err)
	}

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

	webhookTimeout, err := parseDuration(getEnv("WEBHOOK_TIMEOUT", "30s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid WEBHOOK_TIMEOUT: %w", err)
	}
	webhookMaxRetries := mustParsePositiveInt(getEnv("WEBHOOK_MAX_RETRIES", "3"))
	transportBufferSize := mustParsePositiveInt(getEnv("TRANSPORT_BUFFER_SIZE", "100"))

	deliveredRetention, err := parseDuration(getEnv("DELIVERED_RETENTION_PERIOD", "1d"))
	if err != nil {
		return cfg, fmt.Errorf("invalid DELIVERED_RETENTION_PERIOD: %w", err)
	}
	eventCleanupInterval, err := parseDuration(getEnv("CLEANUP_INTERVAL", "1h"))
	if err != nil {
		return cfg, fmt.Errorf("invalid CLEANUP_INTERVAL: %w", err)
	}

	debugRawPayload := parseBool(getEnv("EVENTS_DEBUG_RAW_PAYLOAD", "false"))
	debugDumpDir := strings.TrimSpace(getEnv("EVENTS_DEBUG_DUMP_DIR", "./tmp/debug-events"))
	if debugDumpDir == "" {
		debugDumpDir = "./tmp/debug-events"
	}

	// Message Queue Configuration
	mqEnabled := parseBool(getEnv("MESSAGE_QUEUE_ENABLED", "true"))
	mqPollInterval, err := parseDuration(getEnv("MESSAGE_QUEUE_POLL_INTERVAL", "100ms"))
	if err != nil {
		return cfg, fmt.Errorf("invalid MESSAGE_QUEUE_POLL_INTERVAL: %w", err)
	}
	mqMaxAttempts := mustParsePositiveInt(getEnv("MESSAGE_QUEUE_MAX_ATTEMPTS", "3"))
	mqInitialBackoff, err := parseDuration(getEnv("MESSAGE_QUEUE_INITIAL_BACKOFF", "1s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid MESSAGE_QUEUE_INITIAL_BACKOFF: %w", err)
	}
	mqMaxBackoff, err := parseDuration(getEnv("MESSAGE_QUEUE_MAX_BACKOFF", "5m"))
	if err != nil {
		return cfg, fmt.Errorf("invalid MESSAGE_QUEUE_MAX_BACKOFF: %w", err)
	}
	mqBackoffMultiplier := 2.0 // Fixed exponential backoff multiplier
	if val := getEnv("MESSAGE_QUEUE_BACKOFF_MULTIPLIER", ""); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil && parsed > 0 {
			mqBackoffMultiplier = parsed
		}
	}
	mqDisconnectRetryDelay, err := parseDuration(getEnv("MESSAGE_QUEUE_DISCONNECT_RETRY_DELAY", "30s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid MESSAGE_QUEUE_DISCONNECT_RETRY_DELAY: %w", err)
	}
	mqCompletedRetention, err := parseDuration(getEnv("MESSAGE_QUEUE_COMPLETED_RETENTION", "24h"))
	if err != nil {
		return cfg, fmt.Errorf("invalid MESSAGE_QUEUE_COMPLETED_RETENTION: %w", err)
	}
	mqFailedRetention, err := parseDuration(getEnv("MESSAGE_QUEUE_FAILED_RETENTION", "7d"))
	if err != nil {
		return cfg, fmt.Errorf("invalid MESSAGE_QUEUE_FAILED_RETENTION: %w", err)
	}
	mqCleanupInterval, err := parseDuration(getEnv("MESSAGE_QUEUE_CLEANUP_INTERVAL", "1h"))
	if err != nil {
		return cfg, fmt.Errorf("invalid MESSAGE_QUEUE_CLEANUP_INTERVAL: %w", err)
	}
	mqCleanupTimeout, err := parseDuration(getEnv("MESSAGE_QUEUE_CLEANUP_TIMEOUT", "5m"))
	if err != nil {
		return cfg, fmt.Errorf("invalid MESSAGE_QUEUE_CLEANUP_TIMEOUT: %w", err)
	}

	cfg.MessageQueue = struct {
		Enabled              bool
		PollInterval         time.Duration
		MaxAttempts          int
		InitialBackoff       time.Duration
		MaxBackoff           time.Duration
		BackoffMultiplier    float64
		DisconnectRetryDelay time.Duration
		CompletedRetention   time.Duration
		FailedRetention      time.Duration
		CleanupInterval      time.Duration
		CleanupTimeout       time.Duration
	}{
		Enabled:              mqEnabled,
		PollInterval:         mqPollInterval,
		MaxAttempts:          mqMaxAttempts,
		InitialBackoff:       mqInitialBackoff,
		MaxBackoff:           mqMaxBackoff,
		BackoffMultiplier:    mqBackoffMultiplier,
		DisconnectRetryDelay: mqDisconnectRetryDelay,
		CompletedRetention:   mqCompletedRetention,
		FailedRetention:      mqFailedRetention,
		CleanupInterval:      mqCleanupInterval,
		CleanupTimeout:       mqCleanupTimeout,
	}

	reconciliationEnabled := parseBool(getEnv("RECONCILIATION_ENABLED", "true"))
	reconciliationInterval, err := parseDuration(getEnv("RECONCILIATION_INTERVAL", "10s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid RECONCILIATION_INTERVAL: %w", err)
	}
	cfg.Reconciliation = struct {
		Enabled  bool
		Interval time.Duration
	}{
		Enabled:  reconciliationEnabled,
		Interval: reconciliationInterval,
	}

	autoConnectPaired := parseBool(getEnv("AUTO_CONNECT_PAIRED", "true"))
	autoConnectMaxAttempts := mustParsePositiveInt(getEnv("AUTO_CONNECT_MAX_ATTEMPTS", "3"))
	autoConnectInitialBackoff, err := parseDuration(getEnv("AUTO_CONNECT_INITIAL_BACKOFF", "1s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid AUTO_CONNECT_INITIAL_BACKOFF: %w", err)
	}
	autoConnectMaxBackoff, err := parseDuration(getEnv("AUTO_CONNECT_MAX_BACKOFF", "10s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid AUTO_CONNECT_MAX_BACKOFF: %w", err)
	}
	cfg.Connection = struct {
		AutoConnectPaired bool
		MaxAttempts       int
		InitialBackoff    time.Duration
		MaxBackoff        time.Duration
	}{
		AutoConnectPaired: autoConnectPaired,
		MaxAttempts:       autoConnectMaxAttempts,
		InitialBackoff:    autoConnectInitialBackoff,
		MaxBackoff:        autoConnectMaxBackoff,
	}

	shutdownOverallTimeout, err := parseDuration(getEnv("SHUTDOWN_TIMEOUT", "120s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid SHUTDOWN_TIMEOUT: %w", err)
	}
	shutdownHTTPTimeout, err := parseDuration(getEnv("SHUTDOWN_HTTP_TIMEOUT", "30s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid SHUTDOWN_HTTP_TIMEOUT: %w", err)
	}
	shutdownQueueTimeout, err := parseDuration(getEnv("SHUTDOWN_QUEUE_TIMEOUT", "60s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid SHUTDOWN_QUEUE_TIMEOUT: %w", err)
	}
	shutdownEventTimeout, err := parseDuration(getEnv("SHUTDOWN_EVENT_TIMEOUT", "10s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid SHUTDOWN_EVENT_TIMEOUT: %w", err)
	}
	shutdownClientTimeout, err := parseDuration(getEnv("SHUTDOWN_CLIENT_TIMEOUT", "10s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid SHUTDOWN_CLIENT_TIMEOUT: %w", err)
	}
	shutdownLockTimeout, err := parseDuration(getEnv("SHUTDOWN_LOCK_TIMEOUT", "5s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid SHUTDOWN_LOCK_TIMEOUT: %w", err)
	}
	cfg.Shutdown = struct {
		OverallTimeout          time.Duration
		HTTPTimeout             time.Duration
		QueueDrainTimeout       time.Duration
		EventFlushTimeout       time.Duration
		ClientDisconnectTimeout time.Duration
		LockReleaseTimeout      time.Duration
	}{
		OverallTimeout:          shutdownOverallTimeout,
		HTTPTimeout:             shutdownHTTPTimeout,
		QueueDrainTimeout:       shutdownQueueTimeout,
		EventFlushTimeout:       shutdownEventTimeout,
		ClientDisconnectTimeout: shutdownClientTimeout,
		LockReleaseTimeout:      shutdownLockTimeout,
	}

	cfg.Document = struct {
		MuPDFVersion string
	}{
		MuPDFVersion: getEnv("MUPDF_VERSION", "1.24.10"),
	}

	// Status Cache Configuration
	statusCacheEnabled := parseBool(getEnv("STATUS_CACHE_ENABLED", "false"))
	statusCacheTTL, err := parseDuration(getEnv("STATUS_CACHE_TTL", "24h"))
	if err != nil {
		return cfg, fmt.Errorf("invalid STATUS_CACHE_TTL: %w", err)
	}
	statusCacheTypesStr := getEnv("STATUS_CACHE_TYPES", "read,delivered,played,sent")
	statusCacheTypes := parseStringSlice(statusCacheTypesStr)
	statusCacheScopeStr := getEnv("STATUS_CACHE_SCOPE", "groups")
	statusCacheScope := parseStringSlice(statusCacheScopeStr)
	statusCacheSuppressWebhooks := parseBool(getEnv("STATUS_CACHE_SUPPRESS_WEBHOOKS", "true"))
	statusCacheFlushBatchSize := mustParsePositiveInt(getEnv("STATUS_CACHE_FLUSH_BATCH_SIZE", "100"))
	statusCacheCleanupInterval, err := parseDuration(getEnv("STATUS_CACHE_CLEANUP_INTERVAL", "1h"))
	if err != nil {
		return cfg, fmt.Errorf("invalid STATUS_CACHE_CLEANUP_INTERVAL: %w", err)
	}
	statusCacheOperationTimeout, err := parseDuration(getEnv("STATUS_CACHE_OPERATION_TIMEOUT", "10s"))
	if err != nil {
		return cfg, fmt.Errorf("invalid STATUS_CACHE_OPERATION_TIMEOUT: %w", err)
	}

	cfg.StatusCache = struct {
		Enabled          bool
		TTL              time.Duration
		Types            []string
		Scope            []string
		SuppressWebhooks bool
		FlushBatchSize   int
		CleanupInterval  time.Duration
		OperationTimeout time.Duration
	}{
		Enabled:          statusCacheEnabled,
		TTL:              statusCacheTTL,
		Types:            statusCacheTypes,
		Scope:            statusCacheScope,
		SuppressWebhooks: statusCacheSuppressWebhooks,
		FlushBatchSize:   statusCacheFlushBatchSize,
		CleanupInterval:  statusCacheCleanupInterval,
		OperationTimeout: statusCacheOperationTimeout,
	}

	cfg.Events = struct {
		BufferSize          int
		BatchSize           int
		PollInterval        time.Duration
		ProcessingTimeout   time.Duration
		HandlerTimeout      time.Duration
		ShutdownGracePeriod time.Duration

		MaxRetryAttempts int
		RetryDelays      []time.Duration

		CircuitBreakerEnabled bool
		CBMaxFailures         int
		CBTimeout             time.Duration
		CBCooldown            time.Duration

		DLQRetentionPeriod  time.Duration
		DLQReprocessEnabled bool

		MediaBufferSize      int
		MediaBatchSize       int
		MediaMaxRetries      int
		MediaPollInterval    time.Duration
		MediaDownloadTimeout time.Duration
		MediaUploadTimeout   time.Duration
		MediaMaxFileSize     int64
		MediaChunkSize       int64

		WebhookTimeout      time.Duration
		WebhookMaxRetries   int
		TransportBufferSize int

		DeliveredRetention time.Duration
		CleanupInterval    time.Duration

		DebugRawPayload bool
		DebugDumpDir    string
	}{
		BufferSize:            eventBufferSize,
		BatchSize:             eventBatchSize,
		PollInterval:          eventPollInterval,
		ProcessingTimeout:     eventProcessingTimeout,
		HandlerTimeout:        eventHandlerTimeout,
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
		CleanupInterval:       eventCleanupInterval,
		DebugRawPayload:       debugRawPayload,
		DebugDumpDir:          debugDumpDir,
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

func parseStringSlice(val string) []string {
	parts := strings.Split(val, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
