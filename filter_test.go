package skills

import (
	"errors"
	"path"
	"slices"
	"strings"
	"testing"
	"time"
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

func TestMatchManyDoublestar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		pattern string
		name    string
		want    bool
	}{
		{pattern: "**/a/**/b/**", name: "x/a/y/b/z", want: true},
		{pattern: "**/a/**/b/**", name: "x/a/b/z", want: true},
		{pattern: "**/a/**/b/**", name: "x/a/y/c/z", want: false},
		{pattern: "**/a/**", name: "a", want: true},
		{pattern: "p/**/q/**/r", name: "p/m/n/q/r", want: true},
	}
	for _, test := range tests {
		if got := Match(test.pattern, test.name); got != test.want {
			t.Errorf("Match(%q, %q) = %v, want %v", test.pattern, test.name, got, test.want)
		}
	}
}

// A pattern with the maximum number of ** segments against a long name used to
// revisit every (pattern-suffix, name-suffix) pair once per earlier **, which
// on 300 segments exceeds a 30s budget; memoised, it stays in the segment
// count and returns immediately.
func TestMatchManyDoublestarStaysBounded(t *testing.T) {
	t.Parallel()

	pattern := strings.Repeat("**/", maxDoublestar) + "zz"
	name := strings.Repeat("aa/", 299) + "aa"
	start := time.Now()
	if Match(pattern, name) {
		t.Error("Match() = true, want false for a non-matching literal segment")
	}
	if elapsed := time.Since(start); elapsed > 10*time.Second {
		t.Errorf("Match() took %v; the walk is not bounded", elapsed)
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
	if err := ValidateGlob(strings.Repeat("**/", maxDoublestar+1) + "x"); !errors.Is(err, path.ErrBadPattern) {
		t.Errorf("ValidateGlob() for many ** = %v", err)
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
	if !DirAllExcluded("src/generated", ignore) {
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
