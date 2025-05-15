package supermarket

import (
	"math/rand/v2"
	"time"
)

type RateLimiter struct {
	base   time.Duration
	jitter float64
}

func NewRateLimiter(base time.Duration, jitter float64) *RateLimiter {
	return &RateLimiter{
		base:   base,
		jitter: jitter,
	}
}

func (r *RateLimiter) Wait() {
	if r.base <= 0 {
		return
	}
	jitterRange := float64(r.base) * r.jitter
	jitterDuration := time.Duration(rand.Float64()*jitterRange*2 - jitterRange)
	time.Sleep(r.base + jitterDuration)
}
