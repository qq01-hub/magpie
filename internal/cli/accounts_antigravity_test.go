package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/plugin"
)

func TestAccountsAddAntigravityWithoutPluginExplainsHowToInstall(t *testing.T) {
	oldProviders, oldLogin := listPluginProviders, loginPlugin
	listPluginProviders = func() []plugin.Provider { return nil }
	loginPlugin = func(context.Context, string, string) error {
		t.Fatal("tried to sign in without an Antigravity plugin")
		return nil
	}
	t.Cleanup(func() {
		listPluginProviders, loginPlugin = oldProviders, oldLogin
	})

	err := accountsCmd([]string{"accounts", "add", "antigravity"})
	if err == nil || !strings.Contains(err.Error(), "magpie plugin add @magpie-community/opencode-agy-cli-auth") ||
		!strings.Contains(err.Error(), "magpie plugin login agy-cli") {
		t.Fatalf("missing Antigravity plugin error = %v", err)
	}
}

func TestAccountsAddAntigravityUsesInstalledAgyPlugin(t *testing.T) {
	oldProviders, oldLogin := listPluginProviders, loginPlugin
	listPluginProviders = func() []plugin.Provider {
		return []plugin.Provider{{ID: "agy-cli", Name: "Antigravity (local agy CLI)"}}
	}
	var got string
	loginPlugin = func(_ context.Context, id, method string) error {
		got = id + ":" + method
		return nil
	}
	t.Cleanup(func() {
		listPluginProviders, loginPlugin = oldProviders, oldLogin
	})

	if err := addAccount("antigravity"); err != nil {
		t.Fatal(err)
	}
	if got != "agy-cli:" {
		t.Fatalf("Antigravity sign-in used %q", got)
	}
}
