package redis

import (
	"crypto/tls"

	redis "github.com/redis/go-redis/v9"
)

type Config struct {
	Addr       string
	Username   string
	Password   string
	DB         int
	TLSEnabled bool
}

// NewClient returns a configured Redis client.
func NewClient(cfg Config) *redis.Client {
	options := &redis.Options{
		Addr:     cfg.Addr,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	}
	if cfg.TLSEnabled {
		options.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	return redis.NewClient(options)
}
