package skills

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	dirPerm  = 0o755
	filePerm = 0o644
)

// Stage writes skill into destination on disk: its sibling files read from
// fsys under skill.Dir, a rendered SKILL.md, and schema.json when set.
// destination is cleared first and must resolve, after following symlinks,
// to a path inside workspace. Symlinks in the source are refused rather than
// recreated.
func Stage(fsys fs.FS, skill *Skill, workspace, destination string) error {
	if skill == nil {
		return errors.New("skills: skill is required")
	}
	if workspace == "" {
		return errors.New("skills: workspace is required")
	}
	if err := validateDestination(workspace, destination); err != nil {
		return err
	}
	if err := os.RemoveAll(destination); err != nil {
		return fmt.Errorf("skills: clear destination: %w", err)
	}
	if err := os.MkdirAll(destination, dirPerm); err != nil {
		return fmt.Errorf("skills: create destination: %w", err)
	}
	if fsys != nil && skill.Dir != "" {
		if err := copySiblings(fsys, skill.Dir, destination); err != nil {
			return fmt.Errorf("skills: copy siblings: %w", err)
		}
	}
	content, err := Render(skill)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(destination, Filename), content, filePerm); err != nil {
		return fmt.Errorf("skills: write SKILL.md: %w", err)
	}
	if skill.SchemaJSON != "" {
		if err := os.WriteFile(filepath.Join(destination, schemaFilename), []byte(skill.SchemaJSON), filePerm); err != nil {
			return fmt.Errorf("skills: write schema.json: %w", err)
		}
	}
	return nil
}

func validateDestination(workspace, destination string) error {
	root, err := realAbs(workspace)
	if err != nil {
		return err
	}
	target, err := realAbs(destination)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return err
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("skills: destination %q is outside the workspace", destination)
	}
	return nil
}

// realAbs resolves symlinks in the deepest existing ancestor of p and rejoins
// the missing tail. filepath.Abs alone would let a workspace containing
// .claude/skills -> /elsewhere pass validation while RemoveAll and WriteFile
// follow the link.
func realAbs(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	existing := abs
	var tail []string
	for {
		if _, err := os.Lstat(existing); err == nil {
			break
		}
		tail = append([]string{filepath.Base(existing)}, tail...)
		parent := filepath.Dir(existing)
		if parent == existing {
			break
		}
		existing = parent
	}
	resolved, err := filepath.EvalSymlinks(existing)
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{resolved}, tail...)...), nil
}

func copySiblings(fsys fs.FS, dir, destination string) error {
	return fs.WalkDir(fsys, dir, func(p string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == dir {
			return nil
		}
		rel := strings.TrimPrefix(p, dir+"/")
		if dir == "." {
			rel = p
		}
		if rel == Filename || rel == gitMetadataDir {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		target := filepath.Join(destination, filepath.FromSlash(rel))
		switch {
		case entry.Type()&fs.ModeSymlink != 0:
			return fmt.Errorf("symlink %q not staged", p)
		case entry.IsDir():
			return os.MkdirAll(target, dirPerm)
		case entry.Type().IsRegular():
			return copyFile(fsys, p, target)
		default:
			return fmt.Errorf("unsupported file type %s", p)
		}
	})
}

func copyFile(fsys fs.FS, name, destination string) error {
	mode := fs.FileMode(filePerm)
	if info, err := fs.Stat(fsys, name); err == nil && info.Mode().Perm() != 0 {
		mode = info.Mode().Perm()
	}
	input, err := fsys.Open(path.Clean(name))
	if err != nil {
		return err
	}
	defer func() { _ = input.Close() }()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(output, input); err != nil {
		_ = output.Close()
		return err
	}
	return output.Close()
}
