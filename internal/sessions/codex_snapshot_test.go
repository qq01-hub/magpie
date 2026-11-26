package sessions

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

func snapshotMeta() string {
	return fmt.Sprintf(`{"timestamp":"%s","type":"session_meta","payload":{"id":"child","session_id":"parent","history_mode":"paginated","subagent_history_start_ordinal":219}}`, stamp(0))
}

func snapshotCompaction() string {
	return strings.Replace(cxCompactionLine(0, "snapshot", cxUse(1000, 500, 0, 100, 0), cxUse(30, 10, 0, 2, 0)), `"response_id":"snapshot"`, `"thread_id":"child","response_id":"snapshot"`, 1)
}

func TestCodexUsagePaginatedSnapshotBoundary(t *testing.T) {
	u := cxUse(100, 50, 0, 10, 0)
	for _, before := range []bool{false, true} {
		for split := 1; split <= 8; split++ {
			t.Run(fmt.Sprintf("recordBefore=%v/split=%d", before, split), func(t *testing.T) {
				lines := []string{snapshotMeta()}
				if before {
					lines = append(lines, cxRecordLine(0, "snapshot", cxUse(1000, 500, 0, 100, 0), cxUse(30, 10, 0, 2, 0)))
				}
				lines = append(lines, snapshotCompaction(), tokenCountLine(0, cxUse(900, 450, 0, 90, 0), cxUse(0, 0, 0, 0, 0)))
				lines = append(lines, cxTurnLines(1, "t1", "gpt-6-astra", "high")...)
				lines = append(lines,
					cxRecordLine(2, "a", cxUse(1100, 550, 0, 110, 0), u), tokenCountLine(2, cxUse(1000, 500, 0, 100, 0), u),
					cxCompactionLine(3, "new-compaction", cxUse(1300, 580, 0, 130, 0), cxUse(200, 30, 0, 20, 0)),
					cxRecordLine(4, "b", cxUse(1400, 630, 0, 140, 0), u), tokenCountLine(4, cxUse(1100, 550, 0, 110, 0), u),
					// A real legacy-only request with equal usage remains a new call.
					tokenCountLine(5, cxUse(1200, 600, 0, 120, 0), u))
				path := codexUsageLog(t, lines[:split])
				Calls(time.Time{})
				List(0)
				Reset()
				appendText(t, path, strings.Join(lines[split:], "\n")+"\n")
				cs := Calls(time.Time{})
				if len(cs) != 5 {
					t.Fatalf("calls=%+v", cs)
				}
				for i, c := range cs {
					if c.Model != "gpt-6-astra" {
						t.Fatalf("call[%d].Model = %q, want gpt-6-astra", i, c.Model)
					}
				}
				var total Tokens
				for _, c := range cs {
					total.add(c.Tokens)
				}
				if total != (Tokens{340, 52, 190, 0, 0}) {
					t.Fatalf("tokens=%+v", total)
				}
				ss := List(0)
				if len(ss) != 1 || ss[0].Tokens != total {
					t.Fatalf("summary=%+v", ss)
				}
				if len(ss[0].Models) != 1 || ss[0].Models[0].Model != "gpt-6-astra" {
					t.Fatalf("summary models=%+v, want only gpt-6-astra", ss[0].Models)
				}
				Reset()
				if got := Calls(time.Time{}); !reflect.DeepEqual(cs, got) {
					t.Fatal("snapshot state changed after reload")
				}
			})
		}
	}
}

func TestCodexUsageSnapshotNeedsSameBoundary(t *testing.T) {
	u := cxUse(100, 50, 0, 10, 0)
	for _, kind := range []string{"different_time", "different_thread", "missing_field", "not_paginated", "new_response"} {
		t.Run(kind, func(t *testing.T) {
			meta, compacted := snapshotMeta(), snapshotCompaction()
			zero := tokenCountLine(0, cxUse(900, 450, 0, 90, 0), cxUse(0, 0, 0, 0, 0))
			switch kind {
			case "different_time":
				zero = strings.Replace(zero, stamp(0), stamp(1), 1)
			case "different_thread":
				compacted = strings.Replace(compacted, `"thread_id":"child"`, `"thread_id":"another"`, 1)
			case "missing_field":
				zero = strings.ReplaceAll(zero, `"reasoning_output_tokens":0,`, "")
			case "not_paginated":
				meta = strings.Replace(meta, `"paginated"`, `"full"`, 1)
			}
			lines := []string{meta, compacted}
			if kind == "new_response" {
				lines = append(lines, cxRecordLine(0, "intervening", cxUse(1010, 505, 0, 101, 0), cxUse(10, 5, 0, 1, 0)))
			}
			lines = append(lines, zero, cxRecordLine(2, "a", cxUse(1100, 550, 0, 110, 0), u), tokenCountLine(2, cxUse(1000, 500, 0, 100, 0), u))
			codexUsageLog(t, lines)
			want := 3
			if kind == "new_response" {
				want++
			}
			if cs := Calls(time.Time{}); len(cs) != want {
				t.Fatalf("unproven offset accepted: %+v", cs)
			}
		})
	}
}

