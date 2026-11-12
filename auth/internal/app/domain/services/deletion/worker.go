package deletion

import (
	"context"
	"time"
)

// Retrier is what the worker needs from the deletion service.
type Retrier interface {
	RetryPending(ctx context.Context) (Report, error)
}

// RunWorker retries pending deletions every interval until ctx is cancelled. It runs one pass
// immediately so deletions left over from before a restart are picked up without waiting.
func RunWorker(ctx context.Context, retrier Retrier, interval time.Duration, logf func(format string, args ...any)) {
	pass := func() {
		report, err := retrier.RetryPending(ctx)
		if err != nil {
			logf("account deletion retry failed: %v", err)

			return
		}

		if report.Attempted > 0 {
			logf("account deletion retry: attempted=%d completed=%d failed=%d", report.Attempted, report.Completed, report.Failed)
		}
	}

	pass()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pass()
		}
	}
}
