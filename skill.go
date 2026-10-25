// Package skills finds, parses, validates and stages agent skills: directories
// containing a SKILL.md file as described at [SpecURL]. Parse and Walk work on
// paths on disk; ParseFS, WalkFS and Discover work on any fs.FS, such as an
// unpacked package or an archive.
package skills

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

// SpecURL is the Agent Skills format specification.
const SpecURL = "https://agentskills.io/specification"

const (
	// Filename is the instruction file that marks a skill directory.
	Filename = "SKILL.md"
	// MaxFileSize bounds the SKILL.md and schema.json this package reads
	// into memory. Package archives are untrusted; a larger file is refused
	// rather than loaded.
	MaxFileSize    = 4 << 20
	schemaFilename = "schema.json"
	scriptsDir     = "scripts"
	maxNameLen     = 64
	maxDescLen     = 1024
	maxCompatLen   = 500
)

var (
	frontmatterPattern = regexp.MustCompile(`(?s)\A\x{FEFF}?---\r?\n(.*?)\r?\n---\r?\n?(.*)\z`)
	namePattern        = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
)

var specKeys = map[string]bool{
	"name":          true,
	"description":   true,
	"license":       true,
	"compatibility": true,
	"allowed-tools": true,
	"metadata":      true,
}

// Skill is one parsed SKILL.md and the directory it sits in.
type Skill struct {
	Name          string
	Description   string
	License       string
	Compatibility string
	// AllowedTools is space-delimited, as in the spec. A YAML list is
	// accepted and joined.
	AllowedTools string
	Metadata     map[string]any
	// Extra holds top-level frontmatter keys outside the spec, such as
	// library or requires, which several publishers add.
	Extra map[string]any
	Body  string

	// Dir is the skill directory, slash-separated and relative to the fs.FS
	// root it was read from.
	Dir string
	// SourcePath is the absolute directory on disk for skills read with
	// Parse or Walk, and empty for ParseFS.
	SourcePath string
	// Files lists everything under Dir except SKILL.md, relative to Dir.
	Files []string
	// SchemaJSON is the content of a sibling schema.json, a common extension
	// rather than a spec field.
	SchemaJSON string
	// SourceHash is the SHA-256 of SKILL.md followed by schema.json when
	// present, so a schema-only edit changes it.
	SourceHash string
	Warnings   []string
}

// HasScripts reports whether the skill ships a scripts/ directory, which
// agents may execute.
func (s *Skill) HasScripts() bool {
	prefix := scriptsDir + "/"
	for _, f := range s.Files {
		if strings.HasPrefix(f, prefix) {
			return true
		}
	}
	return false
}

// ParseFS reads the skill at name in fsys. name may be a SKILL.md path, the
// directory holding one, or a standalone markdown file such as review.md, in
// which case siblings are not listed and the name comes from the file stem.
func ParseFS(fsys fs.FS, name string) (*Skill, error) {
	return parseFS(fsys, name, "")
}

// Parse reads the skill at path on disk, accepting the same forms as ParseFS,
// and sets SourcePath to its absolute directory.
func Parse(p string) (*Skill, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return nil, fmt.Errorf("skills: resolve %s: %w", p, err)
	}
	if !strings.EqualFold(filepath.Ext(abs), ".md") {
		abs = filepath.Join(abs, Filename)
	}
	// Read from the parent so the skill directory's own name is available
	// for name inference and warnings.
	dir := filepath.Dir(abs)
	root := filepath.Dir(dir)
	name := filepath.Base(dir) + "/" + filepath.Base(abs)
	if root == dir {
		name = filepath.Base(abs)
	}
	skill, err := parseFS(os.DirFS(root), name, root)
	if err != nil {
		return nil, err
	}
	skill.SourcePath = dir
	return skill, nil
}

