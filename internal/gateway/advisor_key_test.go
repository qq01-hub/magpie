package gateway

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/yetone/magpie/internal/access"
	"github.com/yetone/magpie/internal/provider"
)

// advisorKey is a gateway key whose models are models, and its secret.
func advisorKey(t *testing.T, models ...string) string {
	t.Helper()
	keys, secrets := newCaller(t, "Held")
	if _, err := access.Update("models-key", access.Change{Key: keys[0].ID, Models: models}); err != nil {
		t.Fatal(err)
	}
	return secrets[0]
}

func sendAs(t *testing.T, s *Server, secret, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest("POST", "/v1/messages?beta=true", strings.NewReader(body))
	req.Header.Set("User-Agent", "claude-cli/2.1.316 (external, cli)")
	req.Header.Set("Authorization", "Bearer "+secret)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	return rec
}

// Claude Code's advisor names a second model inside the request. A key
// held to some models may not reach another through it: magpie asks the
// advisor model as the caller (not as a model magpie chose), so the key's
// models hold it, and the advisor fails as an advisor does.
func TestAdvisorHeldToTheKeysModels(t *testing.T) {
	u := advisedSetup(t)
	secret := advisorKey(t, "v/*")
	body := strings.Replace(advisorTurn("v/glm-5.1", "adv/adv-pro"), `"stream":true`, `"stream":false`, 1)
	rec := sendAs(t, New(), secret, body)
	if rec.Code != 200 {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	u.mu.Lock()
	adv := len(u.advisor)
	u.mu.Unlock()
	if adv != 0 {
		t.Fatalf("the advisor model the key may not use was asked %d times", adv)
	}
	r := gjson.Parse(rec.Body.String())
	var result gjson.Result
	r.Get("content").ForEach(func(_, b gjson.Result) bool {
		if b.Get("type").String() == "advisor_tool_result" {
			result = b
		}
		return true
	})
	if result.Get("content.type").String() != "advisor_tool_result_error" {
		t.Fatalf("advisor result %s", rec.Body)
	}
}

// The same on a Claude provider whose upstream runs the advisor itself: a
// held key's advisor is sent upstream only when the key may use its model.
func TestNativeAdvisorHeldToTheKeysModels(t *testing.T) {
	fresh(t)
	u, url := newBillingUpstream(t)
	if err := provider.Save(provider.Provider{ID: "relay", Name: "Relay", Key: "k", Anthropic: url, Models: []string{"claude-opus-5-5", "claude-sonnet-5-5"}}); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		models  []string
		advisor string
		sent    bool
	}{
		{[]string{"relay/claude-sonnet-5-5"}, "relay/claude-opus-5-5", false},
		{[]string{"relay/claude-sonnet-5-5"}, "claude-opus-5-5", false},
		{[]string{"relay/claude-sonnet-5-5"}, "relay/claude-opus-5-5[1m]", false},
		{[]string{"relay/*"}, "relay/claude-opus-5-5", true},
	} {
		secret := advisorKey(t, c.models...)
		rec := sendAs(t, New(), secret, advisorTurn("relay/claude-sonnet-5-5", c.advisor))
		if rec.Code != 200 {
			t.Fatalf("%v %s: %d %s", c.models, c.advisor, rec.Code, rec.Body)
		}
		b := u.last(t)
		sent := false
		for _, tl := range gjson.Get(b, "tools").Array() {
			sent = sent || strings.HasPrefix(tl.Get("type").String(), "advisor_")
		}
		if sent != c.sent {
			t.Errorf("key %v, advisor %s: sent upstream %t, want %t: %s", c.models, c.advisor, sent, c.sent, b)
		}
	}
}

// A held key's model list applies after unprefixing Claude Desktop's advisor
// aliases too, both on Anthropic passthrough and on magpie-served turns.
func TestAdvisorDesktopAliasHeldToTheKeysModels(t *testing.T) {
	u := advisedSetup(t)
	if err := provider.SetModelEfforts("adv/adv-pro", []string{"low", "high"}); err != nil {
		t.Fatal(err)
	}
	var advAlias string
	shown, _ := provider.CatalogFor("claude-desktop")
	for _, e := range shown {
		if e.ID == "adv/adv-pro" {
			advAlias = DesktopID(e)
		}
	}
	if !strings.HasPrefix(advAlias, "mythos-magpie-") {
		t.Fatalf("unexpected adv alias %q", advAlias)
	}
	secret := advisorKey(t, "v/*")
	body := strings.Replace(advisorTurn("v/glm-5.1", advAlias+":high"), `"stream":true`, `"stream":false`, 1)
	rec := sendAs(t, New(), secret, body)
	if rec.Code != 200 {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	u.mu.Lock()
	adv := len(u.advisor)
	u.mu.Unlock()
	if adv != 0 {
		t.Fatalf("the advisor alias the key may not use was asked %d times", adv)
	}

	fresh(t)
	nu, url := newBillingUpstream(t)
	if err := provider.Save(provider.Provider{ID: "relay", Name: "Relay", Key: "k", Anthropic: url, Models: []string{"claude-opus-5-5", "claude-sonnet-5-5"}}); err != nil {
		t.Fatal(err)
	}
	if err := provider.SetModelEfforts("relay/claude-opus-5-5", []string{"low", "high"}); err != nil {
		t.Fatal(err)
	}
	var relayOpusAlias string
	shown, _ = provider.CatalogFor("claude-desktop")
	for _, e := range shown {
		if e.ID == "relay/claude-opus-5-5" {
			relayOpusAlias = DesktopID(e)
		}
	}
	if !strings.HasPrefix(relayOpusAlias, "magpie-") {
		t.Fatalf("unexpected relay opus alias %q", relayOpusAlias)
	}
	for _, c := range []struct {
		models []string
		sent   bool
	}{
		{[]string{"relay/claude-sonnet-5-5"}, false},
		{[]string{"relay/claude-sonnet-5-5", "relay/claude-opus-5-5"}, true},
		{[]string{"relay/*"}, true},
	} {
		sec := advisorKey(t, c.models...)
		resp := sendAs(t, New(), sec, advisorTurn("relay/claude-sonnet-5-5", relayOpusAlias+":high"))
		if resp.Code != 200 {
			t.Fatalf("%v: %d %s", c.models, resp.Code, resp.Body)
		}
		b := nu.last(t)
		sent := false
		for _, tl := range gjson.Get(b, "tools").Array() {
			sent = sent || strings.HasPrefix(tl.Get("type").String(), "advisor_")
		}
		if sent != c.sent {
			t.Errorf("key %v, alias %s: sent upstream %t, want %t: %s", c.models, relayOpusAlias, sent, c.sent, b)
		}
	}
}
