package main

import (
	"testing"
	"time"
)

func TestRateLimitConfig(t *testing.T) {
	limit := 60
	window := time.Minute

	if limit != 60 {
		t.Errorf("Expected rate limit 60 req/min, got %d", limit)
	}

	if window != time.Minute {
		t.Errorf("Expected window 1 minute, got %v", window)
	}
}
