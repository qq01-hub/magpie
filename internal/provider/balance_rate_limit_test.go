package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func balanceRateLimitSandbox(t *testing.T) {
	t.Helper()
	probeSandbox(t)
	old := balanceReadLimits
	balanceReadLimits = newBalanceReadLimits()
	t.Cleanup(func() { balanceReadLimits = old })
}

func saveBalanceRateTestProvider(t *testing.T, p Provider) {
	t.Helper()
	if err := Save(p); err != nil {
		t.Fatal(err)
	}
}

func TestBalanceReadLimitsTakeCanceledContextDoesNotReserveOrigin(t *testing.T) {
	limits := newBalanceReadLimits()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	const endpoint = "https://relay.example/api/usage/token"

	if err := limits.take(ctx, endpoint); err != context.Canceled {
		t.Fatalf("take with canceled context = %v, want context.Canceled", err)
	}
	if state := limits.origins[balanceReadOrigin(endpoint)]; state != nil && !state.lastStart.IsZero() {
		t.Fatalf("canceled read reserved the origin until %s", state.lastStart)
	}
}

func TestBalanceReadLimitsTakeCanceledWhileWaitingReturns(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		limits := newBalanceReadLimits()
		const endpoint = "https://relay.example/api/usage/token"
		if err := limits.take(context.Background(), endpoint); err != nil {
			t.Fatal(err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		done := make(chan error, 1)
		go func() { done <- limits.take(ctx, endpoint) }()
		synctest.Wait()
		cancel()
		synctest.Wait()
		if err := <-done; err != context.Canceled {
			t.Fatalf("take canceled while waiting = %v, want context.Canceled", err)
		}
	})
}

func TestKeyBalancesBackOffAfter429(t *testing.T) {
	for _, tc := range []struct {
		name      string
		header    func(time.Time) string
		waitAfter time.Duration
	}{
		{"seconds", func(time.Time) string { return "30" }, 31 * time.Second},
		{"http date", func(now time.Time) string { return now.Add(45 * time.Second).UTC().Format(http.TimeFormat) }, 46 * time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			balanceRateLimitSandbox(t)
			var clockOffset atomic.Int64
			balanceReadLimits.now = func() time.Time {
				return time.Now().Add(time.Duration(clockOffset.Load()))
			}

			retryAfter := tc.header(balanceReadLimits.now())
			var requests atomic.Int32
			var limited atomic.Bool
			var startsMu sync.Mutex
			var starts []time.Time
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				requests.Add(1)
				startsMu.Lock()
				starts = append(starts, time.Now())
				startsMu.Unlock()
				if limited.Load() {
					w.Header().Set("Retry-After", retryAfter)
					w.WriteHeader(http.StatusTooManyRequests)
					_, _ = w.Write([]byte("TooMany Requests"))
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprint(w, `{"data":{"total_available":1000000}}`)
			}))
			defer srv.Close()

			p := Provider{
				ID:          "balance-rate-test",
				Name:        "Balance rate test",
				Chat:        srv.URL + "/v1",
				Key:         "sk-main",
				KeyName:     "main",
				BalanceURL:  srv.URL + "/api/usage/token",
				BalancePath: "$data.total_available / 500000",
				Keys: []KeyAccount{
					{Name: "second", Key: "sk-second"},
					{Name: "third", Key: "sk-third"},
					{Name: "fourth", Key: "sk-fourth"},
				},
			}
			saveBalanceRateTestProvider(t, p)

			initial := KeyBalances(context.Background())
			if got, want := requests.Load(), int32(4); got != want {
				t.Fatalf("initial refresh sent %d requests, want %d", got, want)
			}
			startsMu.Lock()
			initialStarts := append([]time.Time(nil), starts[:4]...)
			startsMu.Unlock()
			sort.Slice(initialStarts, func(i, j int) bool { return initialStarts[i].Before(initialStarts[j]) })
			for i := 1; i < len(initialStarts); i++ {
				if gap := initialStarts[i].Sub(initialStarts[i-1]); gap < 40*time.Millisecond {
					t.Fatalf("same-origin balance reads started only %s apart, want staggered requests", gap)
				}
			}
			for _, q := range initial {
				if q.Balance != "$2.00" {
					t.Fatalf("initial balance for %q = %q, want $2.00", q.User, q.Balance)
				}
			}

			limited.Store(true)
			RefreshUsage(context.Background(), p.ID, "")
			if got := requests.Load(); got != 5 {
				t.Fatalf("refresh after a 429 sent %d requests, want only 1 new request", got)
			}
			for _, q := range KeyBalances(context.Background()) {
				if q.Balance != "$2.00" || q.ReadError == "" || !strings.Contains(q.ReadError, "429") {
					t.Fatalf("429 discarded last good balance for %q: %+v", q.User, q)
				}
			}

			RefreshUsage(context.Background(), p.ID, "")
			if got := requests.Load(); got != 5 {
				t.Fatalf("refresh during Retry-After sent %d requests, want 5 total", got)
			}

			clockOffset.Store(int64(tc.waitAfter))
			limited.Store(false)
			RefreshUsage(context.Background(), p.ID, "")
			if got, want := requests.Load(), int32(9); got != want {
				t.Fatalf("refresh after Retry-After sent %d requests, want %d", got, want)
			}
		})
	}
}

