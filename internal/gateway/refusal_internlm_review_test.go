package gateway

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/yetone/magpie/internal/provider"
)

func TestInternLMRefusalRequiresTheErrorCode(t *testing.T) {
	for _, c := range []struct {
		name, body string
		refused    bool
	}{
		{"code", `{"code":"content_check_rejected","message":"Blocked"}`, true},
		{"type only", `{"type":"content_check_error","message":"Blocked"}`, true},
		{"envelope", `{"error":{"code":"content_check_rejected","type":"content_check_error","message":"Blocked"}}`, true},
		{"different code", `{"code":"invalid_argument","message":"The request contains sensitive content is an example error message"}`, false},
		{"message only", `{"message":"The request contains sensitive content. Please modify your input and try again."}`, false},
		{"quoted code", `{"code":"invalid_argument","message":"content_check_rejected"}`, false},
		{"similar code", `{"code":"content_check_rejected_unsupported"}`, false},
		{"case differs", `{"code":"CONTENT_CHECK_REJECTED"}`, false},
		{"answer quoting error", `{"choices":[{"message":{"role":"assistant","content":"The request contains sensitive content"}}]}`, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			if note, ok := policyRefusal([]byte(c.body)); ok != c.refused {
				t.Fatalf("policyRefusal = %v (%q), want %v", ok, note, c.refused)
			}
		})
	}
}

// A recognized error must obey the user's existing zero-retry setting and
// leave the account usable, for both HTTP errors and streamed error frames.
func TestInternLMRefusalHonorsZeroRetries(t *testing.T) {
	const body = `{"message":"The request contains sensitive content. Please modify your input and try again.","type":"content_check_error","code":"content_check_rejected"}`
	for _, stream := range []bool{false, true} {
		t.Run(map[bool]string{false: "HTTP", true: "SSE"}[stream], func(t *testing.T) {
			fresh(t)
			refusalRetries(t, 0)
			var calls atomic.Int32
			up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if stream {
					w.Header().Set("Content-Type", "text/event-stream")
					io.WriteString(w, "data: {\"error\":"+body+"}\n\n")
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, body)
			}))
			t.Cleanup(up.Close)
			if err := provider.Save(provider.Provider{ID: "internlm", Name: "InternLM", Models: []string{"m"}, Chat: up.URL + "/v1", Routing: provider.Ordered,
				Key: "k1", Keys: []provider.KeyAccount{{Key: "k2"}, {Key: "k3"}}}); err != nil {
				t.Fatal(err)
			}
			s := New()
			status, got := sendTo(s, "/v1/chat/completions", `{"model":"internlm/m","messages":[{"role":"user","content":"hi"}]}`)
			if status != refusedStatus || !strings.Contains(got, "safety filter") || calls.Load() != 1 {
				t.Fatalf("status %d, calls %d: %s", status, calls.Load(), got)
			}
			route := lastRoute(s)
			if len(route.Tries) != 1 || route.Tries[0].Fail != failRefused || route.Tries[0].Rest != nil {
				t.Fatalf("tries: %+v", route.Tries)
			}
			if _, ok := restOf("internlm"); ok {
				t.Fatal("content refusal rested the provider")
			}
		})
	}
}
