package webshare

import (
	"context"

	"golang.org/x/time/rate"
)

// RateLimiter manages Webshare API rate limits.
// Webshare has: 240 req/min general, 60 req/min proxy list.
type RateLimiter struct {
	general   *rate.Limiter
	proxyList *rate.Limiter
}

// NewRateLimiter creates a rate limiter for Webshare API.
// Webshare limits: 240 req/min general, 60 req/min proxy list.
// We use conservative rates (under limit) to avoid 429s during long syncs.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		general:   rate.NewLimiter(rate.Limit(3.5), 5), // 3.5 req/s, burst 5 (~210 req/min < 240)
		proxyList: rate.NewLimiter(rate.Limit(0.9), 1), // 0.9 req/s, no burst (~54 req/min < 60)
	}
}

// WaitGeneral waits for general rate limit allowance.
func (rl *RateLimiter) WaitGeneral(ctx context.Context) error {
	return rl.general.Wait(ctx)
}

// WaitProxyList waits for proxy list rate limit allowance.
func (rl *RateLimiter) WaitProxyList(ctx context.Context) error {
	if err := rl.general.Wait(ctx); err != nil {
		return err
	}
	return rl.proxyList.Wait(ctx)
}
