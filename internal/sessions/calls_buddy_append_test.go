package sessions

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
)

// Each Buddy app appends reply updates to its JSONL file. Reading an append
// must not parse the earlier replies again, and a later line of the last
// reply replaces that call even after the continuation was read from disk
func TestBuddyCallsAppendAndResume(t *testing.T) {
	for _, agent := range []string{"workbuddy", "workbuddy-ai", "codebuddy"} {
		t.Run(agent, func(t *testing.T) {
			setup(t)
			path := filepath.Join(t.TempDir(), "44444444-aaaa-4bbb-8ccc-000000000009.jsonl")
			line := func(id string, input int) string {
				return fmt.Sprintf(`{"id":"a9","timestamp":1790503210000,"type":"message","role":"assistant","providerData":{"messageId":%q,"model":"hy3","requestModelId":"hy3","usage":{"inputTokens":%d,"outputTokens":0}},"sessionId":"44444444-aaaa-4bbb-8ccc-000000000009","cwd":"/work/grow"}`, id, input)
			}
			write := func(text string) {
				t.Helper()
				if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			initial := line("a", 100) + "\n" + line("b", 200) + "\n"
			write(initial)
			var mu sync.Mutex
			var starts []int64
			scanHook = func(p string, off int64) {
				if p == path {
					mu.Lock()
					starts = append(starts, off)
					mu.Unlock()
				}
			}
			t.Cleanup(func() { scanHook = nil })
			read := func() *callFile {
				t.Helper()
				f := file{agent: agent, path: path, key: agent + ":44444444-aaaa-4bbb-8ccc-000000000009", main: true}
				if !stat(&f) {
					t.Fatal("session not listed")
				}
				mu.Lock()
				starts = nil
				mu.Unlock()
				return readCalls(f)
			}
			check := func(st *callFile, want ...int) {
				t.Helper()
				var got []int
				for _, c := range st.Calls {
					if c.Agent != agent || c.Session != "44444444-aaaa-4bbb-8ccc-000000000009" || c.Model != "hy3" {
						t.Fatalf("lost call identity: %+v", c)
					}
					got = append(got, c.Input)
				}
				if !slices.Equal(got, want) {
					t.Fatalf("call inputs %v, want %v", got, want)
				}
			}
			from := func(want int64) {
				t.Helper()
				mu.Lock()
				got := slices.Clone(starts)
				mu.Unlock()
				if !slices.Equal(got, []int64{want}) {
					t.Fatalf("session scans %v, want only the append from byte %d", got, want)
				}
			}
			before := read()
			check(before, 100, 200)
			end := int64(len(initial))
			appendText(t, path, line("b", 300)+"\n")
			check(read(), 100, 300)
			from(end)
			check(before, 100, 200) // the already published read is immutable
			end += int64(len(line("b", 300)) + 1)

			resetCalls()
			appendText(t, path, line("b", 400)+"\n")
			check(read(), 100, 400)
			from(end)
			resetCalls()
			check(read(), 100, 400) // the replacement is in the shard, too
			end += int64(len(line("b", 400)) + 1)

			// A zero-token usage line still ends the previous reply's run
			text := line("zero", 0) + "\n" + line("b", 500) + "\n"
			appendText(t, path, text)
			check(read(), 100, 400, 500)
			from(end)
			end += int64(len(text))

			partial := line("next", 600)
			appendText(t, path, partial[:70])
			check(read(), 100, 400, 500)
			from(end)
			appendText(t, path, partial[70:]+"\n")
			check(read(), 100, 400, 500, 600)
			from(end)

			write(line("short", 700) + "\n")
			check(read(), 700)
			from(0)
			write(line("overwritten", 800) + "\n" + line("longer", 900) + "\n")
			check(read(), 800, 900)
			from(0)
		})
	}
}

func TestBuddyCallsWaitForWholeLine(t *testing.T) {
	setup(t)
	path := filepath.Join(t.TempDir(), "44444444-aaaa-4bbb-8ccc-000000000009.jsonl")
	line := `{"timestamp":1790503210000,"type":"message","role":"assistant","providerData":{"messageId":"pending","model":"hy3","usage":{"inputTokens":100,"outputTokens":2}}}`
	if err := os.WriteFile(path, []byte(line), 0o600); err != nil {
		t.Fatal(err)
	}
	f := file{agent: "workbuddy", path: path, key: "workbuddy:44444444-aaaa-4bbb-8ccc-000000000009", main: true}
	if !stat(&f) {
		t.Fatal("session not listed")
	}
	if got := readCalls(f).Calls; len(got) != 0 {
		t.Fatalf("counted a reply before its line was complete: %+v", got)
	}
	appendText(t, path, "\n")
	if !stat(&f) {
		t.Fatal("session not listed after newline")
	}
	if got := readCalls(f).Calls; len(got) != 1 || got[0].Input != 100 || got[0].Output != 2 {
		t.Fatalf("completed reply: %+v", got)
	}
}

// The old whole-file reader rechecked older replies against models.json on
// each append. Keep that behavior when a model changes gateway ownership
func TestBuddyCallsRecheckGatewayOnAppend(t *testing.T) {
	for _, agent := range []string{"workbuddy", "workbuddy-ai", "codebuddy"} {
		t.Run(agent, func(t *testing.T) {
			setup(t)
			dir := buddyDir(agent)
			if err := os.MkdirAll(dir, 0o700); err != nil {
				t.Fatal(err)
			}
			models := filepath.Join(dir, "models.json")
			if err := os.WriteFile(models, []byte(`[{"id":"relay/hy3","vendor":"magpie"}]`), 0o600); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "44444444-aaaa-4bbb-8ccc-000000000009.jsonl")
			line := `{"timestamp":1790503210000,"type":"message","role":"assistant","providerData":{"messageId":"a","model":"hy3","requestModelId":"custom-local:relay/hy3","usage":{"inputTokens":100,"outputTokens":2}}}` + "\n"
			if err := os.WriteFile(path, []byte(line), 0o600); err != nil {
				t.Fatal(err)
			}
			read := func() []Call {
				t.Helper()
				f := file{agent: agent, path: path, key: agent + ":44444444-aaaa-4bbb-8ccc-000000000009", main: true}
				if !stat(&f) {
					t.Fatal("session not listed")
				}
				return readCalls(f).Calls
			}
			if got := read(); len(got) != 0 {
				t.Fatalf("gateway call counted locally: %+v", got)
			}
			if err := os.WriteFile(models, []byte(`[{"id":"relay/hy3","vendor":"openai"}]`), 0o600); err != nil {
				t.Fatal(err)
			}
			appendText(t, path, "{}\n")
			if got := read(); len(got) != 1 || got[0].Input != 100 {
				t.Fatalf("the model's direct call stayed excluded after configuration changed: %+v", got)
			}
		})
	}
}

