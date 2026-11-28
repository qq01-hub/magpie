package gateway

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/yetone/magpie/internal/provider"
	"github.com/yetone/magpie/internal/usage"
)

func TestAdvisorDesktopTierRespectsCallerAndExplicitModels(t *testing.T) {
	fresh(t)
	was := usage.PluginAgents
	usage.PluginAgents = func() []usage.Known { return []usage.Known{{ID: "openbot", Names: []string{"openbot"}}} }
	t.Cleanup(func() { usage.PluginAgents = was })
	u, url := newBillingUpstream(t)
	if err := provider.Save(provider.Provider{ID: "relay", Name: "Relay", Key: "k", Anthropic: url, Models: []string{"claude-sonnet-5-5", "claude-opus-5-5", "claude-opus-4-6"}}); err != nil {
		t.Fatal(err)
	}
	if err := SetDesktopTier("opus", "relay/claude-opus-4-6"); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, key, path, ua, advisor, want string
	}{
		{"desktop tier", Token, "/v1/messages", "claude-cli/2.1.321 (external, claude-desktop)", "claude-opus-5-5", "claude-opus-4-6"},
		{"desktop tier 2M effort", Token, "/v1/messages", "claude-cli/2.1.321 (external, claude-desktop)", "claude-opus-5-5:high[2m]", "claude-opus-4-6"},
		{"explicit provider", Token, "/v1/messages", "claude-cli/2.1.321 (external, claude-desktop)", "relay/claude-opus-5-5", "claude-opus-5-5"},
		{"explicit provider 2M effort", Token, "/v1/messages", "claude-cli/2.1.321 (external, claude-desktop)", "relay/claude-opus-5-5:high[2m]", "claude-opus-5-5"},
		{"CLI caller", TokenFor("claude"), "/v1/messages", "claude-cli/2.1.321 (external, claude-desktop)", "claude-opus-5-5", "claude-opus-5-5"},
		{"path caller", Token, "/agent/openbot/v1/messages", "claude-cli/2.1.321 (external, claude-desktop)", "claude-opus-5-5", "claude-opus-5-5"},
		{"unrelated UA", Token, "/v1/messages", "curl/1 claude-desktop-note", "claude-opus-5-5", "claude-opus-5-5"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tc.path, strings.NewReader(advisorTurn("relay/claude-sonnet-5-5", tc.advisor)))
			req.RemoteAddr = "127.0.0.1:5000"
			req.Header.Set("Authorization", "Bearer "+tc.key)
			req.Header.Set("User-Agent", tc.ua)
			rec := httptest.NewRecorder()
			lanGuard(New().Handler()).ServeHTTP(rec, req)
			if rec.Code != 200 {
				t.Fatalf("%d %s", rec.Code, rec.Body.String())
			}
			if got := gjson.Get(u.last(t), "tools.1.model").String(); got != tc.want {
				t.Fatalf("advisor sent as %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBridgeRequestResolvesDesktopAdvisorAliases(t *testing.T) {
	advisedSetup(t)
	if err := SetDesktopTier("opus", "adv/adv-pro"); err != nil {
		t.Fatal(err)
	}
	alias := ""
	shown, _ := provider.CatalogFor("claude-desktop")
	for _, e := range shown {
		if e.ID == "adv/adv-pro" {
			alias = DesktopID(e)
		}
	}
	if alias == "" {
		t.Fatal("advisor missing from Desktop catalog")
	}
	for _, advisor := range []string{alias, alias + ":high[1m]", alias + ":high[2m]", "claude-opus-5-5", "adv~adv-pro[2m]"} {
		t.Run(advisor, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/v1/messages", nil)
			r.Header.Set("User-Agent", "claude-cli/2.1.321 (external, claude-desktop)")
			req, err := bridgeRequest(r, provider.Anthropic, "Claude Code", []byte(advisorTurn("claude/claude-sonnet-5-5", advisor)))
			if err != nil {
				t.Fatal(err)
			}
			if req.Advisor != "adv/adv-pro" || !strings.Contains(allText(req), advisorNote) || strings.Contains(allText(req), advisorGone) {
				t.Fatalf("advisor=%q, instructions=%q", req.Advisor, allText(req))
			}
		})
	}
}

func TestAdvisorDesktopTierPickSuffixes(t *testing.T) {
	u := advisedSetup(t)
	alias := ""
	shown, _ := provider.CatalogFor("claude-desktop")
	for _, e := range shown {
		if e.ID == "adv/adv-pro" {
			alias = DesktopID(e)
		}
	}
	if alias == "" {
		t.Fatal("advisor missing from Desktop catalog")
	}
	for _, tc := range []struct{ name, pick, advisor string }{
		{"tier pick 1M", "adv/adv-pro:high[1m]", "claude-opus-5-5"},
		{"tier pick 2M", "adv/adv-pro:high[2m]", "claude-opus-5-5"},
		{"alias 2M", "", alias + ":high[2m]"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := SetDesktopTier("opus", tc.pick); err != nil {
				t.Fatal(err)
			}
			for _, allowed := range []bool{false, true} {
				models := []string{"v/*"}
				want := "advisor_tool_result_error"
				if allowed {
					models = append(models, "adv/adv-pro")
					want = "advisor_result"
				}
				secret := advisorKey(t, models...)
				u.mu.Lock()
				before := len(u.advisor)
				u.mu.Unlock()
				body := strings.Replace(advisorTurn("v/glm-5.1", tc.advisor), `"stream":true`, `"stream":false`, 1)
				r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
				r.RemoteAddr = "127.0.0.1:5000"
				r.Header.Set("Authorization", "Bearer "+secret)
				r.Header.Set("User-Agent", "claude-cli/2.1.321 (external, claude-desktop)")
				rec := httptest.NewRecorder()
				lanGuard(New().Handler()).ServeHTTP(rec, r)
				result := ""
				for _, block := range gjson.Get(rec.Body.String(), "content").Array() {
					if block.Get("type").String() == "advisor_tool_result" {
						result = block.Get("content.type").String()
					}
				}
				u.mu.Lock()
				asked := len(u.advisor) - before
				u.mu.Unlock()
				if rec.Code != 200 || result != want || asked != map[bool]int{false: 0, true: 1}[allowed] {
					t.Errorf("allowed=%v status=%d result=%s asked=%d: %s", allowed, rec.Code, result, asked, rec.Body.String())
				}
			}
			r := httptest.NewRequest("POST", "/v1/messages", nil)
			r.Header.Set("User-Agent", "claude-cli/2.1.321 (external, claude-desktop)")
			req, err := bridgeRequest(r, provider.Anthropic, "Claude Code", []byte(advisorTurn("claude/claude-sonnet-5-5", tc.advisor)))
			if err != nil {
				t.Fatal(err)
			}
			if req.Advisor != "adv/adv-pro" || strings.Contains(allText(req), advisorGone) {
				t.Errorf("bridge advisor=%q, instructions=%q", req.Advisor, allText(req))
			}
		})
	}
}

