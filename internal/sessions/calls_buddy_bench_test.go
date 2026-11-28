package sessions

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkBuddyCallsAppend expands CodeBuddy Code 2.164.0's real direct
// session bytes, warms its call shard, and measures one appended reply
// All fixture construction and appends are outside the measured interval
func BenchmarkBuddyCallsAppend(b *testing.B) {
	for _, mb := range []int{10, 100} {
		b.Run(fmt.Sprintf("%dMiB", mb), func(b *testing.B) {
			setupCalls(b)
			fixture, err := os.ReadFile(filepath.Join("testdata", "codebuddy", "projects", "Users-me-work", "01a12fea-4cae-78ac-8b2b-a1c0f9da9b1c.jsonl"))
			if err != nil {
				b.Fatal(err)
			}
			path := filepath.Join(b.TempDir(), "44444444-aaaa-4bbb-8ccc-000000000009.jsonl")
			if err := os.WriteFile(path, bytes.Repeat(fixture, (mb<<20)/len(fixture)+1), 0o600); err != nil {
				b.Fatal(err)
			}
			read := func() {
				f := file{agent: "codebuddy", path: path, key: "codebuddy:44444444-aaaa-4bbb-8ccc-000000000009", main: true}
				if !stat(&f) {
					b.Fatal("session not listed")
				}
				if got := readCalls(f).Calls; len(got) == 0 {
					b.Fatal("fixture produced no calls")
				}
			}
			read()
			f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
			if err != nil {
				b.Fatal(err)
			}
			defer f.Close()
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				line := fmt.Sprintf(`{"timestamp":1790503210000,"type":"message","role":"assistant","providerData":{"messageId":"append-%d","model":"fake-m1","usage":{"inputTokens":100,"outputTokens":2}}}`, i) + "\n"
				if _, err := f.WriteString(line); err != nil {
					b.Fatal(err)
				}
				b.StartTimer()
				read()
			}
		})
	}
}
