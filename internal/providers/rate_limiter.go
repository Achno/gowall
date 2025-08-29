package providers

import (
	"context"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiter *rate.Limiter
	enabled bool
}

func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	if config.RPS <= 0 {
		return &RateLimiter{enabled: false}
	}

	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(config.RPS), config.Burst),
		enabled: true,
	}
}

func (r *RateLimiter) Wait(ctx context.Context) error {
	if !r.enabled {
		return nil
	}
	return r.limiter.Wait(ctx)
}

func (r *RateLimiter) IsEnabled() bool {
	return r.enabled
}
