package library

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRegroupGitKeepsRepoCase(t *testing.T) {
	sandbox(t)
	upper, _ := gitRepo(t, map[string]string{"skills/upper/SKILL.md": "---\nname: upper\ndescription: Upper\n---\nCurrent\n"})
	lower, _ := gitRepo(t, map[string]string{"skills/lower/SKILL.md": "---\nname: lower\ndescription: Lower\n---\nCurrent\n"})
	const upperRemote = "https://git.example.com/team/Kit.git"
	const lowerRemote = "https://git.example.com/team/kit.git"
	write(t, os.Getenv("GIT_CONFIG_GLOBAL"), "[url \""+upper+"\"]\n\tinsteadOf = "+upperRemote+"\n[url \""+lower+"\"]\n\tinsteadOf = "+lowerRemote+"\n")
	for _, name := range []string{"upper", "lower"} {
		write(t, filepath.Join(skillDir(name), "SKILL.md"), "---\nname: "+name+"\ndescription: Old\n---\nOlder copy\n")
	}
	ok(t)(change(func(l *Library) error {
		l.Skills = append(l.Skills,
			&Skill{Name: "upper-source", Source: &Source{Kind: "git", Repo: upperRemote}},
			&Skill{Name: "lower-source", Source: &Source{Kind: "git", Repo: lowerRemote}},
			&Skill{Name: "upper"}, &Skill{Name: "lower"})
		return nil
	}))
	if n, err := RegroupSkills(); err != nil || n != 2 {
		t.Fatalf("regroup case-sensitive repositories: %d %v, want two skills grouped", n, err)
	}
	l := mustLoad(t)
	for name, want := range map[string]string{"upper": upperRemote, "lower": lowerRemote} {
		if got := l.skill(name).From; got != want {
			t.Errorf("%s group: %q, want %q", name, got, want)
		}
	}
}

