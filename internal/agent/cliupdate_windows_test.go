package agent

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// junction makes path a junction to the folder target, as install.ps1's
// New-Item -ItemType Junction does, in place of one already there.
func junction(t *testing.T, target, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(path); err == nil {
		if err := os.Remove(path); err != nil { // the junction, not what it leads to
			t.Fatal(err)
		}
	}
	if out, err := exec.Command("cmd", "/c", "mklink", "/J", path, target).CombinedOutput(); err != nil {
		t.Fatalf("mklink /J: %v %s", err, out)
	}
}

// Codex's install.ps1 puts %LOCALAPPDATA%\Programs\OpenAI\Codex\bin on PATH,
// a junction to ~\.codex\packages\standalone\current\bin, and current is a
// junction to releases\<version>-<target>, retargeted by each update
// (xiaoxiaofeixz on Discord: the Agents page kept the old Codex version). It
// is Codex's own install, updated by codex update, and each update is a new
// binary to ask, though npm's archive gives every version's codex.exe the
// same time.
func TestCodexInstallPs1Junctions(t *testing.T) {
	home := t.TempDir()
	releases := filepath.Join(home, `.codex\packages\standalone\releases`)
	old := filepath.Join(releases, "0.161.0-x86_64-pc-windows-msvc")
	cur := filepath.Join(releases, "0.162.0-x86_64-pc-windows-msvc")
	when := time.Date(1985, 10, 26, 8, 15, 0, 0, time.UTC)
	for _, r := range []string{old, cur} {
		exe := file(t, filepath.Join(r, `bin\codex.exe`), "MZ")
		if err := os.Chtimes(exe, when, when); err != nil {
			t.Fatal(err)
		}
	}
	current := filepath.Join(home, `.codex\packages\standalone\current`)
	junction(t, old, current)
	visible := filepath.Join(home, `AppData\Local\Programs\OpenAI\Codex\bin`)
	junction(t, filepath.Join(current, "bin"), visible)
	bin := filepath.Join(visible, "codex.exe")

	if u := howInstalled(cliSpecs["codex"], bin); u == nil || u.via != "self" || u.pkg != "@openai/codex" || !reflect.DeepEqual(u.cmd, []string{bin, "update"}) {
		t.Errorf("install.ps1's codex: %+v", u)
	}

	says := "codex-cli 0.161.0"
	wasRun := runVersion
	runVersion = func(string) string { return says }
	t.Cleanup(func() { runVersion = wasRun })
	if v := installedVersion(bin); v != "0.161.0" {
		t.Fatalf("installed: %q", v)
	}
	// codex update: the new release, current retargeted to it
	junction(t, cur, current)
	says = "codex-cli 0.162.0"
	if v := installedVersion(bin); v != "0.162.0" {
		t.Errorf("after the update: %q, want 0.162.0", v)
	}
}

