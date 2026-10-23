package skills

import (
	"errors"
	"path"
	"slices"
	"testing"
)

func TestMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		pattern string
		name    string
		want    bool
	}{
		{pattern: "**/*.go", name: "internal/skills/parse.go", want: true},
		{pattern: "cmd/*/main.go", name: "cmd/app/main.go", want: true},
		{pattern: "cmd/*/main.go", name: "cmd/a/b/main.go", want: false},
		{pattern: "docs/**", name: "docs/index.md", want: true},
		{pattern: "*.go", name: "main.rb", want: false},
	}
	for _, test := range tests {
		if got := Match(test.pattern, test.name); got != test.want {
			t.Errorf("Match(%q, %q) = %v, want %v", test.pattern, test.name, got, test.want)
		}
	}
}

func TestValidateGlob(t *testing.T) {
	t.Parallel()
	if err := ValidateGlob("**/*.go"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateGlob("[broken"); !errors.Is(err, path.ErrBadPattern) {
		t.Errorf("ValidateGlob() error = %v", err)
	}
}

func TestPathFilters(t *testing.T) {
	t.Parallel()

	paths := []string{"src/**"}
	ignore := []string{"src/generated/**"}
	if !PathIncluded("src/main.go", paths, ignore) {
		t.Fatal("src/main.go was excluded")
	}
	if PathIncluded("docs/readme.md", paths, ignore) {
		t.Fatal("docs/readme.md was included")
	}
	if PathIncluded("src/generated/types.go", paths, ignore) {
		t.Fatal("generated file was included")
	}
	if !PathIncluded(".git/config", nil, []string{".git/**"}) {
		t.Fatal(".git was excluded")
	}
	if !DirAllExcluded("src/generated", paths, ignore) {
		t.Fatal("generated directory was not blanketed")
	}
}

func TestPatternStorageRoundTrip(t *testing.T) {
	t.Parallel()

	patterns := SplitPatterns(" src/** \n\nvendor/**\n*.go ")
	want := []string{"src/**", "vendor/**", "*.go"}
	if !slices.Equal(patterns, want) {
		t.Errorf("SplitPatterns() = %v, want %v", patterns, want)
	}
	if got := JoinPatterns(patterns); got != "src/**\nvendor/**\n*.go" {
		t.Errorf("JoinPatterns() = %q", got)
	}
	if SplitPatterns("") != nil {
		t.Fatal("SplitPatterns(\"\") did not return nil")
	}
}
