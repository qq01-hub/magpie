package gateway

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

// Output marks keep their IDs and a digest, not the potentially huge tool
// arguments. Those retained values must still fit the store's byte cap.
func TestRemoteSentCacheCountsRetainedOutput(t *testing.T) {
	for _, c := range []struct {
		name   string
		output []itemMark
	}{
		{"long id", []itemMark{{typ: "function_call", id: strings.Repeat("x", remoteSentCap/2), callID: "call"}}},
		{"many digests", make([]itemMark, remoteSentCap/2/(64+sha256.Size)+1)},
	} {
		t.Run(c.name, func(t *testing.T) {
			s := newRemoteSentStore()
			s.put(&remoteSent{key: "route\x00scope", output: c.output}, nil)
			if s.size != 0 || len(s.m) != 0 || s.order.Len() != 0 {
				t.Fatalf("oversized output retained: %d bytes, %d keys, %d turns", s.size, len(s.m), s.order.Len())
			}
		})
	}
}

func TestRemoteOutputNormalizationKeepsContentBoundaries(t *testing.T) {
	for _, c := range []struct {
		name, old, next string
		want            bool
	}{
		{"default final omitted", `{"type":"message","role":"assistant","phase":"final_answer","content":[{"type":"output_text","text":"answer"}]}`, userSaw("answer"), true},
		{"default final restored", userSaw("answer"), `{"type":"message","role":"assistant","phase":"final_answer","content":[{"type":"output_text","text":"answer"}]}`, true},
		{"nested metadata omitted", `{"type":"custom","value":{"text":"answer","extra":null,"unused":[],"annotations":[{"note":"metadata"}],"logprobs":[1]}}`, `{"value":{"text":"answer"},"type":"custom"}`, true},
		{"nested content edited", `{"type":"custom","value":{"text":"answer"}}`, `{"type":"custom","value":{"text":"edited"}}`, false},
		{"argument spelling kept", `{"type":"function_call","call_id":"c","name":"tool","arguments":"{\"n\":1}"}`, `{"type":"function_call","call_id":"c","name":"tool","arguments":"{ \"n\": 1 }"}`, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := marksMatch([]itemMark{markOf([]byte(c.old))}, []json.RawMessage{json.RawMessage(c.next)}); got != c.want {
				t.Fatalf("cached output reused = %v, want %v", got, c.want)
			}
		})
	}
}

func TestRemoteSentCacheKeepsScopeAndExpiry(t *testing.T) {
	now := time.Unix(1700000000, 0)
	s := newRemoteSentStore()
	s.now = func() time.Time { return now }
	first, output, tail := json.RawMessage(userAsks("first")), json.RawMessage(userSaw("answer")), json.RawMessage(userAsks("next"))
	turn := &remoteSent{key: "route\x00scope", input: []itemSum{sha256.Sum256(first)}, output: []itemMark{markOf(output)}, id: "resp"}
	s.put(turn, nil)
	items := []json.RawMessage{first, output, tail}
	sums := []itemSum{sha256.Sum256(first), sha256.Sum256(output), sha256.Sum256(tail)}
	if got, n := s.match(turn.key, sums, items); got != turn || n != 2 {
		t.Fatalf("same scope did not continue kept turn: got %p, covered %d", got, n)
	}
	if got, _ := s.match("route\x00other-scope", sums, items); got != nil {
		t.Fatal("another scope reused the turn")
	}
	now = now.Add(prevTTL + time.Second)
	if got, _ := s.match(turn.key, sums, items); got != nil {
		t.Fatal("expired turn reused")
	}
	// Each route remains bounded even when the new content marks are small.
	for i := 0; i <= remotePerRoute; i++ {
		s.put(&remoteSent{key: "other-route\x00scope", id: fmt.Sprint(i)}, nil)
		now = now.Add(time.Second)
	}
	if got := len(s.m["other-route\x00scope"]); got != remotePerRoute {
		t.Fatalf("kept %d turns per route, want %d", got, remotePerRoute)
	}
}
