package registry

import (
	"os"
	"path/filepath"
	"strings"
)

// SkillEntry is the parsed-on-disk representation of a SKILL.md.
type SkillEntry struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	License     string   `yaml:"license"`
	Author      string   `yaml:"author"`
	Version     string   `yaml:"version"`
	Tiers       []string `yaml:"tiers"`
	Path        string   `yaml:"-"`
}

// findSkillFiles walks each root looking for immediate-child directories
// containing a SKILL.md. One-level scan is intentional: nested fixture
// directories or vendor copies should not appear as top-level skills.
func findSkillFiles(roots []string) ([]string, error) {
	var out []string
	for _, root := range roots {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			skillDir := filepath.Join(root, entry.Name())
			info, err := os.Stat(skillDir)
			if err != nil || !info.IsDir() {
				continue
			}
			candidate := filepath.Join(skillDir, "SKILL.md")
			if _, err := os.Stat(candidate); err == nil {
				out = append(out, candidate)
			}
		}
	}
	return out, nil
}

// ParseFrontmatter extracts the front-matter fields the dys CLI cares
// about: name, description, license, and the metadata block (author,
// version, tiers). Unknown fields are ignored.
func ParseFrontmatter(source string) SkillEntry {
	var out SkillEntry
	if !strings.HasPrefix(source, "---\n") {
		return out
	}
	end := strings.Index(source[4:], "\n---")
	if end == -1 {
		return out
	}
	fm := source[4 : 4+end]
	inMetadata := false
	for _, line := range strings.Split(fm, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Indented lines are either metadata sub-keys or block-scalar
		// continuations. We test indentation against the original line,
		// not the trimmed value (TrimSpace would strip the leading spaces).
		isIndented := strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t")
		if isIndented {
			if inMetadata {
				k, v, ok := splitKV(trimmed)
				if !ok {
					continue
				}
				switch k {
				case "author":
					out.Author = unquote(v)
				case "version":
					out.Version = unquote(v)
				case "tiers":
					out.Tiers = parseTiersList(v)
				}
			}
			continue
		}
		k, v, ok := splitKV(trimmed)
		if !ok {
			continue
		}
		switch k {
		case "name":
			out.Name = unquote(v)
		case "description":
			out.Description = unquote(v)
		case "license":
			out.License = unquote(v)
		case "metadata":
			if v == "" {
				inMetadata = true
			}
		}
	}
	return out
}

func splitKV(line string) (string, string, bool) {
	idx := strings.Index(line, ":")
	if idx == -1 {
		return "", "", false
	}
	k := strings.TrimSpace(line[:idx])
	v := strings.TrimSpace(line[idx+1:])
	return k, v, true
}

func unquote(s string) string {
	if len(s) >= 2 && (s[0] == '"' && s[len(s)-1] == '"') {
		return s[1 : len(s)-1]
	}
	if len(s) >= 2 && (s[0] == '\'' && s[len(s)-1] == '\'') {
		return s[1 : len(s)-1]
	}
	return s
}

func parseTiersList(raw string) []string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "[")
	raw = strings.TrimSuffix(raw, "]")
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = unquote(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// LoadSkill reads a single SKILL.md and returns its SkillEntry.
func LoadSkill(file string) (SkillEntry, bool) {
	data, err := os.ReadFile(file)
	if err != nil {
		return SkillEntry{}, false
	}
	entry := ParseFrontmatter(string(data))
	entry.Path = file
	if strings.TrimSpace(entry.Name) == "" {
		entry.Name = filepath.Base(filepath.Dir(file))
	}
	return entry, true
}

// List returns every SkillEntry reachable from the provided roots.
func List(roots []string) []SkillEntry {
	files, _ := findSkillFiles(roots)
	out := make([]SkillEntry, 0, len(files))
	for _, f := range files {
		if e, ok := LoadSkill(f); ok {
			out = append(out, e)
		}
	}
	return out
}
