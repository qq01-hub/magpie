package library

import (
	"os"
	"path/filepath"
	"testing"
)

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
	for _, from := range []string{"lock", "checkout"} {
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
			} else {
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
				l.Skills = append(l.Skills, &Skill{Name: "known", Source: &Source{Kind: "folder", Dir: dir}}, &Skill{Name: "older"})
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