func TestBalance429WithoutRetryAfterBacksOff(t *testing.T) {
	balanceRateLimitSandbox(t)

	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte("TooMany Requests"))
	}))
	defer srv.Close()

	p := Provider{ID: "balance-backoff-test", Name: "Relay", Key: "sk", BalanceURL: srv.URL + "/api/usage/token", BalancePath: "data.total_available"}
	_, _, err := Balance(context.Background(), p)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "429") {
		t.Fatalf("first read error = %v, want 429", err)
	}
	_, _, err = Balance(context.Background(), p)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "retry") {
		t.Fatalf("read during fallback backoff error = %v, want a retry wait", err)
	}
	if got := requests.Load(); got != 1 {
		t.Fatalf("rate-limited endpoint was called %d times before backoff elapsed, want 1", got)
	}
}

func TestBalanceBackoffGrowsAfterRepeated429(t *testing.T) {
	now := time.Date(2026, 11, 26, 8, 0, 0, 0, time.UTC)
	limits := newBalanceReadLimits()
	limits.now = func() time.Time { return now }
	limits.jitter = func(span time.Duration) time.Duration { return span / 2 }
	const endpoint = "https://relay.example/api/usage/token"
	body := []byte("TooMany Requests")

	firstErr := limits.rateLimited(endpoint, "", body)
	first, ok := firstErr.(*balanceRateLimitError)
	if !ok {
		t.Fatalf("first 429 error = %T, want *balanceRateLimitError", firstErr)
	}
	if got, want := first.until.Sub(now), balanceRetryBase; got != want {
		t.Fatalf("first retry delay = %s, want %s", got, want)
	}
	if err := limits.cooldownError(endpoint); err == nil {
		t.Fatal("a request was allowed before the first backoff expired")
	}

	now = first.until
	secondErr := limits.rateLimited(endpoint, "", body)
	second, ok := secondErr.(*balanceRateLimitError)
	if !ok {
		t.Fatalf("second 429 error = %T, want *balanceRateLimitError", secondErr)
	}
	if got, want := second.until.Sub(now), 2*balanceRetryBase; got != want {
		t.Fatalf("second retry delay = %s, want exponential delay %s", got, want)
	}
}

func TestBalanceLateSuccessDoesNotClearActive429Cooldown(t *testing.T) {
	now := time.Date(2026, 11, 26, 8, 0, 0, 0, time.UTC)
	limits := newBalanceReadLimits()
	limits.now = func() time.Time { return now }
	const endpoint = "https://relay.example/api/usage/token"
	if err := limits.rateLimited(endpoint, "30", []byte("TooMany Requests")); err == nil {
		t.Fatal("429 did not establish a cooldown")
	}

	// A request that was already in flight can succeed after another read
	// received the 429; that late success must not release queued reads early.
	limits.clear(endpoint)
	if err := limits.cooldownError(endpoint); err == nil {
		t.Fatal("late success cleared the active Retry-After cooldown")
	}
}

func TestBalanceCredit429DoesNotBecomeRateLimitBackoff(t *testing.T) {
	balanceRateLimitSandbox(t)

	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if requests.Add(1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"code":1113,"message":"余额不足"}`))
			return
		}
		_, _ = fmt.Fprint(w, `{"data":{"total_available":500000}}`)
	}))
	defer srv.Close()

	p := Provider{ID: "balance-credit-test", Name: "Relay", Key: "sk", BalanceURL: srv.URL + "/api/usage/token", BalancePath: "data.total_available"}
	if _, _, err := Balance(context.Background(), p); err == nil {
		t.Fatal("first read succeeded despite the balance refusal")
	}
	if got, ok, err := Balance(context.Background(), p); err != nil || !ok || got != "500000.00" {
		t.Fatalf("second read = %q %v %v, want successful balance read", got, ok, err)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("balance refusal blocked the second request: got %d requests, want 2", got)
	}
}

func TestBalanceProbe429IsNotRememberedAsNoBalance(t *testing.T) {
	balanceRateLimitSandbox(t)

	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path == "/v1/usage" {
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte("TooMany Requests"))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	p := Provider{ID: "balance-probe-rate-test", Name: "Relay", Chat: srv.URL + "/v1", Key: "sk"}
	if _, ok, err := Balance(context.Background(), p); err == nil || !ok || !strings.Contains(err.Error(), "429") {
		t.Fatalf("probe result = %v, %v; want a reported 429", ok, err)
	}
	if got := requests.Load(); got != 1 {
		t.Fatalf("probe sent %d requests after a 429, want it to stop at 1", got)
	}
}