// parseFS reads name from fsys. Errors show paths joined to label when set,
// so disk callers see absolute paths.
func parseFS(fsys fs.FS, name, label string) (*Skill, error) {
	name = path.Clean(name)
	if !strings.EqualFold(path.Ext(name), ".md") {
		name = path.Join(name, Filename)
	}
	dir := path.Dir(name)
	show := func(p string) string {
		if label == "" {
			return p
		}
		return filepath.Join(label, filepath.FromSlash(p))
	}
	raw, err := readLimited(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("skills: read %s: %w", show(name), err)
	}
	skill, err := ParseBytes(raw, name)
	if err != nil {
		return nil, fmt.Errorf("skills: %s: %w", show(name), err)
	}
	if isSkillFile(name) {
		schema, err := readLimited(fsys, path.Join(dir, schemaFilename))
		if err != nil && !isNotExist(err) {
			return nil, fmt.Errorf("skills: read %s: %w", show(path.Join(dir, schemaFilename)), err)
		}
		skill.SchemaJSON = string(schema)
		skill.SourceHash = hash(raw, schema)
		if skill.Files, err = listFiles(fsys, dir); err != nil {
			return nil, fmt.Errorf("skills: list %s: %w", show(dir), err)
		}
	}
	return skill, nil
}

// ParseBytes parses skill content without reading siblings. name is the file
// path, used for Dir and for the name when the frontmatter omits one.
func ParseBytes(raw []byte, name string) (*Skill, error) {
	fm, body, hasFrontmatter := splitFrontmatter(raw)
	skill := &Skill{
		Dir:        path.Dir(path.Clean(name)),
		Metadata:   map[string]any{},
		Extra:      map[string]any{},
		Body:       strings.TrimSpace(body),
		SourceHash: hash(raw),
	}
	if hasFrontmatter {
		var fields map[string]any
		if err := yaml.Unmarshal(fm, &fields); err != nil {
			return nil, fmt.Errorf("yaml: %w", err)
		}
		skill.applyFrontmatter(fields)
	}
	skill.validate(name, hasFrontmatter)
	return skill, nil
}

func (s *Skill) applyFrontmatter(fields map[string]any) {
	s.Name = s.stringField(fields, "name")
	s.Description = s.stringField(fields, "description")
	s.License = s.stringField(fields, "license")
	s.Compatibility = s.stringField(fields, "compatibility")
	s.AllowedTools = s.toolsField(fields["allowed-tools"])
	switch meta := fields["metadata"].(type) {
	case map[string]any:
		s.Metadata = meta
	case nil:
	default:
		s.Warnings = append(s.Warnings, "metadata is not a mapping, ignored")
	}
	for key, value := range fields {
		if !specKeys[key] {
			s.Extra[key] = value
		}
	}
	// The spec expects extensions under metadata; record each top-level one.
	extra := make([]string, 0, len(s.Extra))
	for key := range s.Extra {
		extra = append(extra, key)
	}
	slices.Sort(extra)
	for _, key := range extra {
		s.Warnings = append(s.Warnings, fmt.Sprintf("non-spec frontmatter key %q", key))
	}
}

