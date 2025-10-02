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
	}{
		Endpoint:  getEnv("S3_ENDPOINT", "http://localhost:9000"),
		Region:    getEnv("S3_REGION", "us-east-1"),
		Bucket:    getEnv("S3_BUCKET", "funnelchat-media"),
		AccessKey: os.Getenv("S3_ACCESS_KEY"),
		SecretKey: os.Getenv("S3_SECRET_KEY"),
		UseSSL:    parseBool(getEnv("S3_USE_SSL", "false")),
	}
	expiry, err := parseDuration(getEnv("S3_URL_EXPIRATION", "30d"))
	if err != nil {
		return cfg, fmt.Errorf("invalid S3_URL_EXPIRATION: %w", err)
	}
	cfg.S3.URLExpiry = expiry

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
