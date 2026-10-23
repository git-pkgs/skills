package skills

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseDisk(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "audit", Filename)
	writeFile(t, path, "---\nname: audit\ndescription: Check a repository\nmetadata:\n  example.level: strict\n---\n\nRead the repository.\n")
	writeFile(t, filepath.Join(root, "audit", "scripts", "check.sh"), "#!/bin/sh\n")

	for _, arg := range []string{path, filepath.Join(root, "audit")} {
		skill, err := Parse(arg)
		if err != nil {
			t.Fatal(err)
		}
		if skill.Name != "audit" || skill.Description != "Check a repository" || skill.Body != "Read the repository." {
			t.Errorf("skill = %+v", skill)
		}
		if skill.Metadata["example.level"] != "strict" {
			t.Errorf("metadata = %#v", skill.Metadata)
		}
		if skill.SourcePath != filepath.Join(root, "audit") {
			t.Errorf("SourcePath = %q", skill.SourcePath)
		}
		if !skill.HasScripts() || len(skill.Warnings) != 0 {
			t.Errorf("files = %v warnings = %v", skill.Files, skill.Warnings)
		}
	}
}

func TestParseDiskRelativePath(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "lint", Filename), "---\nname: lint\ndescription: d\n---\n")
	t.Chdir(root)
	skill, err := Parse("lint")
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(skill.SourcePath) || filepath.Base(skill.SourcePath) != "lint" {
		t.Errorf("SourcePath = %q", skill.SourcePath)
	}
}

func TestParseDiskPlainMarkdown(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "review.md")
	writeFile(t, path, "# Review\n\nCheck the code.\n")

	skill, err := Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	if skill.Name != "review" || skill.Body != "# Review\n\nCheck the code." {
		t.Errorf("skill = %+v", skill)
	}
	if skill.SourcePath != dir || len(skill.Warnings) != 0 {
		t.Errorf("SourcePath = %q warnings = %v", skill.SourcePath, skill.Warnings)
	}
}

func TestParseDiskSourceHashIncludesSchema(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "audit")
	path := filepath.Join(dir, Filename)
	writeFile(t, path, "---\nname: audit\ndescription: Audit\n---\nCheck it.\n")

	withoutSchema, err := Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, schemaFilename), `{"type":"object"}`)
	withSchema, err := Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	if withSchema.SourceHash == withoutSchema.SourceHash || withSchema.SchemaJSON != `{"type":"object"}` {
		t.Fatalf("hash unchanged or schema = %q", withSchema.SchemaJSON)
	}
	writeFile(t, filepath.Join(dir, schemaFilename), `{"type":"array"}`)
	changedSchema, err := Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	if changedSchema.SourceHash == withSchema.SourceHash {
		t.Fatal("SourceHash did not change after a schema-only edit")
	}
}

func TestParseDiskErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad", Filename)
	writeFile(t, path, "---\nname: [\n---\nbody\n")
	_, err := Parse(path)
	if err == nil || !strings.Contains(err.Error(), path) {
		t.Fatalf("err = %v, want absolute path in message", err)
	}
	if _, err := Parse(filepath.Join(dir, "missing")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("err = %v, want not exist", err)
	}
}

func TestWalkDisk(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "one", Filename), "---\nname: one\ndescription: One\n---\none\n")
	writeFile(t, filepath.Join(root, "nested", "two", Filename), "---\nname: two\ndescription: Two\n---\ntwo\n")
	writeFile(t, filepath.Join(root, ".git", "hidden", Filename), "hidden\n")

	var names, paths []string
	err := Walk(root, func(skill *Skill) error {
		names = append(names, skill.Name)
		paths = append(paths, skill.SourcePath)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(names, ",") != "two,one" {
		t.Errorf("walked %v", names)
	}
	if paths[0] != filepath.Join(root, "nested", "two") || paths[1] != filepath.Join(root, "one") {
		t.Errorf("source paths = %v", paths)
	}
}

func TestWalkDiskStopsOnParseError(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "bad", Filename), "---\nname: [\n---\n")
	if err := Walk(root, func(*Skill) error { return nil }); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestStageFromSourcePath(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "audit", Filename), "---\nname: audit\ndescription: Audit\n---\nbody\n")
	writeFile(t, filepath.Join(root, "audit", "references", "api.md"), "# API\n")
	skill, err := Parse(filepath.Join(root, "audit"))
	if err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	destination := filepath.Join(workspace, ".claude", "skills", "audit")
	if err := Stage(nil, skill, workspace, destination); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(destination, "references", "api.md")); err != nil {
		t.Fatal(err)
	}
}

func TestValidateName(t *testing.T) {
	for _, name := range []string{"audit", "security-review"} {
		if err := ValidateName(name); err != nil {
			t.Errorf("ValidateName(%q) = %v", name, err)
		}
	}
	for _, name := range []string{"", "..", "../outside", "a/b", "/abs"} {
		if err := ValidateName(name); err == nil {
			t.Errorf("ValidateName(%q) accepted", name)
		}
	}
}