func (s *Skill) stringField(fields map[string]any, key string) string {
	switch v := fields[key].(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	default:
		s.Warnings = append(s.Warnings, fmt.Sprintf("%s is not a string", key))
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func (s *Skill) toolsField(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case []any:
		tools := make([]string, 0, len(v))
		for _, tool := range v {
			tools = append(tools, strings.TrimSpace(fmt.Sprint(tool)))
		}
		s.Warnings = append(s.Warnings, "allowed-tools is a list, expected a space-delimited string")
		return strings.Join(tools, " ")
	default:
		s.Warnings = append(s.Warnings, "allowed-tools is not a string")
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func splitFrontmatter(raw []byte) ([]byte, string, bool) {
	match := frontmatterPattern.FindSubmatch(raw)
	if match == nil {
		return nil, string(bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf"))), false
	}
	return match[1], string(match[2]), true
}

func readLimited(fsys fs.FS, name string) ([]byte, error) {
	info, err := fs.Stat(fsys, name)
	if err != nil {
		return nil, err
	}
	if info.Size() > MaxFileSize {
		return nil, fmt.Errorf("%d bytes exceeds the %d-byte limit", info.Size(), MaxFileSize)
	}
	return fs.ReadFile(fsys, name)
}

// hash writes a NUL after each part so the boundary between SKILL.md and
// schema.json is fixed.
func hash(parts ...[]byte) string {
	hasher := sha256.New()
	for _, part := range parts {
		_, _ = hasher.Write(part)
		_, _ = hasher.Write([]byte{0})
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

// validate records spec violations as warnings. The client guide asks
// implementations to load non-conformant skills rather than reject them.
func (s *Skill) validate(name string, hadFrontmatter bool) {
	inferred := inferredName(name)
	switch {
	case s.Name == "" && inferred == "":
		s.Warnings = append(s.Warnings, "name missing and no directory name to use")
	case s.Name == "":
		s.Name = inferred
		if hadFrontmatter {
			s.Warnings = append(s.Warnings, "name missing, using path name")
		}
	case isSkillFile(name) && inferred != "" && s.Name != inferred:
		s.Warnings = append(s.Warnings, fmt.Sprintf("name %q does not match directory %q", s.Name, inferred))
	}
	if hadFrontmatter && s.Description == "" {
		s.Warnings = append(s.Warnings, "description is missing")
	}
	if len(s.Name) > maxNameLen {
		s.Warnings = append(s.Warnings, fmt.Sprintf("name %q exceeds %d characters", s.Name, maxNameLen))
	}
	if s.Name != "" && !namePattern.MatchString(s.Name) {
		s.Warnings = append(s.Warnings,
			fmt.Sprintf("name %q is not spec-conformant (lowercase, digits, hyphens only)", s.Name))
	}
	if len(s.Description) > maxDescLen {
		s.Warnings = append(s.Warnings, fmt.Sprintf("description exceeds %d characters", maxDescLen))
	}
	if len(s.Compatibility) > maxCompatLen {
		s.Warnings = append(s.Warnings, fmt.Sprintf("compatibility exceeds %d characters", maxCompatLen))
	}
}

func isSkillFile(name string) bool {
	return strings.EqualFold(path.Base(name), Filename)
}

func inferredName(name string) string {
	if isSkillFile(name) {
		dir := path.Dir(path.Clean(name))
		if dir == "." {
			return ""
		}
		return path.Base(dir)
	}
	base := path.Base(name)
	return strings.TrimSuffix(base, path.Ext(base))
}

// walkSiblings visits the files that belong to a skill at dir: everything
// under it except SKILL.md, .git at any depth, and nested skill directories,
// whose files belong to that nested skill. listFiles and copySiblings share
// this walk so Files and the staged tree cannot disagree.
func walkSiblings(fsys fs.FS, dir string, fn func(rel string, entry fs.DirEntry) error) error {
	return fs.WalkDir(fsys, dir, func(p string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(p, dir+"/")
		if dir == "." {
			rel = p
		}
		if entry.IsDir() {
			if p == dir {
				return nil
			}
			if entry.Name() == gitMetadataDir {
				return fs.SkipDir
			}
			if _, err := fs.Stat(fsys, path.Join(p, Filename)); err == nil {
				return fs.SkipDir
			}
			return fn(rel, entry)
		}
		if rel == Filename {
			return nil
		}
		return fn(rel, entry)
	})
}

func listFiles(fsys fs.FS, dir string) ([]string, error) {
	var files []string
	err := walkSiblings(fsys, dir, func(rel string, entry fs.DirEntry) error {
		if !entry.IsDir() {
			files = append(files, rel)
		}
		return nil
	})
	return files, err
}

// ValidateNamespace rejects metadata keys starting with prefix that are not
// in allowed. Keys outside the namespace are ignored.
func ValidateNamespace(meta map[string]any, prefix string, allowed map[string]bool) error {
	for key := range meta {
		if strings.HasPrefix(key, prefix) && !allowed[key] {
			return fmt.Errorf("skills: unknown metadata key %q", key)
		}
	}
	return nil
}
