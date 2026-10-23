package skills

import (
	"fmt"

	"github.com/git-pkgs/archives"
)

// DiscoverArchive returns the skills inside an opened package archive. Use
// archives.OpenWithPrefix to strip wrapper directories such as npm's
// package/ so skill paths are relative to the package root.
func DiscoverArchive(r archives.Reader, opts ...Option) ([]*Skill, error) {
	fsys, err := archives.NewFS(r)
	if err != nil {
		return nil, fmt.Errorf("skills: index archive: %w", err)
	}
	return Discover(fsys, opts...)
}
