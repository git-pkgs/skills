package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), filePerm); err != nil {
		t.Fatal(err)
	}
}

func TestStage(t *testing.T) {
	source := t.TempDir()
	writeFile(t, filepath.Join(source, "audit", Filename), "---\nname: audit\ndescription: Audit\n---\nOld body\n")
	writeFile(t, filepath.Join(source, "audit", "scripts", "check.sh"), "#!/bin/sh\n")
	if err := os.Chmod(filepath.Join(source, "audit", "scripts", "check.sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	fsys := os.DirFS(source)
	skill, err := ParseFS(fsys, "audit")
	if err != nil {
		t.Fatal(err)
	}
	skill.Body = "New body"
	skill.SchemaJSON = `{"type":"object"}`

	workspace := t.TempDir()
	destination := filepath.Join(workspace, ".claude", "skills", "audit")
	if err := Stage(fsys, skill, workspace, destination); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(destination, Filename))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "New body") || strings.Contains(string(content), "Old body") {
		t.Errorf("SKILL.md = %q", content)
	}
	info, err := os.Stat(filepath.Join(destination, "scripts", "check.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("script mode = %v", info.Mode())
	}
	schema, err := os.ReadFile(filepath.Join(destination, schemaFilename))
	if err != nil {
		t.Fatal(err)
	}
	if string(schema) != skill.SchemaJSON {
		t.Errorf("schema.json = %q", schema)
	}

	writeFile(t, filepath.Join(destination, "stale.txt"), "stale")
	if err := Stage(fsys, skill, workspace, destination); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(destination, "stale.txt")); !os.IsNotExist(err) {
		t.Errorf("stale file survived, error = %v", err)
	}
}

func TestStageWithoutSource(t *testing.T) {
	workspace := t.TempDir()
	destination := filepath.Join(workspace, "skills", "x")
	if err := Stage(nil, &Skill{Name: "x", Body: "body"}, workspace, destination); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(destination, Filename))
	if err != nil || string(content) != "body\n" {
		t.Errorf("content = %q err = %v", content, err)
	}
}

func TestStageRejectsDestinationOutsideWorkspace(t *testing.T) {
	workspace := t.TempDir()
	for _, destination := range []string{workspace, filepath.Join(workspace, "..", "elsewhere")} {
		if err := Stage(nil, &Skill{Name: "x"}, workspace, destination); err == nil {
			t.Errorf("Stage accepted %s", destination)
		}
	}
}

func TestStageRejectsSymlinkedDestination(t *testing.T) {
	workspace := t.TempDir()
	victim := t.TempDir()
	writeFile(t, filepath.Join(victim, "keep.txt"), "keep")
	if err := os.MkdirAll(filepath.Join(workspace, ".claude"), dirPerm); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(victim, filepath.Join(workspace, ".claude", "skills")); err != nil {
		t.Skip("symlinks unsupported:", err)
	}
	err := Stage(nil, &Skill{Name: "x"}, workspace, filepath.Join(workspace, ".claude", "skills", "x"))
	if err == nil {
		t.Fatal("Stage followed a symlink out of the workspace")
	}
	if _, err := os.Stat(filepath.Join(victim, "keep.txt")); err != nil {
		t.Errorf("victim modified: %v", err)
	}
}

func TestStageRejectsSourceSymlink(t *testing.T) {
	source := t.TempDir()
	writeFile(t, filepath.Join(source, "x", Filename), "body")
	if err := os.Symlink("/etc", filepath.Join(source, "x", "escape")); err != nil {
		t.Skip("symlinks unsupported:", err)
	}
	fsys := os.DirFS(source)
	skill, err := ParseFS(fsys, "x")
	if err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	err = Stage(fsys, skill, workspace, filepath.Join(workspace, "skills", "x"))
	if err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("err = %v", err)
	}
}

func TestStageSkipsNestedSkillAndGit(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "group", Filename), "---\nname: group\ndescription: d\n---\nbody\n")
	writeFile(t, filepath.Join(src, "group", "notes.md"), "notes")
	writeFile(t, filepath.Join(src, "group", "two", Filename), "---\nname: two\ndescription: d\n---\nbody\n")
	writeFile(t, filepath.Join(src, "group", "two", "scripts", "run.sh"), "#!/bin/sh\n")
	writeFile(t, filepath.Join(src, "group", "sub", ".git", "config"), "x")
	skill, err := Parse(filepath.Join(src, "group"))
	if err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	destination := filepath.Join(workspace, "skills", "group")
	if err := Stage(nil, skill, workspace, destination); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"two", "sub/.git"} {
		if _, err := os.Stat(filepath.Join(destination, filepath.FromSlash(p))); !os.IsNotExist(err) {
			t.Errorf("%s was staged", p)
		}
	}
	if _, err := os.Stat(filepath.Join(destination, "notes.md")); err != nil {
		t.Errorf("notes.md not staged: %v", err)
	}
}

func TestStageExtraKeys(t *testing.T) {
	skill, err := ParseBytes([]byte("---\nname: x\ndescription: d\nmodel: opus\n---\nbody\n"), "x/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	destination := filepath.Join(workspace, "skills", "x")

	if err := Stage(nil, skill, workspace, destination); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(destination, Filename))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "model: opus") {
		t.Errorf("default Stage dropped extra keys: %q", content)
	}

	if err := Stage(nil, skill, workspace, destination, WithoutExtra()); err != nil {
		t.Fatal(err)
	}
	content, err = os.ReadFile(filepath.Join(destination, Filename))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), "model") || !strings.Contains(string(content), "description: d") {
		t.Errorf("WithoutExtra SKILL.md = %q", content)
	}
	if skill.Extra["model"] != "opus" {
		t.Error("WithoutExtra modified the caller's skill")
	}
}
