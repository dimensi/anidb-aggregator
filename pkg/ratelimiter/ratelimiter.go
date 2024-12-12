package ratelimiter

import (
	"time"
)

type RateLimiter struct {
	requests chan time.Time
	ticker   *time.Ticker
}

func New(requestsPerSecond, requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		requests: make(chan time.Time, requestsPerMinute),
		ticker:   time.NewTicker(time.Second / time.Duration(requestsPerSecond)),
	}

	go func() {
		minuteTicker := time.NewTicker(time.Minute)
		for range minuteTicker.C {
			for len(rl.requests) > 0 {
				<-rl.requests
			}
		}
	}()

	return rl
}

func (rl *RateLimiter) Wait() {
	<-rl.ticker.C
	rl.requests <- time.Now()

	if len(rl.requests) >= cap(rl.requests) {
		<-rl.requests
	}
}
