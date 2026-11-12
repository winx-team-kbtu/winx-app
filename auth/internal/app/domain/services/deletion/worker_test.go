package deletion

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type countingRetrier struct {
	calls  atomic.Int32
	report Report
	err    error
}

func (c *countingRetrier) RetryPending(context.Context) (Report, error) {
	c.calls.Add(1)

	return c.report, c.err
}

func runWorker(t *testing.T, r Retrier, interval time.Duration, d time.Duration) (logs *atomic.Int32) {
	t.Helper()

	logs = &atomic.Int32{}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		RunWorker(ctx, r, interval, func(string, ...any) { logs.Add(1) })
		close(done)
	}()

	time.Sleep(d)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after the context was cancelled")
	}

	return logs
}

func TestWorker_RunsImmediatelyThenOnEveryTick(t *testing.T) {
	r := &countingRetrier{}

	runWorker(t, r, 20*time.Millisecond, 110*time.Millisecond)

	assert.GreaterOrEqual(t, int(r.calls.Load()), 4, "one pass at start plus several ticks")
}

func TestWorker_PicksUpLeftoversWithoutWaitingForTheFirstTick(t *testing.T) {
	r := &countingRetrier{}

	runWorker(t, r, time.Hour, 50*time.Millisecond)

	assert.EqualValues(t, 1, r.calls.Load(), "deletions left over from before a restart are retried at once")
}

func TestWorker_StopsOnCancel(t *testing.T) {
	r := &countingRetrier{}

	runWorker(t, r, 10*time.Millisecond, 30*time.Millisecond)
	after := r.calls.Load()

	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, after, r.calls.Load(), "no passes after shutdown")
}

func TestWorker_KeepsRunningAfterAFailedPassAndLogsIt(t *testing.T) {
	r := &countingRetrier{err: errors.New("db down")}

	logs := runWorker(t, r, 20*time.Millisecond, 90*time.Millisecond)

	require.GreaterOrEqual(t, int(r.calls.Load()), 3, "a failing pass must not kill the worker")
	assert.GreaterOrEqual(t, int(logs.Load()), 3, "every failure is logged")
}

func TestWorker_LogsOnlyWhenSomethingWasAttempted(t *testing.T) {
	quiet := &countingRetrier{}
	busy := &countingRetrier{report: Report{Attempted: 2, Completed: 1, Failed: 1}}

	assert.Zero(t, runWorker(t, quiet, 20*time.Millisecond, 60*time.Millisecond).Load(), "idle passes are silent")
	assert.Positive(t, runWorker(t, busy, 20*time.Millisecond, 60*time.Millisecond).Load())
}
