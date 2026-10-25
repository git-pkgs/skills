package skills

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

const (
	dirPerm  = 0o755
	filePerm = 0o644
)

type stageConfig struct {
	dropExtra bool
}

// StageOption configures Stage.
type StageOption func(*stageConfig)

// WithoutExtra leaves non-spec top-level frontmatter keys out of the staged
// SKILL.md, so agent-specific settings such as model from an untrusted skill
// do not reach the agent.
func WithoutExtra() StageOption {
	return func(c *stageConfig) { c.dropExtra = true }
}

// ValidateName reports whether name can be used as a single directory name
// for a staged skill. Check it before building a destination from the name:
// filepath.Join cleans "../x" away, so Stage cannot catch it afterwards.
func ValidateName(name string) error {
	if name == "" || !filepath.IsLocal(name) || filepath.Base(name) != name {
		return fmt.Errorf("skills: skill name %q is not a single path element", name)
	}
	return nil
}

// Stage writes skill into destination on disk: its sibling files, a rendered
// SKILL.md, and schema.json when set. Siblings are read from fsys under
// skill.Dir, or from skill.SourcePath when fsys is nil.
//
// workspace is created if missing. destination must resolve to a path strictly
// inside workspace and is cleared first. Every write goes through an
// [os.Root] opened on workspace, so a symlink in or below the destination
// cannot redirect a write outside it. Symlinks in the source are refused
// rather than followed.
func Stage(fsys fs.FS, skill *Skill, workspace, destination string, opts ...StageOption) error {
	if skill == nil {
		return errors.New("skills: skill is required")
	}
	var cfg stageConfig
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.dropExtra {
		trimmed := *skill
		trimmed.Extra = nil
		skill = &trimmed
	}
	if workspace == "" {
		return errors.New("skills: workspace is required")
	}
	rel, err := destinationInWorkspace(workspace, destination)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(workspace, dirPerm); err != nil {
		return fmt.Errorf("skills: create workspace: %w", err)
	}
	root, err := os.OpenRoot(workspace)
	if err != nil {
		return fmt.Errorf("skills: open workspace: %w", err)
	}
	defer func() { _ = root.Close() }()
	if err := root.RemoveAll(rel); err != nil {
		return fmt.Errorf("skills: clear destination: %w", err)
	}
	if err := root.MkdirAll(rel, dirPerm); err != nil {
		return fmt.Errorf("skills: create destination: %w", err)
	}
	source, dir := fsys, skill.Dir
	if source == nil && skill.SourcePath != "" {
		source, dir = os.DirFS(skill.SourcePath), "."
	}
	if source != nil && dir != "" {
		if err := copySiblings(source, dir, root, rel); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("skills: copy siblings: %w", err)
		}
	}
	content, err := Render(skill)
	if err != nil {
		return err
	}
	if err := root.WriteFile(filepath.Join(rel, Filename), content, filePerm); err != nil {
		return fmt.Errorf("skills: write SKILL.md: %w", err)
	}
	if skill.SchemaJSON != "" {
		if err := root.WriteFile(filepath.Join(rel, schemaFilename), []byte(skill.SchemaJSON), filePerm); err != nil {
			return fmt.Errorf("skills: write schema.json: %w", err)
		}
	}
	return nil
}

func destinationInWorkspace(workspace, destination string) (string, error) {
	ws, err := filepath.Abs(workspace)
	if err != nil {
		return "", fmt.Errorf("skills: resolve workspace: %w", err)
	}
	dest, err := filepath.Abs(destination)
	if err != nil {
		return "", fmt.Errorf("skills: resolve destination: %w", err)
	}
	rel, err := filepath.Rel(ws, dest)
	if err != nil || rel == "." || !filepath.IsLocal(rel) {
		return "", fmt.Errorf("skills: destination %q is outside the workspace", destination)
	}
	return rel, nil
}

// copySiblings writes the skill's sibling files into root at rel. It walks the
// same tree as listFiles so Files and the staged tree agree. schema.json is
// skipped: Stage writes skill.SchemaJSON afterwards, which may have been
// bundled or edited since parsing.
func copySiblings(fsys fs.FS, dir string, root *os.Root, rel string) error {
	return walkSiblings(fsys, dir, func(name string, entry fs.DirEntry) error {
		target := filepath.Join(rel, filepath.FromSlash(name))
		switch {
		case entry.Type()&fs.ModeSymlink != 0:
			return fmt.Errorf("symlink %q not staged", path.Join(dir, name))
		case entry.IsDir():
			return root.MkdirAll(target, dirPerm)
		case name == schemaFilename:
			return nil
		case entry.Type().IsRegular():
			return copyFile(fsys, path.Join(dir, name), root, target)
		default:
			return fmt.Errorf("unsupported file type at %s", path.Join(dir, name))
		}
	})
}

func copyFile(fsys fs.FS, name string, root *os.Root, target string) error {
	input, err := fsys.Open(name)
	if err != nil {
		return err
	}
	defer func() { _ = input.Close() }()
	mode := fs.FileMode(filePerm)
	if info, err := input.Stat(); err == nil && info.Mode().Perm() != 0 {
		mode = info.Mode().Perm()
	}
	output, err := root.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(output, input); err != nil {
		_ = output.Close()
		return err
	}
	return output.Close()
}
