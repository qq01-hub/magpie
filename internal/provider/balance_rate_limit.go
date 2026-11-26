package provider

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	balanceRetryBase = 5 * time.Second
	balanceRetryMax  = 15 * time.Minute
	// Preserve concurrent key reads but spread their requests to one origin.
	balanceReadStagger = 50 * time.Millisecond
)

type balanceRateLimitError struct {
	until time.Time
	now   func() time.Time
}

func (e *balanceRateLimitError) Error() string {
	now := time.Now()
	if e.now != nil {
		now = e.now()
	}
	wait := e.until.Sub(now).Round(time.Second)
	if wait < time.Second {
		wait = time.Second
	}
	return fmt.Sprintf("429 Too Many Requests; balance lookup paused, retry after %s", wait)
}

type balanceOriginState struct {
	lastStart time.Time
	changed   chan struct{}
}

type balanceRetryState struct {
	until    time.Time
	failures int
}

// balanceReadLimits spaces reads to one balance origin and pauses an endpoint after
// a rate limit. Balance polling can fan out across a provider's keys; the
// upstream balance endpoint is shared by those requests.
type balanceReadLimitsT struct {
	mu      sync.Mutex
	origins map[string]*balanceOriginState
	retries map[string]balanceRetryState
	now     func() time.Time
	jitter  func(time.Duration) time.Duration
}

var balanceReadLimits = newBalanceReadLimits()

func newBalanceReadLimits() *balanceReadLimitsT {
	return &balanceReadLimitsT{
		origins: make(map[string]*balanceOriginState),
		retries: make(map[string]balanceRetryState),
		now:     time.Now,
		jitter:  randomBalanceJitter,
	}
}

func balanceReadOrigin(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return strings.ToLower(u.Scheme) + "://" + strings.ToLower(u.Host)
}

// balanceReadID scopes a cooldown to an endpoint path while keeping its URL
// query out of memory: custom Balance URLs may put the key in that query.
func balanceReadID(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return keyTag("balance-rate", raw)
	}
	return keyTag("balance-rate", strings.ToLower(u.Scheme)+"://"+strings.ToLower(u.Host)+u.EscapedPath())
}

// take staggers balance reads to one origin. A cooldown is checked both
// before and after its wait, so a request queued behind a 429 never reaches
// the same upstream limit window.
func (l *balanceReadLimitsT) take(ctx context.Context, raw string) error {
	origin := balanceReadOrigin(raw)
	if origin == "" {
		return nil
	}
	for {
		if err := l.cooldownError(raw); err != nil {
			return err
		}
		now := l.now()
		l.mu.Lock()
		s := l.origins[origin]
		if s == nil {
			s = &balanceOriginState{changed: make(chan struct{})}
			l.origins[origin] = s
		}
		wait := s.lastStart.Add(balanceReadStagger).Sub(now)
		if wait <= 0 {
			s.lastStart = now
			l.mu.Unlock()
			return l.cooldownError(raw)
		}
		changed := s.changed
		l.mu.Unlock()
		t := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			if !t.Stop() {
				<-t.C
			}
			return ctx.Err()
		case <-changed:
			if !t.Stop() {
				<-t.C
			}
		case <-t.C:
		}
	}
}

func (l *balanceReadLimitsT) cooldownError(raw string) error {
	id := balanceReadID(raw)
	now := l.now()
	l.mu.Lock()
	r := l.retries[id]
	l.mu.Unlock()
	if !r.until.After(now) {
		return nil
	}
	return &balanceRateLimitError{until: r.until, now: l.now}
}

// rateLimited records only a rate-limit refusal. Credit exhaustion is a
// different failure and must not be retried as a transient 429.
func (l *balanceReadLimitsT) rateLimited(raw, retryAfter string, body []byte) error {
	if balanceCreditsExhausted(body) {
		l.clear(raw)
		return nil
	}
	id := balanceReadID(raw)
	now := l.now()
	l.mu.Lock()
	r := l.retries[id]
	r.failures = min(r.failures+1, 10)
	until := balanceRetryAfter(retryAfter, now)
	if !until.After(now) {
		until = now.Add(l.backoff(r.failures))
	}
	if r.until.After(until) {
		until = r.until
	}
	r.until = until
	l.retries[id] = r
	if s := l.origins[balanceReadOrigin(raw)]; s != nil {
		close(s.changed)
		s.changed = make(chan struct{})
	}
	l.mu.Unlock()
	return &balanceRateLimitError{until: until, now: l.now}
}

func (l *balanceReadLimitsT) clear(raw string) {
	id := balanceReadID(raw)
	now := l.now()
	l.mu.Lock()
	if r := l.retries[id]; r.until.After(now) {
		l.mu.Unlock()
		return
	}
	delete(l.retries, id)
	l.mu.Unlock()
}

func (l *balanceReadLimitsT) backoff(failures int) time.Duration {
	d := balanceRetryBase
	for range min(max(0, failures-1), 8) {
		d = min(d*2, balanceRetryMax)
	}
	spread := d / 4
	if spread <= 0 {
		return d
	}
	d += l.jitter(spread*2) - spread
	return max(time.Second, min(d, balanceRetryMax))
}

func balanceRetryAfter(value string, now time.Time) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}
	if secs, err := strconv.ParseInt(value, 10, 64); err == nil {
		if secs <= 0 {
			return time.Time{}
		}
		maxSeconds := int64((1<<63 - 1) / int64(time.Second))
		if secs > maxSeconds {
			return now.Add(time.Duration(1<<63 - 1))
		}
		return now.Add(time.Duration(secs) * time.Second)
	}
	until, err := http.ParseTime(value)
	if err != nil || !until.After(now) {
		return time.Time{}
	}
	return until
}

func balanceCreditsExhausted(body []byte) bool {
	s := strings.ToLower(string(body))
	for _, phrase := range []string{
		"insufficient balance", "balance is insufficient", "balance exhausted",
		"insufficient credit", "credit exhausted", "insufficient_quota",
		"余额不足", "余额已耗尽", "额度已用尽", "积分不足", "欠费", "请充值",
	} {
		if strings.Contains(s, phrase) {
			return true
		}
	}
	return false
}

func randomBalanceJitter(span time.Duration) time.Duration {
	if span <= 0 {
		return 0
	}
	return time.Duration(rand.Int64N(int64(span)))
}
