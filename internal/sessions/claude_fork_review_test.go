package sessions

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

// A reordered fork-context-ref takes the full JSON path, as do messages whose
// id comes before their model. Neither shape makes the copied head a new call
func TestClaudeForkHeadFullJSON(t *testing.T) {
	for _, marker := range []string{
		`{"agentId":"f1","type":"fork-context-ref","contextLength":1}`,
		` { "agentId": "f1", "type" : "fork-context-ref", "contextLength": 1 }`,
	} {
		for _, fast := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/fast=%t", marker, fast), func(t *testing.T) {
				s := &state{}
				at := time.Date(2026, 9, 21, 10, 0, 0, 0, time.UTC)
				claudeLine(s, []byte(marker), false)
				claudeLine(s, claudeUsageLine(at, "parent", "req-parent", "m", Tokens{Input: 100, Output: 10}, "", fast), false)
				claudeLine(s, []byte(`{"type":"user","message": {"role":"user","content":"port it"}}`), false)
				claudeLine(s, claudeUsageLine(at.Add(time.Second), "fork", "req-fork", "m", Tokens{Input: 50, Output: 5}, "", fast), false)
				if got := s.Models.of("m"); got != (Tokens{Input: 50, Output: 5}) {
					t.Fatalf("fork tokens = %+v, want only its own 50 input and 5 output", got)
				}
			})
		}
	}
}

// A scan can stop between a fork marker and its first prompt. The binary
// cache must carry that state to the next scan, including further head copies
func TestClaudeForkHeadCacheResume(t *testing.T) {
	for _, fast := range []bool{false, true} {
		t.Run(fmt.Sprint("fast=", fast), func(t *testing.T) {
			s := &state{Agent: "claude"}
			at := time.Date(2026, 9, 21, 10, 0, 0, 0, time.UTC)
			claudeLine(s, []byte(`{"type":"fork-context-ref","agentId":"f1","contextLength":2}`), false)
			claudeLine(s, claudeUsageLine(at, "parent-one", "req-one", "m", Tokens{Input: 100, Output: 10}, "", fast), false)
			var b bytes.Buffer
			if err := encodeCache(&b, cacheFile{Version: cacheVersion, Parsers: parseVersions, Files: map[string]*state{"fork": s}}); err != nil {
				t.Fatal(err)
			}
			files, ok := decodeCache(b.Bytes(), "")
			if !ok {
				t.Fatal("could not decode fork cache")
			}
			resumed := files["fork"]
			if resumed == nil {
				t.Fatal("fork parse missing from cache")
			}
			claudeLine(resumed, claudeUsageLine(at, "parent-two", "req-two", "m", Tokens{Input: 200, Output: 20}, "", fast), false)
			claudeLine(resumed, []byte(`{"message":{"role":"user","content":"port it"},"type":"user"}`), false)
			claudeLine(resumed, claudeUsageLine(at.Add(time.Second), "fork", "req-fork", "m", Tokens{Input: 50, Output: 5}, "", fast), false)
			if got := resumed.Models.of("m"); got != (Tokens{Input: 50, Output: 5}) {
				t.Fatalf("resumed fork tokens = %+v, want only its own 50 input and 5 output", got)
			}
		})
	}
}

// An ordinary subagent has no fork marker. Its assistant reply before its
// first user line, even one mentioning the marker as text, is still its own
func TestClaudeOrdinarySubagentKeepsHeadReply(t *testing.T) {
	for _, fast := range []bool{false, true} {
		t.Run(fmt.Sprint("fast=", fast), func(t *testing.T) {
			s := &state{}
			at := time.Date(2026, 9, 21, 10, 0, 0, 0, time.UTC)
			claudeLine(s, claudeUsageLine(at, "first", "req-first", "m", Tokens{Input: 100, Output: 10}, `[{"type":"text","text":"fork-context-ref"}]`, fast), false)
			claudeLine(s, []byte(`{"type":"user","message":{"role":"user","content":"port it"}}`), false)
			claudeLine(s, claudeUsageLine(at.Add(time.Second), "next", "req-next", "m", Tokens{Input: 50, Output: 5}, "", fast), false)
			if got := s.Models.of("m"); got != (Tokens{Input: 150, Output: 15}) {
				t.Fatalf("ordinary subagent tokens = %+v, want both replies", got)
			}
		})
	}
}
