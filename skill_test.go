package skills

import (
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

func TestParseWithFrontmatter(t *testing.T) {
	fsys := fstest.MapFS{
		"audit/SKILL.md": {Data: []byte(`---
name: audit
description: Check a repository
license: MIT
allowed-tools: Read Grep
metadata:
  example.level: strict
---

Read the repository.
`)},
		"audit/schema.json":       {Data: []byte(`{"type":"object"}`)},
		"audit/scripts/run.sh":    {Data: []byte("#!/bin/sh\n"), Mode: 0o755},
		"audit/references/api.md": {Data: []byte("# API\n")},
	}

	skill, err := Parse(fsys, "audit/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	if skill.Name != "audit" || skill.Description != "Check a repository" || skill.License != "MIT" {
		t.Errorf("skill = %+v", skill)
	}
	if skill.AllowedTools != "Read Grep" {
		t.Errorf("allowed-tools = %q", skill.AllowedTools)
	}
	if skill.Metadata["example.level"] != "strict" {
		t.Errorf("metadata = %#v", skill.Metadata)
	}
	if skill.Body != "Read the repository." {
		t.Errorf("body = %q", skill.Body)
	}
	if skill.Dir != "audit" {
		t.Errorf("dir = %q", skill.Dir)
	}
	if skill.SchemaJSON != `{"type":"object"}` {
		t.Errorf("schema = %q", skill.SchemaJSON)
	}
	want := []string{"references/api.md", "schema.json", "scripts/run.sh"}
	if !slices.Equal(skill.Files, want) {
		t.Errorf("files = %v, want %v", skill.Files, want)
	}
	if !skill.HasScripts() {
		t.Error("HasScripts = false")
	}
	if len(skill.Warnings) != 0 {
		t.Errorf("warnings = %v", skill.Warnings)
	}
}

func TestParseDirectoryPath(t *testing.T) {
	fsys := fstest.MapFS{"lint/SKILL.md": {Data: []byte("---\nname: lint\ndescription: d\n---\nbody\n")}}
	skill, err := Parse(fsys, "lint")
	if err != nil {
		t.Fatal(err)
	}
	if skill.Name != "lint" || skill.Body != "body" {
		t.Errorf("skill = %+v", skill)
	}
}

func TestParsePlainMarkdown(t *testing.T) {
	fsys := fstest.MapFS{
		"prompts/review.md": {Data: []byte("# Review\n\nCheck the code.\n")},
		"prompts/other.md":  {Data: []byte("x")},
	}
	skill, err := Parse(fsys, "prompts/review.md")
	if err != nil {
		t.Fatal(err)
	}
	if skill.Name != "review" {
		t.Errorf("name = %q", skill.Name)
	}
	if skill.Body != "# Review\n\nCheck the code." {
		t.Errorf("body = %q", skill.Body)
	}
	if len(skill.Files) != 0 || len(skill.Warnings) != 0 {
		t.Errorf("files = %v warnings = %v", skill.Files, skill.Warnings)
	}
}

func TestParseKeepsNonSpecKeys(t *testing.T) {
	// Frontmatter shape published by TanStack Intent packages such as
	// @reduxjs/toolkit.
	raw := []byte(`---
name: build-modern-redux-apps/modern-redux
description: >
  Use this when setting up a new Redux Toolkit app.
type: lifecycle
library: '@reduxjs/toolkit'
library_version: '2.11.2'
requires:
  - build-modern-redux-apps/redux-dataflow
---
Body
`)
	skill, err := ParseBytes(raw, "skills/build-modern-redux-apps/modern-redux/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	if skill.Extra["library"] != "@reduxjs/toolkit" || skill.Extra["library_version"] != "2.11.2" {
		t.Errorf("extra = %#v", skill.Extra)
	}
	if requires, ok := skill.Extra["requires"].([]any); !ok || len(requires) != 1 {
		t.Errorf("requires = %#v", skill.Extra["requires"])
	}
	if _, ok := skill.Extra["name"]; ok {
		t.Error("spec key leaked into Extra")
	}
	if !hasWarning(skill, "not spec-conformant") || !hasWarning(skill, "does not match directory") {
		t.Errorf("warnings = %v", skill.Warnings)
	}
	var extra []string
	for _, w := range skill.Warnings {
		if strings.HasPrefix(w, "non-spec frontmatter key") {
			extra = append(extra, w)
		}
	}
	want := []string{
		`non-spec frontmatter key "library"`,
		`non-spec frontmatter key "library_version"`,
		`non-spec frontmatter key "requires"`,
		`non-spec frontmatter key "type"`,
	}
	if !slices.Equal(extra, want) {
		t.Errorf("extra warnings = %v, want %v", extra, want)
	}
}

func TestParseOpenClawMetadata(t *testing.T) {
	raw := []byte(`---
name: weather
description: Get the weather
metadata:
  openclaw:
    requires:
      bins: [curl]
    install:
      - kind: node
        package: wttr-cli
---
`)
	skill, err := ParseBytes(raw, "weather/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	openclaw, ok := skill.Metadata["openclaw"].(map[string]any)
	if !ok {
		t.Fatalf("metadata = %#v", skill.Metadata)
	}
	install, ok := openclaw["install"].([]any)
	if !ok || len(install) != 1 {
		t.Errorf("install = %#v", openclaw["install"])
	}
}

func TestParseAllowedToolsList(t *testing.T) {
	skill, err := ParseBytes([]byte("---\nname: x\ndescription: d\nallowed-tools: [Read, Grep]\n---\n"), "x/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	if skill.AllowedTools != "Read Grep" {
		t.Errorf("allowed-tools = %q", skill.AllowedTools)
	}
	if !hasWarning(skill, "allowed-tools is a list") {
		t.Errorf("warnings = %v", skill.Warnings)
	}
}

func TestParseWarnings(t *testing.T) {
	long := strings.Repeat("a", maxDescLen+1)
	skill, err := ParseBytes([]byte("---\ndescription: "+long+"\n---\nbody"), "Bad_Name/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	if skill.Name != "Bad_Name" {
		t.Errorf("name = %q", skill.Name)
	}
	for _, want := range []string{"name missing", "not spec-conformant", "description exceeds"} {
		if !hasWarning(skill, want) {
			t.Errorf("missing warning %q in %v", want, skill.Warnings)
		}
	}
}

func TestParseMissingDescriptionWarns(t *testing.T) {
	skill, err := ParseBytes([]byte("---\nname: x\n---\nbody"), "x/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	if !hasWarning(skill, "description is missing") {
		t.Errorf("warnings = %v", skill.Warnings)
	}
}

func TestParseInvalidYAML(t *testing.T) {
	if _, err := ParseBytes([]byte("---\nname: [unclosed\n---\n"), "x/SKILL.md"); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseBOMAndCRLF(t *testing.T) {
	skill, err := ParseBytes([]byte("\xef\xbb\xbf---\r\nname: x\r\ndescription: d\r\n---\r\nbody\r\n"), "x/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	if skill.Name != "x" || skill.Description != "d" || skill.Body != "body" {
		t.Errorf("skill = %+v", skill)
	}
}

func TestSourceHashIncludesSchema(t *testing.T) {
	fsys := fstest.MapFS{"x/SKILL.md": {Data: []byte("---\nname: x\ndescription: d\n---\n")}}
	without, err := Parse(fsys, "x")
	if err != nil {
		t.Fatal(err)
	}
	fsys["x/schema.json"] = &fstest.MapFile{Data: []byte("{}")}
	with, err := Parse(fsys, "x")
	if err != nil {
		t.Fatal(err)
	}
	if without.SourceHash == with.SourceHash {
		t.Error("schema change did not change SourceHash")
	}
}

func TestValidateNamespace(t *testing.T) {
	allowed := map[string]bool{"example.level": true}
	if err := ValidateNamespace(map[string]any{"example.level": 1, "other.key": 2}, "example.", allowed); err != nil {
		t.Fatal(err)
	}
	if err := ValidateNamespace(map[string]any{"example.levle": 1}, "example.", allowed); err == nil {
		t.Fatal("expected error for unknown key")
	}
}

func hasWarning(skill *Skill, substr string) bool {
	return slices.ContainsFunc(skill.Warnings, func(w string) bool { return strings.Contains(w, substr) })
}
