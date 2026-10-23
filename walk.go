package skills

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

// DefaultMaxDepth bounds how many directories below the root Walk descends.
const DefaultMaxDepth = 6

// DefaultSkipDirs are directory names Walk never enters: VCS metadata and
// vendored dependency trees whose skills belong to some other package.
var DefaultSkipDirs = []string{gitMetadataDir, "node_modules", ".venv", "__pycache__", "vendor"}

type walkConfig struct {
	maxDepth int
	skipDirs map[string]bool
}

// Option configures Walk and Discover.
type Option func(*walkConfig)

// WithMaxDepth sets the maximum directory depth below the root.
func WithMaxDepth(depth int) Option {
	return func(c *walkConfig) { c.maxDepth = depth }
}

// WithSkipDirs replaces DefaultSkipDirs.
func WithSkipDirs(names ...string) Option {
	return func(c *walkConfig) {
		c.skipDirs = make(map[string]bool, len(names))
		for _, name := range names {
			c.skipDirs[name] = true
		}
	}
}

func newWalkConfig(opts []Option) walkConfig {
	c := walkConfig{maxDepth: DefaultMaxDepth}
	WithSkipDirs(DefaultSkipDirs...)(&c)
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

// Walk calls fn for each SKILL.md under root in fsys, in lexical order. A
// parse failure is passed to fn with a nil skill; returning it stops the
// walk, returning nil skips that skill. Unreadable directories are skipped.
func Walk(fsys fs.FS, root string, fn func(*Skill, error) error, opts ...Option) error {
	if fn == nil {
		return errors.New("skills: walk callback is required")
	}
	cfg := newWalkConfig(opts)
	root = path.Clean(root)
	return fs.WalkDir(fsys, root, func(p string, entry fs.DirEntry, err error) error {
		if err != nil {
			if p == root {
				return fmt.Errorf("skills: walk %s: %w", root, err)
			}
			return nil
		}
		if entry.IsDir() {
			if p != root && (cfg.skipDirs[entry.Name()] || depth(root, p) > cfg.maxDepth) {
				return fs.SkipDir
			}
			return nil
		}
		if entry.Name() != Filename {
			return nil
		}
		return fn(Parse(fsys, p))
	})
}

// Discover returns every skill under the root of fsys. Skills that fail to
// parse are left out and their errors joined into the returned error, so a
// single bad SKILL.md does not hide the rest.
func Discover(fsys fs.FS, opts ...Option) ([]*Skill, error) {
	var found []*Skill
	var errs []error
	err := Walk(fsys, ".", func(skill *Skill, err error) error {
		if err != nil {
			errs = append(errs, err)
			return nil
		}
		found = append(found, skill)
		return nil
	}, opts...)
	if err != nil {
		errs = append(errs, err)
	}
	return found, errors.Join(errs...)
}

func depth(root, p string) int {
	rel := strings.TrimPrefix(p, root+"/")
	if root == "." {
		rel = p
	}
	if rel == "" || rel == "." || rel == root {
		return 0
	}
	return strings.Count(rel, "/") + 1
}

func isNotExist(err error) bool {
	return errors.Is(err, fs.ErrNotExist)
}
