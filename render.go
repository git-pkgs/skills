package skills

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Render returns SKILL.md bytes for skill with one trailing newline. Spec
// fields are written first in spec order, then Extra keys sorted, then
// metadata. A skill with no frontmatter fields beyond the name renders as its
// body alone.
func Render(skill *Skill) ([]byte, error) {
	if skill == nil {
		return nil, errors.New("skills: skill is required")
	}
	if skill.Description == "" && skill.License == "" && skill.Compatibility == "" &&
		skill.AllowedTools == "" && len(skill.Metadata) == 0 && len(skill.Extra) == 0 {
		return withTrailingNewline(skill.Body), nil
	}
	doc := &yaml.Node{Kind: yaml.MappingNode}
	add := func(key string, value any) error {
		var v yaml.Node
		if err := v.Encode(value); err != nil {
			return fmt.Errorf("skills: encode %s: %w", key, err)
		}
		doc.Content = append(doc.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: key}, &v)
		return nil
	}
	for _, field := range []struct{ key, value string }{
		{"name", skill.Name},
		{"description", skill.Description},
		{"license", skill.License},
		{"compatibility", skill.Compatibility},
		{"allowed-tools", skill.AllowedTools},
	} {
		if field.value == "" {
			continue
		}
		if err := add(field.key, field.value); err != nil {
			return nil, err
		}
	}
	keys := make([]string, 0, len(skill.Extra))
	for key := range skill.Extra {
		if specKeys[key] {
			continue
		}
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		if err := add(key, skill.Extra[key]); err != nil {
			return nil, err
		}
	}
	if len(skill.Metadata) > 0 {
		if err := add("metadata", skill.Metadata); err != nil {
			return nil, err
		}
	}
	fm, err := yaml.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("skills: marshal frontmatter: %w", err)
	}
	var output strings.Builder
	output.WriteString("---\n")
	output.Write(fm)
	output.WriteString("---\n\n")
	output.WriteString(skill.Body)
	return withTrailingNewline(output.String()), nil
}

func withTrailingNewline(s string) []byte {
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	return []byte(s)
}

// Concat joins skill bodies into one system prompt.
func Concat(skills ...*Skill) string {
	parts := make([]string, 0, len(skills))
	for _, skill := range skills {
		if skill == nil {
			continue
		}
		if body := strings.TrimSpace(skill.Body); body != "" {
			parts = append(parts, body)
		}
	}
	return strings.Join(parts, "\n\n---\n\n")
}