func TestAdvisorDesktopTierAndStandInRespectKeyModels(t *testing.T) {
	u, url := advisedSetupWithURL(t)
	if err := provider.Save(provider.Provider{ID: "adv2", Name: "Adv2", Key: "k", Chat: url + "/v1", Models: []string{"adv-pro"}}); err != nil {
		t.Fatal(err)
	}
	if err := SetDesktopTier("opus", "adv/adv-pro"); err != nil {
		t.Fatal(err)
	}
	if err := provider.SetAutoGroups(false); err != nil {
		t.Fatal(err)
	}
	for _, advisor := range []string{"claude-opus-5-5", "auto/adv-pro"} {
		for _, allowed := range []bool{false, true} {
			models := []string{"v/*"}
			want := "advisor_tool_result_error"
			if allowed {
				models = append(models, "adv/adv-pro", "adv2/adv-pro")
				want = "advisor_result"
			}
			secret := advisorKey(t, models...)
			u.mu.Lock()
			before := len(u.advisor)
			u.mu.Unlock()
			body := strings.Replace(advisorTurn("v/glm-5.1", advisor), `"stream":true`, `"stream":false`, 1)
			r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
			r.Header.Set("Authorization", "Bearer "+secret)
			r.Header.Set("User-Agent", "claude-cli/2.1.321 (external, claude-desktop)")
			rec := httptest.NewRecorder()
			New().Handler().ServeHTTP(rec, r)
			result := ""
			for _, block := range gjson.Get(rec.Body.String(), "content").Array() {
				if block.Get("type").String() == "advisor_tool_result" {
					result = block.Get("content.type").String()
				}
			}
			u.mu.Lock()
			asked := len(u.advisor) - before
			u.mu.Unlock()
			if rec.Code != 200 || result != want || asked != map[bool]int{false: 0, true: 1}[allowed] {
				t.Fatalf("advisor=%s allowed=%v status=%d result=%s asked=%d: %s", advisor, allowed, rec.Code, result, asked, rec.Body.String())
			}
		}
	}
}
