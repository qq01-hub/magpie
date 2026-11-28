package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/catalog"
	"github.com/yetone/magpie/internal/gateway"
	"github.com/yetone/magpie/internal/provider"
)

func TestDshSetAndDropUseEffectiveStartEntry(t *testing.T) {
	_, _, web := dshRouteHome(t)
	first := "- id: agent-default-model\n  config:\n    provider: deepseek-official\n    model: first-user-model\n"
	effective := "- id: agent-default-model\n  name: '@deepseek-ai/dsh-agent-default-model'\n  config:\n    provider: deepseek-official\n    model: saved-user-model\n"
	decoys := "- id: agent-default-model\n  name: another-plugin\n  config:\n    provider: deepseek-official\n    model: mismatched-name-model\n" +
		"- id: agent-default-model\n  insert: []\n  config:\n    provider: deepseek-official\n    model: inserted-model\n"
	if err := os.WriteFile(web, []byte(first+effective+decoys), 0o644); err != nil {
		t.Fatal(err)
	}
	models := magpieModels("dsh")
	if err := dshSetFile(web, "magpie/deepseek/pro", true, models, gateway.URL()); err != nil {
		t.Fatal(err)
	}
	_, items, err := dshRead(web)
	if err != nil {
		t.Fatal(err)
	}
	i := 1
	if len(items) <= i {
		t.Fatal("no effective start after setting magpie model")
	}
	if got := dshConfig(items[i]); got["provider"] != magpieID || got["model"] != "deepseek/pro" {
		t.Fatalf("effective start after set: %v", got)
	}
	if strings.Join(items[0].lines, "\n") != strings.TrimSuffix(first, "\n") {
		t.Fatalf("the earlier duplicate changed: %v", items[0].lines)
	}
	if items[i+1].lines[1] != "  name: another-plugin" || items[i+2].lines[1] != "  insert: []" {
		t.Fatalf("non-matching entries changed: %v / %v", items[i+1].lines, items[i+2].lines)
	}

	if err := dshSetFile(web, "", true, models, gateway.URL()); err != nil {
		t.Fatal(err)
	}
	_, items, err = dshRead(web)
	if err != nil {
		t.Fatal(err)
	}
	i = 1
	if len(items) <= i {
		t.Fatalf("effective start after disconnect: index %d, entries %d", i, len(items))
	}
	if got := dshConfig(items[i]); got["model"] != "saved-user-model" || got["provider"] != "deepseek-official" {
		t.Fatalf("effective start after disconnect: index %d, config %v", i, got)
	}
}