func TestRegroupGitKeepsRefCase(t *testing.T) {
	sandbox(t)
	addr, commit := gitRepo(t, map[string]string{
		"skills/upper/SKILL.md": "---\nname: upper\ndescription: Upper\n---\nCurrent upper\n",
	})
	bare := strings.TrimPrefix(addr, "file://")
	if runtime.GOOS == "windows" && len(bare) > 2 && bare[0] == '/' && bare[2] == ':' {
		bare = bare[1:]
	}
	run := func(args ...string) {
		t.Helper()
		if out, err := exec.Command(realGit, append([]string{"--git-dir", filepath.FromSlash(bare)}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	// Different namespaces let Windows serve these case-sensitive refs too
	run("tag", "Release", "refs/heads/main")
	commit(map[string]string{
		"skills/upper":          gitRemoved,
		"skills/lower/SKILL.md": "---\nname: lower\ndescription: Lower\n---\nCurrent lower\n",
	})
	run("branch", "release", "refs/heads/main")
	for _, name := range []string{"upper", "lower"} {
		write(t, filepath.Join(skillDir(name), "SKILL.md"), "---\nname: "+name+"\ndescription: Old\n---\nOlder copy\n")
	}
	ok(t)(change(func(l *Library) error {
		l.Skills = append(l.Skills,
			&Skill{Name: "upper-source", Source: &Source{Kind: "git", Repo: addr, Ref: "Release"}},
			&Skill{Name: "lower-source", Source: &Source{Kind: "git", Repo: addr, Ref: "release"}},
			&Skill{Name: "upper"}, &Skill{Name: "lower"})
		return nil
	}))
	if n, err := RegroupSkills(); err != nil || n != 2 {
		t.Fatalf("regroup both case-sensitive refs: %d %v, want two skills grouped", n, err)
	}
	for _, name := range []string{"upper", "lower"} {
		if s := mustLoad(t).skill(name); s.From != addr {
			t.Errorf("%s's ref was not fetched: %+v", name, s)
		}
	}
}

func TestRegroupGitRejectsStaleFrom(t *testing.T) {
	for _, from := range []string{"canonical", "url"} {
		for _, linkedCheckout := range []bool{false, true} {
			t.Run(from+"/checkout="+map[bool]string{false: "no", true: "yes"}[linkedCheckout], func(t *testing.T) {
				h := sandbox(t)
				oldAddr, _ := gitRepo(t, map[string]string{
					"skills/known/SKILL.md":    "---\nname: known\ndescription: Old source\n---\nOld\n",
					"skills/stranger/SKILL.md": "---\nname: stranger\ndescription: Old source\n---\nOld\n",
				})
				const oldRemote = "https://git.example.com/old/skills.git"
				const newRemote = "https://git.example.com/current/skills.git"
				config := "[url \"" + oldAddr + "\"]\n\tinsteadOf = " + oldRemote + "\n"
				dir := filepath.Join(h, ".agents", "skills", "known")
				if linkedCheckout {
					newAddr, _ := gitRepo(t, map[string]string{
						"skills/known/SKILL.md":   "---\nname: known\ndescription: Current checkout\n---\nNew\n",
						"skills/current/SKILL.md": "---\nname: current\ndescription: Current\n---\nNew\n",
					})
					config += "[url \"" + newAddr + "\"]\n\tinsteadOf = " + newRemote + "\n"
					clone := filepath.Join(h, ".agents", "skills", "checkout")
					if out, err := exec.Command(realGit, "clone", "-q", newAddr, clone).CombinedOutput(); err != nil {
						t.Fatal(err, string(out))
					}
					if out, err := exec.Command(realGit, "-C", clone, "remote", "set-url", "origin", newRemote).CombinedOutput(); err != nil {
						t.Fatal(err, string(out))
					}
					dir = filepath.Join(clone, "skills", "known")
				} else {
					skill(t, dir, "known", "New unrelated skill")
				}
				write(t, os.Getenv("GIT_CONFIG_GLOBAL"), config)
				installed := time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339Nano)
				write(t, filepath.Join(h, ".agents", ".skill-lock.json"), `{"skills":{"known":{"sourceType":"git","sourceUrl":"`+oldRemote+`","installedAt":"`+installed+`"}}}`)
				if err := os.MkdirAll(filepath.Dir(skillDir("known")), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := dirLink(dir, skillDir("known")); err != nil {
					t.Fatal(err)
				}
				for _, name := range []string{"stranger", "current"} {
					write(t, filepath.Join(skillDir(name), "SKILL.md"), "---\nname: "+name+"\ndescription: Mine\n---\nUntraced old copy\n")
				}
				oldFrom := "git.example.com/old/skills"
				if from == "url" {
					oldFrom = oldRemote
				}
				ok(t)(change(func(l *Library) error {
					l.Skills = append(l.Skills, &Skill{Name: "known", Source: &Source{Kind: "folder", Dir: dir}, From: oldFrom}, &Skill{Name: "stranger"}, &Skill{Name: "current"})
					return nil
				}))
				wantCount, wantRepo := 0, ""
				if linkedCheckout {
					wantCount, wantRepo = 2, "git.example.com/current/skills"
				}
				if got := repoIn(t)["known"]; got != wantRepo {
					t.Errorf("known traced as %q, want %q", got, wantRepo)
				}
				if n, err := RegroupSkills(); err != nil || n != wantCount {
					t.Errorf("regroup: %d %v, want %d", n, err, wantCount)
				}
				l := mustLoad(t)
				if s := l.skill("stranger"); s.From != "" || s.Source != nil {
					t.Errorf("stale source claimed another untraced skill: %+v", s)
				}
				wantFrom := ""
				if linkedCheckout {
					wantFrom = newRemote
				}
				if got := l.skill("current").From; got != wantFrom {
					t.Errorf("current checkout group: %q, want %q", got, wantFrom)
				}
			})
		}
	}
}

func TestGitProbeNotesAndBindsTheSameCopy(t *testing.T) {
	sandbox(t)
	const body = "---\nname: review\ndescription: Review code\n---\nBody\n"
	addr, _ := gitRepo(t, map[string]string{"skills/review/SKILL.md": body})
	write(t, filepath.Join(skillDir("review"), "SKILL.md"), body)
	ok(t)(change(func(l *Library) error {
		l.Skills = append(l.Skills, &Skill{Name: "review", Agents: []string{"claude"}})
		return nil
	}))
	if _, err := ProbeSkills(addr); err != nil {
		t.Fatal(err)
	}
	s := mustLoad(t).skill("review")
	if s.Source == nil || s.Source.Kind != "git" || s.Source.Repo != addr || s.Source.Path != "skills/review" || len(s.Commit) != 40 {
		t.Fatalf("same copy wasn't bound to the probed git repository: %+v", s)
	}
}

func TestRegroupGitTracedFromLockOrCheckout(t *testing.T) {
	for _, from := range []string{"lock", "checkout", "from"} {
		t.Run(from, func(t *testing.T) {
			h := sandbox(t)
			addr, _ := gitRepo(t, map[string]string{
				"skills/known/SKILL.md": "---\nname: known\ndescription: Known\n---\nCurrent\n",
				"skills/older/SKILL.md": "---\nname: older\ndescription: Updated\n---\nCurrent\n",
			})
			const remote = "https://git.example.com/team/skills.git"
			write(t, os.Getenv("GIT_CONFIG_GLOBAL"), "[url \""+addr+"\"]\n\tinsteadOf = "+remote+"\n")
			dir := filepath.Join(h, "checkout", "known")
			if from == "lock" {
				dir = filepath.Join(h, ".agents", "skills", "known")
				write(t, filepath.Join(h, ".agents", ".skill-lock.json"), `{"skills":{"known":{"sourceType":"git","sourceUrl":"`+remote+`"}}}`)
			} else if from == "checkout" {
				checkout(t, filepath.Dir(dir), "[remote \"origin\"]\n\turl = "+remote+"\n")
				write(t, filepath.Join(h, ".agents", ".skill-lock.json"), `{"skills":{"known":{"sourceType":"git","sourceUrl":"https://git.example.com/unrelated/skills.git"}}}`)
			}
			skill(t, dir, "known", "Known")
			if err := os.MkdirAll(filepath.Dir(skillDir("known")), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := dirLink(dir, skillDir("known")); err != nil {
				t.Fatal(err)
			}
			write(t, filepath.Join(skillDir("older"), "SKILL.md"), "---\nname: older\ndescription: Old\n---\nOlder\n")
			ok(t)(change(func(l *Library) error {
				known := &Skill{Name: "known", Source: &Source{Kind: "folder", Dir: dir}}
				if from == "from" {
					known.From = remote
				}
				l.Skills = append(l.Skills, known, &Skill{Name: "older"})
				return nil
			}))
			if n, err := RegroupSkills(); err != nil || n != 1 {
				t.Fatalf("regroup from %s: %d %v", from, n, err)
			}
			if s := mustLoad(t).skill("older"); s.From != remote {
				t.Errorf("older copy grouped as %q", s.From)
			}
		})
	}
}

func TestRegroupGitNotesOlderCopiesAndKeepsOtherSources(t *testing.T) {
	sandbox(t)
	const body = "---\nname: known\ndescription: Known\n---\nCurrent\n"
	addr, _ := gitRepo(t, map[string]string{
		"skills/known/SKILL.md": body,
		"skills/older/SKILL.md": "---\nname: older\ndescription: Updated\n---\nCurrent\n",
		"skills/other/SKILL.md": "---\nname: other\ndescription: Other\n---\nCurrent\n",
	})
	ok(t)(InstallSkills(addr, []string{"skills/known"}, []string{"claude"}))
	write(t, filepath.Join(skillDir("older"), "SKILL.md"), "---\nname: older\ndescription: Old\n---\nOld copy\n")
	write(t, filepath.Join(skillDir("other"), "SKILL.md"), "---\nname: other\ndescription: Other\n---\nCurrent\n")
	ok(t)(change(func(l *Library) error {
		l.Skills = append(l.Skills, &Skill{Name: "older"}, &Skill{Name: "other", From: "git.example.com/other/skills"})
		return nil
	}))
	if n, err := RegroupSkills(); err != nil || n != 1 {
		t.Fatalf("git regroup: %d, %v, want one older copy grouped", n, err)
	}
	l := mustLoad(t)
	if s := l.skill("older"); s.From != addr || s.Source != nil {
		t.Errorf("older copy: %+v", s)
	}
	if s := l.skill("other"); s.From != "git.example.com/other/skills" || s.Source != nil {
		t.Errorf("other repository's skill moved: %+v", s)
	}
	if n, err := RegroupSkills(); err != nil || n != 0 {
		t.Errorf("second regroup: %d, %v", n, err)
	}
}
