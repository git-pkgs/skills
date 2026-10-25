package skills

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

// DefaultMaxDepth bounds how many directories below the root Walk descends.
const DefaultMaxDepth = 6

var defaultSkipDirs = []string{gitMetadataDir, "node_modules", ".venv", "__pycache__", "vendor"}

// DefaultSkipDirs returns the directory names Walk never enters by default:
// VCS metadata and vendored dependency trees whose skills belong to another
// package. The slice is a copy; append to it and pass to WithSkipDirs to
// extend the set.
func DefaultSkipDirs() []string {
	return slices.Clone(defaultSkipDirs)
}

type walkConfig struct {
	maxDepth int
	skipDirs map[string]bool
}

// WalkOption configures Walk, WalkFS and Discover.
type WalkOption func(*walkConfig)

// WithMaxDepth sets the maximum directory depth below the root.
func WithMaxDepth(depth int) WalkOption {
	return func(c *walkConfig) { c.maxDepth = depth }
}

// WithSkipDirs replaces the default skip set.
func WithSkipDirs(names ...string) WalkOption {
	return func(c *walkConfig) {
		c.skipDirs = make(map[string]bool, len(names))
		for _, name := range names {
			c.skipDirs[name] = true
		}
	}
}

func newWalkConfig(opts []WalkOption) walkConfig {
	c := walkConfig{maxDepth: DefaultMaxDepth}
	WithSkipDirs(defaultSkipDirs...)(&c)
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

// WalkFS calls fn for each SKILL.md under root in fsys, in lexical order. A
// parse failure is passed to fn with a nil skill; returning it stops the
// walk, returning nil skips that skill. Unreadable directories are skipped.
//
// Each skill's directory is walked a second time to populate Files, so an
// fs.FS backed by a sequential archive should cache its entry list.
func WalkFS(fsys fs.FS, root string, fn func(*Skill, error) error, opts ...WalkOption) error {
	if fn == nil {
		return errors.New("skills: walk callback is required")
	}
	root = path.Clean(root)
	if err := walkFS(fsys, root, "", fn, opts); err != nil {
		return fmt.Errorf("skills: walk %s: %w", root, err)
	}
	return nil
}

// walkFS is WalkFS without the error prefix. label is passed to parseFS so
// disk callers see absolute paths in parse errors.
func walkFS(fsys fs.FS, root, label string, fn func(*Skill, error) error, opts []WalkOption) error {
	cfg := newWalkConfig(opts)
	return fs.WalkDir(fsys, root, func(p string, entry fs.DirEntry, err error) error {
		if err != nil {
			if p == root {
				return err
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
		return fn(parseFS(fsys, p, label))
	})
}

// Walk calls fn for each SKILL.md under root on disk, with SourcePath set.
// It stops at the first parse error or error returned by fn.
func Walk(root string, fn func(*Skill) error, opts ...WalkOption) error {
	if fn == nil {
		return errors.New("skills: walk callback is required")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("skills: resolve %s: %w", root, err)
	}
	err = walkFS(os.DirFS(abs), ".", abs, func(skill *Skill, err error) error {
		if err != nil {
			return err
		}
		skill.SourcePath = filepath.Join(abs, filepath.FromSlash(skill.Dir))
		return fn(skill)
	}, opts)
	if err != nil {
		return fmt.Errorf("skills: walk %s: %w", abs, err)
	}
	return nil
}

// Discover returns every skill under the root of fsys. Skills that fail to
// parse are left out and their errors joined into the returned error, so a
// single bad SKILL.md does not hide the rest.
func Discover(fsys fs.FS, opts ...WalkOption) ([]*Skill, error) {
	var found []*Skill
	var errs []error
	err := WalkFS(fsys, ".", func(skill *Skill, err error) error {
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
