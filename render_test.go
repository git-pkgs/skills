package skills

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestRenderRoundTrip(t *testing.T) {
	raw := []byte(`---
name: audit
description: Check a repository
license: MIT
library: '@scope/pkg'
metadata:
  example.level: strict
---

Read the repository.
`)
	skill, err := ParseBytes(raw, "audit/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	out, err := Render(skill)
	if err != nil {
		t.Fatal(err)
	}
	want := `---
name: audit
description: Check a repository
license: MIT
library: '@scope/pkg'
metadata:
    example.level: strict
---

Read the repository.
`
	if string(out) != want {
		t.Errorf("render =\n%s\nwant\n%s", out, want)
	}
	again, err := ParseBytes(out, "audit/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	if again.Extra["library"] != "@scope/pkg" || again.Metadata["example.level"] != "strict" {
		t.Errorf("round trip lost fields: %+v", again)
	}
}

func TestRenderBodyOnly(t *testing.T) {
	out, err := Render(&Skill{Name: "x", Body: "just text"})
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "just text\n" {
		t.Errorf("render = %q", out)
	}
}

func TestRenderNil(t *testing.T) {
	if _, err := Render(nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestRenderSkipsSpecKeyInExtra(t *testing.T) {
	skill := &Skill{Name: "x", Description: "d", Extra: map[string]any{"name": "evil", "library": "lib"}}
	out, err := Render(skill)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "evil") || strings.Count(string(out), "name:") != 1 {
		t.Errorf("Render emitted a spec key from Extra:\n%s", out)
	}
	if !strings.Contains(string(out), "library: lib") {
		t.Errorf("Render dropped a non-spec Extra key:\n%s", out)
	}
}

func TestConcat(t *testing.T) {
	got := Concat(&Skill{Body: "one"}, nil, &Skill{Body: "  "}, &Skill{Body: "two\n"})
	if got != "one\n\n---\n\ntwo" {
		t.Errorf("concat = %q", got)
	}
}

func TestParseFromMapFSThenRender(t *testing.T) {
	fsys := fstest.MapFS{"x/SKILL.md": {Data: []byte("plain body\n")}}
	skill, err := ParseFS(fsys, "x")
	if err != nil {
		t.Fatal(err)
	}
	out, err := Render(skill)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "plain body\n" {
		t.Errorf("render = %q", out)
	}
}