// Volta's Windows shims are recognized only when its default package
// metadata names the package and tool, and `volta which` returns the matching
// entry under the same Volta installation. The fake manager exercises the
// actual CLI and UpdateCLI command path without changing a real CLI.
func TestVoltaDefaultPackagesUpdateThroughVolta(t *testing.T) {
	root := t.TempDir()
	volta := filepath.Join(root, "volta.cmd")
	list := filepath.Join(root, "list.txt")
	calls := filepath.Join(root, "volta-calls.txt")
	file(t, volta, "@echo off\r\necho %*>>\"%VOLTA_CALLS%\"\r\nif /I \"%1\"==\"which\" type \"%VOLTA_TARGET_FILE%\"\r\nif /I \"%1\"==\"list\" type \"%VOLTA_LIST_FILE%\"\r\n")
	t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("VOLTA_LIST_FILE", list)
	t.Setenv("VOLTA_CALLS", calls)
	data := filepath.Join(root, "Volta")
	tools := filepath.Join(data, "tools", "image", "node", "22.23.2")
	packageTools := filepath.Join(data, "tools", "image", "packages")
	bins := map[string]string{
		"codex":  filepath.Join(data, "bin", "codex.cmd"),
		"claude": filepath.Join(data, "bin", "claude.cmd"),
		"omp":    filepath.Join(tools, "omp.ps1"),
	}
	packages := map[string]string{
		"codex":  "@openai/codex",
		"claude": "@anthropic-ai/claude-code",
		"omp":    "@oh-my-pi/pi-coding-agent",
	}
	for _, bin := range bins {
		file(t, bin, "@echo off\r\n")
	}
	for name, pkg := range packages {
		entry := filepath.Join(packageTools, filepath.FromSlash(pkg), name+".cmd")
		if name == "omp" {
			entry = bins[name]
		} else {
			file(t, entry, "@echo off\r\n")
		}
		t.Setenv("VOLTA_TARGET_FILE", filepath.Join(root, name+"-target.txt"))
		file(t, os.Getenv("VOLTA_TARGET_FILE"), entry+"\r\n")
		version := map[string]string{"codex": "0.171.0", "claude": "2.1.330", "omp": "19.1.1"}[name]
		line := "package " + pkg + "@" + version + " / " + name + " / node@22.23.2 npm@built-in (default)\r\n"
		if err := os.WriteFile(list, []byte(line), 0o600); err != nil {
			t.Fatal(err)
		}
		u := howInstalled(cliSpecs[name], bins[name])
		if u == nil || u.via != "volta" || !reflect.DeepEqual(u.cmd, []string{volta, "install", pkg + "@latest"}) {
			b, _ := os.ReadFile(calls)
			t.Fatalf("%s: %+v; target=%q list=%q calls=%q", name, u, os.Getenv("VOLTA_TARGET_FILE"), line, b)
		}
	}

	// The public CLI row and update call use Volta's command for each agent,
	// routed through the existing executor seam rather than running it.
	oldRun, oldVersion, oldGOOS, oldLetGo := runUpdate, runVersion, updateGOOS, letGoBeforeUpdate
	t.Cleanup(func() {
		runUpdate, runVersion, updateGOOS, letGoBeforeUpdate = oldRun, oldVersion, oldGOOS, oldLetGo
		versions, latests = memo{}, memo{}
	})
	updateGOOS = "windows"
	letGoBeforeUpdate = func(context.Context, string) {}
	current := map[string]string{}
	runVersion = func(bin string) string { return current[filepath.Clean(bin)] }
	latests = memo{}
	versions = memo{}
	var ran []string
	runUpdate = func(_ context.Context, u *updater) ([]byte, error) {
		ran = append(ran, u.shown())
		for name, pkg := range packages {
			if u.pkg == pkg {
				current[filepath.Clean(bins[name])] = map[string]string{"codex": "0.172.0", "claude": "2.1.331", "omp": "19.1.2"}[name]
			}
		}
		if u.pkg == packages["omp"] {
			current[filepath.Clean(bins["omp"])] = "19.1.2"
		}
		return nil, nil
	}
	for name, pkg := range packages {
		entry := filepath.Join(packageTools, filepath.FromSlash(pkg), name+".cmd")
		if name == "omp" {
			entry = bins[name]
		}
		current[filepath.Clean(bins[name])] = map[string]string{"codex": "0.171.0", "claude": "2.1.330", "omp": "19.1.1"}[name]
		latest := map[string]string{"codex": "0.172.0", "claude": "2.1.331", "omp": "19.1.2"}[name]
		latests.get("npm:"+pkg, "", 1<<40, 1<<40, func() (string, error) { return latest, nil })
		t.Setenv("VOLTA_TARGET_FILE", filepath.Join(root, name+"-target.txt"))
		if err := os.WriteFile(os.Getenv("VOLTA_TARGET_FILE"), []byte(entry+"\r\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		line := "package " + pkg + "@" + current[filepath.Clean(bins[name])] + " / " + name + " / node@22.23.2 npm@built-in (default)\r\n"
		if err := os.WriteFile(list, []byte(line), 0o600); err != nil {
			t.Fatal(err)
		}
		a := &Agent{ID: name, Name: name, Bin: bins[name]}
		before, ok := a.CLI()
		if !ok || before.Via != "volta" || before.Command != "volta.cmd install "+pkg+"@latest" || !before.Update {
			t.Fatalf("%s before update: %+v %v", name, before, ok)
		}
		after, err := a.UpdateCLI()
		want := "volta.cmd install " + pkg + "@latest"
		if err != nil || after.Version != latest || after.Update || !reflect.DeepEqual(ran, []string{want}) {
			t.Fatalf("%s after update: %+v, %v, ran %q", name, after, err, ran)
		}
		ran = nil
	}

	// A plain npm install remains npm even while Volta is on PATH.
	managerCalls, _ := os.ReadFile(calls)
	prefix := filepath.Join(root, "npm")
	entry := filepath.Join(prefix, "node_modules", "@openai", "codex", "bin", "codex.js")
	file(t, entry, "")
	pathBin := filepath.Join(prefix, "codex.cmd")
	file(t, pathBin, "node_modules/@openai/codex/bin/codex.js\r\n")
	t.Setenv("VOLTA_TARGET_FILE", filepath.Join(root, "ordinary-target.txt"))
	file(t, os.Getenv("VOLTA_TARGET_FILE"), filepath.Join(root, "outside", "codex.cmd")+"\r\n")
	if u := howInstalled(cliSpecs["codex"], pathBin); u == nil || u.via != "npm" {
		t.Fatalf("ordinary npm: %+v", u)
	}
	if after, _ := os.ReadFile(calls); string(after) != string(managerCalls) {
		t.Fatalf("ordinary npm asked Volta: %q", after)
	}

	// A Volta-shaped node_modules entry with fetched metadata is unknown, not
	// an npm install merely because the package path is under Volta.
	voltaNodeBin := filepath.Join(tools, "node_modules", "@openai", "codex", "bin", "codex.cmd")
	file(t, voltaNodeBin, "")
	t.Setenv("VOLTA_TARGET_FILE", filepath.Join(root, "fetched-target.txt"))
	file(t, os.Getenv("VOLTA_TARGET_FILE"), voltaNodeBin+"\r\n")
	for _, row := range []string{
		"package @openai/codex@0.171.0 (fetched)\r\n",
		"package @openai/codex@project / codex / node@project npm@project (current @ C:\\work)\r\n",
		"not a Volta list row\r\n",
	} {
		if err := os.WriteFile(list, []byte(row), 0o600); err != nil {
			t.Fatal(err)
		}
		if u := howInstalled(cliSpecs["codex"], voltaNodeBin); u != nil {
			t.Fatalf("non-default Volta package %q: %+v", row, u)
		}
	}
}
