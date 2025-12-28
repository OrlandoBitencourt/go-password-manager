package telegram

import (
	"sync"
	"time"
)

// RateLimiter implements a simple rate limiter for commands
type RateLimiter struct {
	requests map[int64][]time.Time
	mu       sync.Mutex
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
// limit: maximum number of requests allowed
// window: time window for the limit
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[int64][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a user is allowed to make a request
func (rl *RateLimiter) Allow(userID int64) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Get user's request history
	userRequests, exists := rl.requests[userID]
	if !exists {
		userRequests = []time.Time{}
	}

	// Remove requests outside the window
	validRequests := make([]time.Time, 0)
	for _, reqTime := range userRequests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if limit is exceeded
	if len(validRequests) >= rl.limit {
		rl.requests[userID] = validRequests
		return false
	}

	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[userID] = validRequests

	return true
}

// Reset clears the rate limit for a user
func (rl *RateLimiter) Reset(userID int64) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.requests, userID)
}

// CleanupOldEntries removes old entries to prevent memory leak
func (rl *RateLimiter) CleanupOldEntries() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window * 2)

	for userID, requests := range rl.requests {
		validRequests := make([]time.Time, 0)
		for _, reqTime := range requests {
			if reqTime.After(cutoff) {
				validRequests = append(validRequests, reqTime)
			}
		}

		if len(validRequests) == 0 {
			delete(rl.requests, userID)
		} else {
			rl.requests[userID] = validRequests
		}
	}
}
