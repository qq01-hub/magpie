package provider

import (
	"context"
	"time"
)

// KeepNewListed fetches missing account lists at gateway startup and each
// minute after, even after every account present at startup has its list
// FetchNew's retry delay and existing-list checks still apply (#7945)
// Based on kkkzbh's KeepNewListed patch for #1805
func KeepNewListed(ctx context.Context) {
	keepNewListed(ctx, time.Minute)
}

func keepNewListed(ctx context.Context, every time.Duration) {
	t := time.NewTimer(0)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
		if ctx.Err() != nil {
			return
		}
		// A page or the startup fetch may already own the work; never wait
		// for its mutex across the serving gateway's cancellation
		if newFetches.TryLock() {
			newRunning.Add(1)
			fetchNewHeld(ctx, 20*time.Second)
			newRunning.Add(-1)
			newFetches.Unlock()
		}
		t.Reset(every)
	}
}
