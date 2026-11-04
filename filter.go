package skills

import (
	"fmt"
	"path"
	"slices"
	"strings"
)

const (
	gitMetadataDir    = ".git"
	gitMetadataPrefix = gitMetadataDir + "/"
)

// DirAllExcluded reports whether an ignore pattern excludes every file below
// rel. A pattern ending in /** is the only form that can blanket a directory.
func DirAllExcluded(rel string, ignore []string) bool {
	if rel == gitMetadataDir || strings.HasPrefix(rel, gitMetadataPrefix) {
		return false
	}
	return dirBlanketed(rel, ignore)
}

func dirBlanketed(rel string, patterns []string) bool {
	for _, pattern := range patterns {
		prefix, ok := strings.CutSuffix(pattern, "/**")
		if ok && Match(prefix, rel) {
			return true
		}
	}
	return false
}

// PathIncluded applies an optional include list and an ignore list to rel.
// Paths are slash-separated and relative to the caller's root. The .git
// directory is always retained for git-aware skills.
func PathIncluded(rel string, paths, ignore []string) bool {
	if rel == gitMetadataDir || strings.HasPrefix(rel, gitMetadataPrefix) {
		return true
	}
	if len(paths) > 0 && !matchAny(paths, rel) {
		return false
	}
	return !matchAny(ignore, rel)
}

func matchAny(patterns []string, name string) bool {
	return slices.ContainsFunc(patterns, func(pattern string) bool {
		return Match(pattern, name)
	})
}

const maxDoublestar = 8

// ValidateGlob returns path.ErrBadPattern for a malformed segment or a
// pattern with more than eight ** segments. Match is not linear in the number
// of ** segments; validate untrusted patterns before matching them against a
// large tree.
func ValidateGlob(pattern string) error {
	doublestar := 0
	for segment := range strings.SplitSeq(pattern, "/") {
		if segment == "**" {
			doublestar++
		}
		if _, err := path.Match(segment, "x"); err != nil {
			return err
		}
	}
	if doublestar > maxDoublestar {
		return fmt.Errorf("%w: more than %d ** segments", path.ErrBadPattern, maxDoublestar)
	}
	return nil
}

// Match reports whether name matches a shell glob with doublestar support.
func Match(pattern, name string) bool {
	if pattern == "" || name == "" {
		return pattern == name
	}
	return matchSegments(strings.Split(pattern, "/"), strings.Split(name, "/"))
}

// matchSegments reports whether name matches the pattern segments. Segments
// ahead of the first ** pair off one to one, so the plain loop runs until then
// and only the ** tail pays for memoisation.
func matchSegments(pattern, name []string) bool {
	for len(pattern) > 0 && pattern[0] != "**" {
		if len(name) == 0 {
			return false
		}
		ok, _ := path.Match(pattern[0], name[0])
		if !ok {
			return false
		}
		pattern, name = pattern[1:], name[1:]
	}
	if len(pattern) == 0 {
		return len(name) == 0
	}
	return matchTail(pattern, name)
}

// matchTail walks the segments from the first **, memoised on suffix lengths
// so several ** segments stay at len(pattern) × len(name) comparisons.
func matchTail(pattern, name []string) bool {
	width := len(name) + 1
	seen := make([]int8, (len(pattern)+1)*width) // 0 unknown, 1 no match, 2 match
	var walk func(pattern, name []string) bool
	walk = func(pattern, name []string) bool {
		key := len(pattern)*width + len(name)
		if hit := seen[key]; hit != 0 {
			return hit == 2
		}
		var matched bool
		switch {
		case len(pattern) == 0:
			matched = len(name) == 0
		case pattern[0] == "**":
			rest := pattern[1:]
			if len(rest) == 0 {
				matched = true
			} else {
				for index := 0; index <= len(name) && !matched; index++ {
					matched = walk(rest, name[index:])
				}
			}
		case len(name) > 0:
			ok, _ := path.Match(pattern[0], name[0])
			matched = ok && walk(pattern[1:], name[1:])
		}
		if matched {
			seen[key] = 2
		} else {
			seen[key] = 1
		}
		return matched
	}
	return walk(pattern, name)
}

// SplitPatterns parses the newline-separated storage form into a clean slice,
// trimming whitespace and dropping empty lines.
func SplitPatterns(value string) []string {
	if value == "" {
		return nil
	}
	var patterns []string
	for line := range strings.SplitSeq(value, "\n") {
		if pattern := strings.TrimSpace(line); pattern != "" {
			patterns = append(patterns, pattern)
		}
	}
	return patterns
}

// JoinPatterns serialises patterns into the newline-separated storage form.
func JoinPatterns(patterns []string) string {
	return strings.Join(patterns, "\n")
}