func TestBuddyCallVersionRebuildsOldShards(t *testing.T) {
	for _, agent := range []string{"workbuddy", "workbuddy-ai", "codebuddy"} {
		t.Run(agent, func(t *testing.T) {
			setup(t)
			path := filepath.Join(t.TempDir(), "44444444-aaaa-4bbb-8ccc-000000000009.jsonl")
			line := `{"timestamp":1790503210000,"type":"message","role":"assistant","providerData":{"messageId":"a","model":"hy3","usage":{"inputTokens":100,"outputTokens":2}}}` + "\n"
			if err := os.WriteFile(path, []byte(line), 0o600); err != nil {
				t.Fatal(err)
			}
			f := file{agent: agent, path: path, key: agent + ":44444444-aaaa-4bbb-8ccc-000000000009", main: true}
			if !stat(&f) {
				t.Fatal("session not listed")
			}
			current := callParseVersions
			callParseVersions = maps.Clone(current)
			callParseVersions[agent] = 0
			t.Cleanup(func() { callParseVersions = current })
			if got := readCalls(f).Calls; len(got) != 1 || got[0].Input != 100 {
				t.Fatalf("old parser shard: %+v", got)
			}
			resetCalls()
			callParseVersions = current
			var mu sync.Mutex
			var scans []int64
			scanHook = func(p string, off int64) {
				if p == path {
					mu.Lock()
					scans = append(scans, off)
					mu.Unlock()
				}
			}
			t.Cleanup(func() { scanHook = nil })
			if got := readCalls(f).Calls; len(got) != 1 || got[0].Input != 100 {
				t.Fatalf("rebuilt parser shard: %+v", got)
			}
			mu.Lock()
			got := slices.Clone(scans)
			scans = nil
			mu.Unlock()
			if !slices.Equal(got, []int64{0}) {
				t.Fatalf("version-0 shard reused without rebuilding: scans %v", got)
			}
			resetCalls()
			readCalls(f)
			mu.Lock()
			got = slices.Clone(scans)
			mu.Unlock()
			if len(got) != 0 {
				t.Fatalf("current shard rebuilt again: scans %v", got)
			}
		})
	}
}