func TestDshKeepPickUsesEffectiveStartEntry(t *testing.T) {
	_, _, web := dshRouteHome(t)
	first := "- id: agent-default-model\n  config:\n    provider: deepseek-official\n    model: earlier-user-model\n"
	effective := "- id: agent-default-model # magpie\n  config:\n    provider: deepseek-official\n    model: dsh-saved-model\n"
	decoys := "- id: agent-default-model\n  name: another-plugin\n  config:\n    provider: deepseek-official\n    model: mismatched-name-model\n" +
		"- id: agent-default-model\n  insert: []\n  config:\n    provider: deepseek-official\n    model: inserted-model\n"
	if err := os.WriteFile(web, []byte(first+effective+decoys), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := dshKeepPick(web); err != nil {
		t.Fatal(err)
	}
	_, items, err := dshRead(web)
	if err != nil {
		t.Fatal(err)
	}
	i := 1
	if len(items) <= i {
		t.Fatalf("effective dsh pick: index %d, entries %d", i, len(items))
	}
	if got := dshConfig(items[i]); dshOurs(items[i]) || got["model"] != "dsh-saved-model" {
		t.Fatalf("effective dsh pick: index %d, ours %v, config %v", i, dshOurs(items[i]), got)
	}
	if strings.Join(items[0].lines, "\n") != strings.TrimSuffix(first, "\n") {
		t.Fatalf("the earlier duplicate changed: %v", items[0].lines)
	}
}

func TestDshFillNewProfilesChecksEffectiveStartEntry(t *testing.T) {
	_, dir, web := dshRouteHome(t)
	if err := os.MkdirAll(filepath.Dir(web), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(web, []byte(dshRouteFixture("deepseek/pro", "deepseek/flash")+
		"- id: agent-default-model # magpie\n  config:\n    provider: magpie\n    model: deepseek/flash\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	desktop := filepath.Join(dir, "profiles", "desktop", "cordis.patch.yml")
	if err := os.MkdirAll(filepath.Dir(desktop), 0o755); err != nil {
		t.Fatal(err)
	}
	first := "- id: agent-default-model\n  config:\n    provider: deepseek-official\n    model: desktop-user-model\n"
	effective := "- id: agent-default-model\n  name: '@deepseek-ai/dsh-agent-default-model'\n  config:\n    provider: magpie\n    model: deepseek/pro\n"
	decoys := "- id: agent-default-model\n  name: another-plugin\n  config:\n    provider: deepseek-official\n    model: mismatched-name-model\n" +
		"- id: agent-default-model\n  insert: []\n  config:\n    provider: deepseek-official\n    model: inserted-model\n"
	if err := os.WriteFile(desktop, []byte(first+effective+decoys), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := dshFillNewProfiles([]string{web, desktop}, magpieModels("dsh"), gateway.URL()); err != nil {
		t.Fatal(err)
	}
	_, items, err := dshRead(desktop)
	if err != nil {
		t.Fatal(err)
	}
	i := 1
	if len(items) <= i {
		t.Fatalf("new profile effective start: index %d, entries %d", i, len(items))
	}
	if got := dshConfig(items[i]); got["model"] != "deepseek/flash" {
		t.Fatalf("new profile effective start: index %d, config %v", i, got)
	}
}

func TestDshFindStartMatcherBoundaries(t *testing.T) {
	raw := "- id: agent-default-model\n  config:\n    provider: deepseek-official\n    model: first\n" +
		"- id: agent-default-model\n  name: ''\n  config:\n    provider: deepseek-official\n    model: empty-name\n" +
		"- id: agent-default-model\n  name: '@deepseek-ai/dsh-agent-default-model'\n  config:\n    provider: deepseek-official\n    model: exact-name\n" +
		"- id: agent-default-model\n  name: '@deepseek-ai/dsh-agent-default-model-extra'\n  config:\n    provider: deepseek-official\n    model: suffix-name\n" +
		"- id: agent-default-model\n  insert: []\n  config:\n    provider: deepseek-official\n    model: inserted\n" +
		"- id: agent-default-model\n  name: another-plugin\n  config:\n    provider: deepseek-official\n    model: mismatched-name\n"
	_, items, err := dshParse("matcher", raw)
	if err != nil {
		t.Fatal(err)
	}
	if got := dshFindStart(items); got != 2 {
		t.Fatalf("effective start index = %d, want 2", got)
	}
}

func TestDshEffectiveEffortAndRouteSkipDecoys(t *testing.T) {
	if err := provider.Save(provider.Provider{ID: "deepseek", Name: "DeepSeek", Key: "k", Chat: "http://127.0.0.1:1/v1", Models: []string{"pro", "flash"}}); err != nil {
		t.Fatal(err)
	}
	if err := provider.Save(provider.Provider{ID: "spare", Name: "Spare", Key: "k", Chat: "http://127.0.0.1:1/v1", Models: []string{"pro"}}); err != nil {
		t.Fatal(err)
	}
	raw := "- id: agent-default-model\n  config:\n    provider: magpie\n    model: group/auto-pro\n    reasoningEffort: low\n" +
		"- id: agent-default-model\n  name: '@deepseek-ai/dsh-agent-default-model'\n  config:\n    provider: magpie\n    model: group/auto-pro\n    reasoningEffort: high\n" +
		"- id: agent-default-model\n  name: another-plugin\n  config:\n    provider: magpie\n    model: group/auto-pro\n    reasoningEffort: max\n" +
		"- id: agent-default-model\n  insert: []\n  config:\n    provider: magpie\n    model: group/auto-pro\n    reasoningEffort: off\n"
	_, items, err := dshParse("effort-route", raw)
	if err != nil {
		t.Fatal(err)
	}
	if got := dshEffortOf(items); got != "high" {
		t.Fatalf("effective effort = %q, want high", got)
	}
	items = dshStartOnItsRoute(items, []catalog.Model{{ID: "auto/pro"}})
	if got := dshConfig(items[1]); got["model"] != "auto/pro" {
		t.Fatalf("effective route migration = %v", got)
	}
	for _, i := range []int{0, 2, 3} {
		if got := dshConfig(items[i]); got["model"] != "group/auto-pro" {
			t.Fatalf("decoy %d changed: %v", i, got)
		}
	}
}

func TestDshHomeConfigThenMetadataKeepsConfigSource(t *testing.T) {
	_, dir, web := dshRouteHome(t)
	patch := dshHomePatch(dir)
	if err := os.WriteFile(web, []byte("[]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	own := "- id: agent-default-model\n  config:\n    provider: deepseek-official\n    model: deepseek-v4-pro\n    reasoningEffort: low\n" +
		"- id: agent-default-model\n  disabled: false\n"
	if err := os.WriteFile(patch, []byte(own), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := dshSetFile(patch, "magpie/deepseek/flash", true, magpieModels("dsh"), gateway.URL()); err != nil {
		t.Fatal(err)
	}
	_, items, err := dshRead(patch)
	if err != nil {
		t.Fatal(err)
	}
	if i := dshFindStartConfig(items); i < 0 {
		t.Fatalf("no effective home config after setting: %s", dshJoin(items))
	} else if got := dshConfig(items[i]); got["provider"] != magpieID || got["model"] != "deepseek/flash" {
		t.Fatalf("config source changed incorrectly: %v", got)
	}
	for _, it := range items {
		if it.id == "agent-default-model" && strings.Contains(strings.Join(it.lines, "\n"), "disabled: false") {
			if got := dshConfig(it); len(got) != 0 {
				t.Fatalf("metadata patch changed: %v / %v", got, it.lines)
			}
		}
	}
}

func TestDshSetEffortUsesEffectiveStartEntry(t *testing.T) {
	_, dir, web := dshRouteHome(t)
	dshKeyFixture(dir)
	if err := os.WriteFile(web, []byte(dshRouteFixture("deepseek/pro", "deepseek/flash")+
		"- id: agent-default-model # magpie\n  config:\n    provider: magpie\n    model: deepseek/pro\n    reasoningEffort: low\n"+
		"- id: agent-default-model # magpie\n  name: '@deepseek-ai/dsh-agent-default-model'\n  config:\n    provider: magpie\n    model: deepseek/pro\n    reasoningEffort: high\n"+
		"- id: agent-default-model\n  name: another-plugin\n  config:\n    provider: magpie\n    model: deepseek/pro\n    reasoningEffort: max\n"+
		"- id: agent-default-model\n  insert: []\n  config:\n    provider: magpie\n    model: deepseek/pro\n    reasoningEffort: off\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := dshSetEffort(dir, "", gateway.URL(), ""); err != nil {
		t.Fatal(err)
	}
	_, items, err := dshRead(web)
	if err != nil {
		t.Fatal(err)
	}
	var starts []dshItem
	for _, it := range items {
		if it.id == "agent-default-model" {
			starts = append(starts, it)
		}
	}
	if len(starts) != 4 {
		t.Fatalf("start fixture entries: %d\n%s", len(starts), dshJoin(items))
	}
	i := dshFindStartConfig(items)
	if i < 0 {
		t.Fatalf("no effective effort entry: %s", dshJoin(items))
	}
	if got := dshConfig(items[i]); got["reasoningEffort"] != "" {
		t.Fatalf("effective effort entry: %v", got)
	}
	for i, want := range map[int]string{0: "low", 2: "max", 3: "off"} {
		if got := dshConfig(starts[i]); got["reasoningEffort"] != want {
			t.Fatalf("decoy %d unexpectedly changed: %v", i, got)
		}
	}
}
