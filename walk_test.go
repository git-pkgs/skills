package skills

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"slices"
	"testing"
	"testing/fstest"

	"github.com/git-pkgs/archives"
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
	err := Walk(fsys, ".", func(*Skill, error) error {
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
	err := Walk(fsys, "x/skills", func(s *Skill, err error) error {
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

func TestWalkMissingRoot(t *testing.T) {
	if err := Walk(fstest.MapFS{}, "nope", func(*Skill, error) error { return nil }); err == nil {
		t.Fatal("expected error for missing root")
	}
}

func npmTarball(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, body := range files {
		if err := tw.WriteHeader(&tar.Header{Name: "package/" + name, Mode: 0o644, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestDiscoverArchive(t *testing.T) {
	data := npmTarball(t, map[string]string{
		"package.json":  `{"name":"@reduxjs/toolkit","keywords":["tanstack-intent"]}`,
		"dist/index.js": "module.exports = {}",
		"skills/build-modern-redux-apps/modern-redux/SKILL.md": "---\nname: build-modern-redux-apps/modern-redux\n" +
			"description: Set up Redux\nlibrary: '@reduxjs/toolkit'\n---\nUse configureStore.\n",
		"skills/build-modern-redux-apps/modern-redux/scripts/check.sh": "#!/bin/sh\n",
	})
	reader, err := archives.OpenBytesWithPrefix("toolkit-2.11.2.tgz", data, "package/")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = reader.Close() }()

	found, err := DiscoverArchive(reader)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 1 {
		t.Fatalf("found %d skills", len(found))
	}
	skill := found[0]
	if skill.Dir != "skills/build-modern-redux-apps/modern-redux" {
		t.Errorf("dir = %q", skill.Dir)
	}
	if skill.Extra["library"] != "@reduxjs/toolkit" || skill.Body != "Use configureStore." {
		t.Errorf("skill = %+v", skill)
	}
	if !skill.HasScripts() || !slices.Equal(skill.Files, []string{"scripts/check.sh"}) {
		t.Errorf("files = %v", skill.Files)
	}
}
