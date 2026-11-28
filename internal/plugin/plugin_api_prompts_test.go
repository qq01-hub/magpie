package plugin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/catalog"
)

// An "api" method whose own questions ask its key (inkstone's API key way,
// Devin's session token in plugins#928) was asked a key of magpie's beside
// them, which went nowhere, and every answer was copied into the
// account's metadata, so the token sat there a second time in plain text.
// Such a method is marked keyAsked, signs in with no key, and keeps only
// its choices' answers beside what its authorize gave; a method with no
// authorize still keeps its answers as OpenCode does.
func TestAPIPromptsKeyAskedAndMetadata(t *testing.T) {
	sandbox(t)
	catalog.Reset()
	t.Cleanup(catalog.Reset)
	if err := os.MkdirAll(filepath.Dir(catalog.CachePath()), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(catalog.CachePath(), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	abs, _ := filepath.Abs("testdata/apiprompts")
	if _, err := Add(ctx, abs); err != nil {
		t.Fatal(err)
	}
	ps, err := Providers(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var ms []Method
	for _, p := range ps {
		if p.ID == "apiprompts" {
			ms = p.Methods
		}
	}
	if len(ms) != 5 || !ms[0].KeyAsked || !ms[0].NoKey() || ms[1].KeyAsked || ms[1].NoKey() || ms[2].KeyAsked || ms[2].NoKey() || ms[3].NoKey() || ms[4].NoKey() {
		t.Fatalf("methods = %+v, want only the first keyAsked", ms)
	}

	if _, err := APIKey(ctx, "apiprompts", 0, map[string]string{"site": "eu", "token": "tok-secret", "note": "cookie-secret"}, "", NewAccount); err != nil {
		t.Fatal(err)
	}
	if _, err := APIKey(ctx, "apiprompts", 1, map[string]string{"tenant": "acme"}, "k-site", NewAccount); err != nil {
		t.Fatal(err)
	}
	if _, err := APIKey(ctx, "apiprompts", 2, nil, "k-own", NewAccount); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(AuthPath())
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(string(b), "tok-secret"); n != 1 {
		t.Errorf("the token is kept %d times, want once (as the key): %s", n, b)
	}
	if strings.Contains(string(b), "cookie-secret") {
		t.Errorf("a typed answer is kept in metadata: %s", b)
	}
	var auth map[string]struct {
		Type     string            `json:"type"`
		Key      string            `json:"key"`
		Metadata map[string]string `json:"metadata"`
	}
	if err := json.Unmarshal(b, &auth); err != nil {
		t.Fatal(err)
	}
	want := map[string]map[string]string{
		"tok-secret": {"site": "eu", "email": "who@apiprompts.invalid"},
		"k-site":     {"tenant": "acme"},
		"k-own":      {"checked": "yes"},
	}
	got := map[string]map[string]string{}
	for _, a := range auth {
		if a.Type == "api" {
			got[a.Key] = a.Metadata
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("accounts by key = %v, want %v", got, want)
	}

	// no key typed and none given by authorize: nothing is saved
	if _, err := APIKey(ctx, "apiprompts", 2, nil, "", NewAccount); err == nil {
		t.Error("a sign-in with no key typed and none given was taken")
	}
}

func TestAPINonsecretPromptsKeepTheKeyField(t *testing.T) {
	sandbox(t)
	catalog.Reset()
	t.Cleanup(catalog.Reset)
	if err := os.MkdirAll(filepath.Dir(catalog.CachePath()), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(catalog.CachePath(), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	abs, _ := filepath.Abs("testdata/apiprompts")
	if _, err := Add(ctx, abs); err != nil {
		t.Fatal(err)
	}
	ps, err := Providers(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var methods []Method
	for _, p := range ps {
		if p.ID == "apiprompts" {
			methods = p.Methods
		}
	}
	if len(methods) != 5 {
		t.Fatalf("methods: %+v", methods)
	}
	for _, method := range []int{3, 4} {
		// Follow the GUI/CLI decision: a NoKey method gets no key field
		key := ""
		if !methods[method].NoKey() {
			key = map[int]string{3: "system-token", 4: "api-token"}[method]
		}
		if _, err := APIKey(ctx, "apiprompts", method, map[string]string{"userID": "1234", "endpoint": "slb"}, key, NewAccount); err != nil {
			t.Errorf("nonsecret method %d lost its separate key field: %v", method, err)
		}
	}
	if t.Failed() {
		return
	}
	b, err := os.ReadFile(AuthPath())
	if err != nil {
		t.Fatal(err)
	}
	var auth map[string]struct {
		Key      string            `json:"key"`
		Metadata map[string]string `json:"metadata"`
	}
	if err := json.Unmarshal(b, &auth); err != nil {
		t.Fatal(err)
	}
	got := map[string]map[string]string{}
	for _, a := range auth {
		got[a.Key] = a.Metadata
	}
	want := map[string]map[string]string{
		"system-token": {"mode": "token", "userID": "1234", "endpoint": "slb", "email": "packyapi-1234"},
		"api-token":    {"mode": "keys", "endpoint": "slb"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("signed-in keys and metadata: %v, want %v", got, want)
	}
}