// Early usage recorded before a rollout's first turn_context or
// thread_settings_applied (such as a guardian subagent's initial compacted
// snapshot or an orphan fork's leading token_count lines, #7818) is
// backfilled with the rollout's first named model in both the per-call and
// summary readers, leaving no empty-model unpriced entry.
func TestCodexEarlyUsageBeforeFirstTurnContextBackfillsModel(t *testing.T) {
	t.Run("guardian compacted before turn_context", func(t *testing.T) {
		for _, split := range []int{0, 4} {
			t.Run(fmt.Sprintf("split=%d", split), func(t *testing.T) {
				lines := []string{
					fmt.Sprintf(`{"timestamp":"%s","type":"session_meta","payload":{"id":"g1","source":{"subagent":{"other":"guardian"}},"model_provider":"openai"}}`, stamp(0)),
					cxCompactionLine(0, "cmp-1", cxUse(2695553, 2390528, 0, 2386, 0), cxUse(251432, 4864, 0, 1473, 0)),
					fmt.Sprintf(`{"timestamp":"%s","type":"event_msg","payload":{"type":"thread_settings_applied"}}`, stamp(0)),
					tokenCountLine(0, cxUse(2695553, 2390528, 0, 2386, 0), cxUse(0, 0, 0, 0, 0)),
					fmt.Sprintf(`{"timestamp":"%s","type":"event_msg","payload":{"type":"task_started","turn_id":"t1"}}`, stamp(0)),
					fmt.Sprintf(`{"timestamp":"%s","type":"turn_context","payload":{"turn_id":"t1","model":"codex-auto-review","effort":"low"}}`, stamp(1)),
				}
				var path string
				if split == 0 {
					path = codexUsageLog(t, lines)
				} else {
					path = codexUsageLog(t, lines[:split])
					Calls(time.Time{})
					List(0)
					Reset()
					appendText(t, path, strings.Join(lines[split:], "\n")+"\n")
				}
				cs := Calls(time.Time{})
				if len(cs) != 1 || cs[0].Model != "codex-auto-review" {
					t.Fatalf("calls=%+v, want 1 call with model codex-auto-review", cs)
				}
				ss := List(0)
				if len(ss) != 1 || len(ss[0].Models) != 1 || ss[0].Models[0].Model != "codex-auto-review" {
					t.Fatalf("summary=%+v, want only codex-auto-review", ss)
				}
				Reset()
				if got := Calls(time.Time{}); !reflect.DeepEqual(cs, got) {
					t.Fatalf("calls changed after reload: got %+v, want %+v", got, cs)
				}
			})
		}
	})

	t.Run("early token_count before turn_context keeps later model switch", func(t *testing.T) {
		for _, split := range []int{0, 3} {
			t.Run(fmt.Sprintf("split=%d", split), func(t *testing.T) {
				u1 := cxUse(179015, 177920, 0, 218, 0)
				u2 := cxUse(225290, 224000, 0, 110, 0)
				u3 := cxUse(1000, 500, 0, 50, 0)
				lines := []string{
					fmt.Sprintf(`{"timestamp":"%s","type":"session_meta","payload":{"id":"fork2","forked_from_id":"missing-parent","model_provider":"openai"}}`, stamp(0)),
					tokenCountLine(1, cxUse(30783290, 29000000, 0, 10000, 0), u1),
					tokenCountLine(2, cxUse(33698658, 31000000, 0, 10110, 0), u2),
					fmt.Sprintf(`{"timestamp":"%s","type":"turn_context","payload":{"turn_id":"t1","model":"gpt-5.6-sol","effort":"high"}}`, stamp(5)),
					fmt.Sprintf(`{"timestamp":"%s","type":"turn_context","payload":{"turn_id":"t2","model":"gpt-6-luna","effort":"high"}}`, stamp(6)),
					tokenCountLine(7, cxUse(33699658, 31000500, 0, 10160, 0), u3),
				}
				if split == 0 {
					codexUsageLog(t, lines)
				} else {
					path := codexUsageLog(t, lines[:split])
					Calls(time.Time{})
					List(0)
					Reset()
					appendText(t, path, strings.Join(lines[split:], "\n")+"\n")
				}
				cs := Calls(time.Time{})
				if len(cs) != 3 || cs[0].Model != "gpt-6-luna" || cs[1].Model != "gpt-5.6-sol" || cs[2].Model != "gpt-5.6-sol" {
					t.Fatalf("calls=%+v", cs)
				}
				ss := List(0)
				if len(ss) != 1 || len(ss[0].Models) != 2 {
					t.Fatalf("summary=%+v", ss)
				}
				for _, m := range ss[0].Models {
					if m.Model == "" {
						t.Fatalf("empty model remained in summary: %+v", ss[0].Models)
					}
				}
			})
		}
	})

	t.Run("thread_settings_applied backfills early compaction", func(t *testing.T) {
		lines := []string{
			fmt.Sprintf(`{"timestamp":"%s","type":"session_meta","payload":{"id":"g2","model_provider":"openai"}}`, stamp(0)),
			cxCompactionLine(0, "cmp-2", cxUse(5000, 1000, 0, 200, 0), cxUse(500, 100, 0, 20, 0)),
			fmt.Sprintf(`{"timestamp":"%s","type":"event_msg","payload":{"type":"thread_settings_applied","thread_settings":{"model":"codex-auto-review"}}}`, stamp(1)),
		}
		codexUsageLog(t, lines)
		cs := Calls(time.Time{})
		if len(cs) != 1 || cs[0].Model != "codex-auto-review" {
			t.Fatalf("calls=%+v", cs)
		}
		ss := List(0)
		if len(ss) != 1 || len(ss[0].Models) != 1 || ss[0].Models[0].Model != "codex-auto-review" {
			t.Fatalf("summary=%+v", ss)
		}
	})
}

