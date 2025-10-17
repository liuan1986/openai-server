package ratelimit

import (
	"sync"
	"time"
)

type bucket struct {
	tokens     int
	lastRefill time.Time
}

type Limiter struct {
	capacity int
	refill   time.Duration

	mu      sync.Mutex
	buckets map[string]*bucket
}

func New(capacity int, refillIntervalSeconds int) *Limiter {
	return &Limiter{
		capacity: capacity,
		refill:   time.Duration(refillIntervalSeconds) * time.Second,
		buckets:  make(map[string]*bucket),
	}
}

func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.capacity, lastRefill: time.Now()}
		l.buckets[key] = b
	}

	now := time.Now()
	if now.Sub(b.lastRefill) >= l.refill {
		b.tokens = l.capacity
		b.lastRefill = now
	}

	if b.tokens <= 0 {
		return false
	}

	b.tokens--
	return true
}
