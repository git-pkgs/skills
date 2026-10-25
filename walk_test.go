package skills

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

func skillDoc(name string) *fstest.MapFile {
	return &fstest.MapFile{Data: []byte("---\nname: " + name + "\ndescription: d\n---\nbody\n")}
}

func dirs(found []*Skill) []string {
	out := make([]string, 0, len(found))
	for _, s := range found {
		out = append(out, s.Dir)
	}
	return out
}

func TestDiscover(t *testing.T) {
	fsys := fstest.MapFS{
		"skills/one/SKILL.md":                    skillDoc("one"),
		"skills/group/two/SKILL.md":              skillDoc("two"),
		"mypkg/skills/three/SKILL.md":            skillDoc("three"),
		"node_modules/dep/skills/dep/SKILL.md":   skillDoc("dep"),
		"vendor/x/SKILL.md":                      skillDoc("x"),
		".git/SKILL.md":                          skillDoc("git"),
		"a/b/c/d/e/f/g/deep/SKILL.md":            skillDoc("deep"),
		"README.md":                              {Data: []byte("readme")},
		"skills/one/references/not-a-skill.md":   {Data: []byte("ref")},
		"skills/one/references/nested/SKILL.txt": {Data: []byte("x")},
	}
	found, err := Discover(fsys)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"mypkg/skills/three", "skills/group/two", "skills/one"}
	if got := dirs(found); !slices.Equal(got, want) {
		t.Errorf("dirs = %v, want %v", got, want)
	}
}

func TestDiscoverOptions(t *testing.T) {
	fsys := fstest.MapFS{
		"a/b/SKILL.md":                  skillDoc("b"),
		"node_modules/dep/one/SKILL.md": skillDoc("one"),
	}
	found, err := Discover(fsys, WithMaxDepth(1), WithSkipDirs())
	if err != nil {
		t.Fatal(err)
	}
	if got := dirs(found); !slices.Equal(got, []string{}) {
		t.Errorf("depth 1 found %v", got)
	}
	found, err = Discover(fsys, WithSkipDirs())
	if err != nil {
		t.Fatal(err)
	}
	if got := dirs(found); !slices.Equal(got, []string{"a/b", "node_modules/dep/one"}) {
		t.Errorf("no skip found %v", got)
	}
}

func TestDiscoverKeepsGoodSkillsWhenOneFails(t *testing.T) {
	fsys := fstest.MapFS{
		"bad/SKILL.md":  {Data: []byte("---\nname: [unclosed\n---\n")},
		"good/SKILL.md": skillDoc("good"),
	}
	found, err := Discover(fsys)
	if err == nil {
		t.Fatal("expected joined parse error")
	}
	if got := dirs(found); !slices.Equal(got, []string{"good"}) {
		t.Errorf("found %v", got)
	}
}

func TestWalkStopsOnCallbackError(t *testing.T) {
	fsys := fstest.MapFS{"a/SKILL.md": skillDoc("a"), "b/SKILL.md": skillDoc("b")}
	stop := errors.New("stop")
	calls := 0
	err := WalkFS(fsys, ".", func(*Skill, error) error {
		calls++
		return stop
	})
	if !errors.Is(err, stop) || calls != 1 {
		t.Errorf("err = %v calls = %d", err, calls)
	}
}

func TestWalkSubtree(t *testing.T) {
	fsys := fstest.MapFS{"x/skills/a/SKILL.md": skillDoc("a"), "y/SKILL.md": skillDoc("y")}
	var got []string
	err := WalkFS(fsys, "x/skills", func(s *Skill, err error) error {
		if err != nil {
			return err
		}
		got = append(got, s.Dir)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, []string{"x/skills/a"}) {
		t.Errorf("got %v", got)
	}
}

func TestDiscoverKeepsMissingDescription(t *testing.T) {
	fsys := fstest.MapFS{
		"good/SKILL.md":   skillDoc("good"),
		"nodesc/SKILL.md": {Data: []byte("---\nname: nodesc\n---\nbody\n")},
	}
	found, err := Discover(fsys)
	if err != nil {
		t.Fatal(err)
	}
	if got := dirs(found); !slices.Equal(got, []string{"good", "nodesc"}) {
		t.Fatalf("found %v", got)
	}
	if !hasWarning(found[1], "description is missing") {
		t.Errorf("warnings = %v", found[1].Warnings)
	}
}

func TestWalkFSRootError(t *testing.T) {
	err := WalkFS(fstest.MapFS{}, "nope", func(*Skill, error) error { return nil })
	if err == nil || strings.Count(err.Error(), "skills: walk") != 1 {
		t.Errorf("err = %v", err)
	}
}

func TestDefaultSkipDirsIsCopy(t *testing.T) {
	a := DefaultSkipDirs()
	a[0] = "mutated"
	if b := DefaultSkipDirs(); b[0] == "mutated" || !slices.Contains(b, gitMetadataDir) {
		t.Errorf("DefaultSkipDirs() shares backing storage: %v", b)
	}
}