// A depth-2 fork whose parent is a paginated rollout (history_base pointing
// back to the root thread) filters checkpoints inherited from that ancestor
// chain too (#7818).
func TestCodexDepth2ForkOfPaginatedParentFiltersRootCheckpoints(t *testing.T) {
	d := setupCalls(t)
	const rootID = "019d186b-a100-7aaa-8bbb-000000000001"
	const midID = "019d186b-a200-7aaa-8bbb-000000000002"
	const leafID = "019d186b-a300-7aaa-8bbb-000000000003"
	uRoot := cxUse(41200, 3072, 0, 812, 256)
	uOwn := cxUse(5000, 2000, 0, 100, 10)
	dir := d.codex + "/sessions/2026/03/23"
	writeLines(t, dir+"/rollout-2026-03-23T10-00-00-"+rootID+".jsonl",
		fmt.Sprintf(`{"timestamp":"%s","type":"session_meta","payload":{"id":%q,"model_provider":"openai"}}`, stamp(0), rootID),
		fmt.Sprintf(`{"timestamp":"%s","type":"turn_context","payload":{"turn_id":"t-root","model":"gpt-5.6-sol"}}`, stamp(1)),
		tokenCountLine(2, uRoot, uRoot),
	)
	writeLines(t, dir+"/rollout-2026-03-23T10-01-00-"+midID+".jsonl",
		fmt.Sprintf(`{"timestamp":"%s","type":"session_meta","payload":{"id":%q,"history_mode":"paginated","history_base":{"thread_id":%q,"end_ordinal_exclusive":3,"end_byte_offset":300},"model_provider":"openai"}}`, stamp(3), midID, rootID),
		fmt.Sprintf(`{"timestamp":"%s","type":"turn_context","payload":{"turn_id":"t-mid","model":"gpt-5.6-sol"}}`, stamp(4)),
	)
	writeLines(t, dir+"/rollout-2026-03-23T10-02-00-"+leafID+".jsonl",
		fmt.Sprintf(`{"timestamp":"%s","type":"session_meta","payload":{"id":%q,"forked_from_id":%q,"model_provider":"openai"}}`, stamp(5), leafID, midID),
		tokenCountLine(5, uRoot, uRoot),
		fmt.Sprintf(`{"timestamp":"%s","type":"turn_context","payload":{"turn_id":"t-leaf","model":"gpt-5.6-sol"}}`, stamp(6)),
		tokenCountLine(7, cxUse(46200, 5072, 0, 912, 266), uOwn),
	)
	var leafCalls []Call
	for _, c := range Calls(time.Time{}) {
		if strings.Contains(c.File, leafID) {
			leafCalls = append(leafCalls, c)
		}
	}
	want := Tokens{Input: 3000, CacheRead: 2000, Output: 100}
	if len(leafCalls) != 1 || leafCalls[0].Tokens != want || leafCalls[0].Model != "gpt-5.6-sol" {
		t.Fatalf("leaf calls = %+v, want 1 own call %+v", leafCalls, want)
	}
}
