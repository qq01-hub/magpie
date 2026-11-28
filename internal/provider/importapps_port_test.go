package provider

import (
	"net/url"
	"path/filepath"
	"testing"

	"github.com/yetone/magpie/internal/settings"
)

// magpie's own gateway is on the port Settings has (Magic_zero on
// Discord): an app's entry there is magpie's, not one to import, and one
// on 3425 is another program's once magpie has moved off it.
func TestGatewayURLOnSettingsPort(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("MAGPIE_ADDR", "")
	is := func(u string) bool { p, _ := url.Parse(u); return gatewayURL(p) }
	if !is("http://127.0.0.1:3425/v1") || is("http://127.0.0.1:3591/v1") {
		t.Fatal("the default port")
	}
	s := settings.Load()
	s.Port = 3591
	if err := settings.Save(s); err != nil {
		t.Fatal(err)
	}
	if !is("http://127.0.0.1:3591/v1") || !is("http://localhost:3591") || !is("http://[::1]:3591/v1") || !is("http://[::ffff:127.0.0.1]:3591/v1") || is("http://127.0.0.1:3425/v1") || is("http://10.0.0.2:3591") {
		t.Fatal("Settings' port")
	}
	t.Setenv("MAGPIE_ADDR", "127.0.0.1:3426")
	if !is("http://127.0.0.1:3426") {
		t.Fatal("MAGPIE_ADDR's port")
	}
}

func TestImportedAcceptsLANPortAndIPv6(t *testing.T) {
	for _, tc := range []struct {
		name, base, skip string
	}{
		{name: "single-label LAN host", base: "http://codex-relay:8080/v1"},
		{name: "IPv6 literal", base: "http://[fd00::5]:8080/v1"},
		{name: "single-label host without port", base: "http://codex-relay/v1", skip: "its base URL isn't valid"},
		{name: "port out of range", base: "http://codex-relay:65536/v1", skip: "its base URL isn't valid"},
		{name: "bracketed non-IP host", base: "http://[bad]:8080/v1", skip: "its base URL isn't valid"},
		{name: "malformed IPv6", base: "http://[fd00::5/v1", skip: "its base URL isn't valid"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, skip := imported(tc.name, "sk-x", endpoints{responses: tc.base}, nil)
			if skip != tc.skip {
				t.Fatalf("skip = %q, want %q", skip, tc.skip)
			}
			if tc.skip == "" && p.Responses != tc.base {
				t.Fatalf("responses = %q, want %q", p.Responses, tc.base)
			}
		})
	}
}
