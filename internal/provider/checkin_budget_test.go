package provider

import (
	"context"
	"errors"
	"os"
	"testing"
	"testing/synctest"
	"time"

	"github.com/yetone/magpie/internal/plugin"
	"github.com/yetone/magpie/internal/settings"
)

func TestCheckinRoundKeepsBuiltinTwoMinuteDeadline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		now := time.Date(2026, 11, 28, 12, 0, 0, 0, time.UTC)
		runner := checkinRunnerFunc{
			clockFunc: func() time.Time { return now },
			checkinFunc: func(ctx context.Context, _ bool) []WorkBuddyCheckin {
				time.Sleep(3 * time.Minute)
				if err := ctx.Err(); err != nil {
					return []WorkBuddyCheckin{{Outcome: CheckinFailed, Msg: err.Error()}}
				}
				return []WorkBuddyCheckin{{Outcome: CheckinDone}}
			},
		}
		var loop wbCheckinLoop
		got := loop.tick(context.Background(), runner)
		if len(got) != 1 || got[0].Outcome != CheckinFailed || got[0].Msg != context.DeadlineExceeded.Error() {
			t.Fatalf("built-in check-in after a 3m operation: %+v", got)
		}
	})
}

func TestPluginCheckinOwnBudgetLetsBothAccountsFinishPastRound(t *testing.T) {
	pluginCheckinBudgetFixture(t, "budget-a", "budget-b")
	old := pluginAccountCheckin
	var calls []string
	pluginAccountCheckin = func(ctx context.Context, id, account string) (plugin.Checkin, error) {
		calls = append(calls, account)
		if id != "fakeco" || ctx.Value(checkinBudgetTestKey{}) != "round-value" {
			t.Fatalf("plugin call context/provider: id=%q value=%v", id, ctx.Value(checkinBudgetTestKey{}))
		}
		time.Sleep(3 * time.Minute)
		if err := ctx.Err(); err != nil {
			return plugin.Checkin{}, err
		}
		return plugin.Checkin{Outcome: CheckinClaimed}, nil
	}
	t.Cleanup(func() { pluginAccountCheckin = old })
	synctest.Test(t, func(t *testing.T) {
		parent := context.WithValue(context.Background(), checkinBudgetTestKey{}, "round-value")
		var loop wbCheckinLoop
		c := pluginCheckiner{now: func() time.Time { return time.Date(2026, 11, 28, 12, 0, 0, 0, time.UTC) }}
		rs := loop.tick(parent, c)
		if len(rs) != 2 || rs[0].Outcome != CheckinClaimed || rs[1].Outcome != CheckinClaimed {
			t.Fatalf("plugin accounts after round budget: %+v", rs)
		}
	})
	if len(calls) != 2 {
		t.Fatalf("plugin account calls = %d, want 2", len(calls))
	}
}

func TestPluginCheckinParentCancellationAfterRoundStopsNextAccount(t *testing.T) {
	pluginCheckinBudgetFixture(t, "cancel-a", "cancel-b")
	old := pluginAccountCheckin
	var calls int
	pluginAccountCheckin = func(ctx context.Context, _, _ string) (plugin.Checkin, error) {
		calls++
		if calls == 1 {
			time.Sleep(3 * time.Minute)
			return plugin.Checkin{Outcome: CheckinClaimed}, nil
		}
		select {
		case <-ctx.Done():
			return plugin.Checkin{}, ctx.Err()
		case <-time.After(3 * time.Minute):
			return plugin.Checkin{Outcome: CheckinClaimed}, nil
		}
	}
	t.Cleanup(func() { pluginAccountCheckin = old })
	synctest.Test(t, func(t *testing.T) {
		parent, cancel := context.WithCancel(context.Background())
		defer cancel()
		time.AfterFunc(4*time.Minute, cancel)
		var loop wbCheckinLoop
		c := pluginCheckiner{now: func() time.Time { return time.Date(2026, 11, 28, 12, 0, 0, 0, time.UTC) }}
		rs := loop.tick(parent, c)
		if len(rs) != 2 || rs[0].Outcome != CheckinClaimed || rs[1].Outcome != CheckinFailed {
			t.Fatalf("parent cancellation after round: %+v", rs)
		}
	})
}

func pluginCheckinBudgetFixture(t *testing.T, keys ...string) []pluginCheckinProvider {
	t.Helper()
	pluginCheckinSandbox(t, keys...)
	s := settings.Load()
	s.PluginCheckins = map[string]bool{"fakeco": true}
	if err := settings.Save(s); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(pluginCheckinPath()); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	ps := pluginCheckinProviders()
	if len(ps) != 1 || len(ps[0].accounts) != len(keys) {
		t.Fatalf("plugin check-in fixture: %+v", ps)
	}
	return ps
}

func TestCheckinBudgetKeepsParentValuesAndOwnDeadline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		parent := context.WithValue(context.Background(), checkinBudgetTestKey{}, "round-value")
		ctx, cancel := withCheckinBudget(parent)
		defer cancel()
		if got := ctx.Value(checkinBudgetTestKey{}); got != "round-value" {
			t.Fatalf("budget context value = %v, want round-value", got)
		}
		if deadline, ok := ctx.Deadline(); !ok || !deadline.Equal(time.Now().Add(checkinBudget)) {
			t.Fatalf("budget deadline = %v, %v; want ten minutes", deadline, ok)
		}
		synctest.Wait()
		time.Sleep(checkinBudget)
		synctest.Wait()
		if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
			t.Fatalf("budget error = %v, want deadline exceeded", ctx.Err())
		}
	})
}

func TestCheckinBudgetHonorsParentCancellationAndDeadline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		parent, cancel := context.WithCancel(context.Background())
		ctx, stop := withCheckinBudget(parent)
		cancel()
		synctest.Wait()
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Fatalf("canceled parent error = %v, want canceled", ctx.Err())
		}
		stop()

		parent, cancel = context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		ctx, stop = withCheckinBudget(parent)
		defer stop()
		time.Sleep(time.Minute)
		synctest.Wait()
		if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
			t.Fatalf("parent deadline error = %v, want deadline exceeded", ctx.Err())
		}
	})
}

type checkinBudgetTestKey struct{}

type checkinRunnerFunc struct {
	clockFunc   func() time.Time
	checkinFunc func(context.Context, bool) []WorkBuddyCheckin
}

func (f checkinRunnerFunc) clock() time.Time {
	return f.clockFunc()
}

func (f checkinRunnerFunc) checkinNow(ctx context.Context, soon bool) []WorkBuddyCheckin {
	return f.checkinFunc(ctx, soon)
}
