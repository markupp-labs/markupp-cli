package session

import (
	"context"
	"testing"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/config"
)

func TestLockIsHeldForLessThanItTakesToGoStale(t *testing.T) {
	if margin := config.StaleAfter - config.MaxHold; margin < 30*time.Second {
		t.Errorf("StaleAfter %s leaves %s over MaxHold %s, want at least 30s", config.StaleAfter, margin, config.MaxHold)
	}
}

func TestWithLockBoundsTheCriticalSection(t *testing.T) {
	store := newStore(t)
	start := time.Now()
	var deadline time.Time
	err := withLock(context.Background(), store, func(ctx context.Context, _ *config.Settings) error {
		deadline, _ = ctx.Deadline()
		return nil
	})
	if err != nil {
		t.Fatalf("withLock: %v", err)
	}
	if deadline.IsZero() || deadline.After(start.Add(config.MaxHold).Add(time.Second)) {
		t.Errorf("critical section deadline = %v, want within %s of %v", deadline, config.MaxHold, start)
	}
}
