package scheduler

import (
	"context"
	"time"
)

// Schedule runs job immediately and then on every interval tick until ctx is cancelled.
// Each invocation runs in its own goroutine so the caller is not blocked.
func Schedule(ctx context.Context, interval time.Duration, job func(ctx context.Context)) {
	go func() {
		job(ctx)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				job(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}
