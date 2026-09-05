package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFrontmatterNameDescriptionLicense(t *testing.T) {
	src := `---
name: x
description: "Trigger: example. what this does"
license: Apache-2.0
---
body
`
	e := ParseFrontmatter(src)
	if e.Name != "x" {
		t.Errorf("Name = %q, want x", e.Name)
	}
	if e.Description == "" {
		t.Error("Description is empty")
	}
	if e.License != "Apache-2.0" {
		t.Errorf("License = %q", e.License)
	}
}

func TestParseFrontmatterMetadataAuthorVersionTiers(t *testing.T) {
	src := `---
name: x
description: Trigger: t
license: Apache-2.0
metadata:
  author: ardelperal
  version: "1.0"
  last_verified: 2026-09-05
  scope: ['universal', 'vba']
  tiers: [universal, vba, runtime]
  auto_invoke: ["doing x"]
---
`
	e := ParseFrontmatter(src)
	if e.Author != "ardelperal" {
		t.Errorf("Author = %q", e.Author)
	}
	if e.Version != "1.0" {
		t.Errorf("Version = %q", e.Version)
	}
	if len(e.Tiers) != 3 {
		t.Fatalf("Tiers len = %d, want 3; got %v", len(e.Tiers), e.Tiers)
	}
	if e.Tiers[0] != "universal" || e.Tiers[1] != "vba" || e.Tiers[2] != "runtime" {
		t.Errorf("Tiers = %v", e.Tiers)
	}
}

func TestParseFrontmatterAuthorWithUnicodeSpace(t *testing.T) {
	// Author value contains a space (full name). Make sure the parser does
	// not split on the space.
	src := `---
name: x
description: Trigger: t
metadata:
  author: "Andrés Román"
---
`
	e := ParseFrontmatter(src)
	if e.Author != "Andrés Román" {
		t.Errorf("Author = %q, want %q", e.Author, "Andrés Román")
	}
}

func TestParseFrontmatterTopLevelKeyAfterMetadata(t *testing.T) {
	// After metadata block ends, top-level keys should resume.
	src := `---
name: x
description: Trigger: t
metadata:
  author: a
license: Apache-2.0
---
`
	e := ParseFrontmatter(src)
	if e.Author != "a" {
		t.Errorf("Author = %q", e.Author)
	}
	if e.License != "Apache-2.0" {
		t.Errorf("License = %q (should resume after metadata)", e.License)
	}
}

func TestParseFrontmatterNoFrontmatter(t *testing.T) {
	e := ParseFrontmatter("body only, no frontmatter")
	if e.Name != "" || e.Author != "" {
		t.Errorf("expected zero-value SkillEntry, got %+v", e)
	}
}

func TestParseTiersListVariants(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty", "[]", nil},
		{"single", "[universal]", []string{"universal"}},
		{"multiple", "[a, b, c]", []string{"a", "b", "c"}},
		{"trimmed", "[ a , b ]", []string{"a", "b"}},
		{"quoted", `["a b", "c"]`, []string{"a b", "c"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseTiersList(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("len = %d, want %d (got %v)", len(got), len(tc.want), got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("[%d] = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestListFindsImmediateChildSkills(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "alpha"))
	mustMkdir(t, filepath.Join(root, "beta"))
	mustMkdir(t, filepath.Join(root, "nested", "ignored"))
	mustWrite(t, filepath.Join(root, "alpha", "SKILL.md"), `---
name: alpha
description: Trigger: a
---
`)
	mustWrite(t, filepath.Join(root, "beta", "SKILL.md"), `---
name: beta
description: Trigger: b
---
`)
	entries := List([]string{root})
	if len(entries) != 2 {
		t.Fatalf("len = %d, want 2", len(entries))
	}
	// Should NOT include nested/ignored because the scan is one-level.
}

func TestListSkipsMissingDir(t *testing.T) {
	entries := List([]string{"/this/path/does/not/exist"})
	if len(entries) != 0 {
		t.Errorf("missing dir should return empty, got %d", len(entries))
	}
}

func TestLoadSkillFallsBackToDirname(t *testing.T) {
	dir := t.TempDir()
	mkdir(t, filepath.Join(dir, "fallback-name"))
	mustWrite(t, filepath.Join(dir, "fallback-name", "SKILL.md"), `---
description: no name field
---
`)
	e, ok := LoadSkill(filepath.Join(dir, "fallback-name", "SKILL.md"))
	if !ok {
		t.Fatal("LoadSkill returned false")
	}
	if e.Name != "fallback-name" {
		t.Errorf("Name = %q, want fallback-name", e.Name)
	}
}

// helpers

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustMkdir(t *testing.T, path string) { mkdir(t, path) }

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
