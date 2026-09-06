package registry

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestScanCustomTiersMergesWithCanonical(t *testing.T) {
	root := t.TempDir()
	// Three skills: one with a custom tier, one with a known tier, one
	// with both. The scan must surface every unique tier.
	mk := func(name string, body string) {
		if err := os.MkdirAll(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, name, "SKILL.md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("alpha", "---\nname: alpha\ndescription: Trigger: x\nlicense: Apache-2.0\nmetadata:\n  author: a\n  version: \"1.0\"\n  last_verified: 2026-09-05\n  tiers: ['cadete']\n---\n")
	mk("beta", "---\nname: beta\ndescription: Trigger: x\nlicense: Apache-2.0\nmetadata:\n  author: a\n  version: \"1.0\"\n  last_verified: 2026-09-05\n  tiers: [vba]\n---\n")
	mk("gamma", "---\nname: gamma\ndescription: Trigger: x\nlicense: Apache-2.0\nmetadata:\n  author: a\n  version: \"1.0\"\n  last_verified: 2026-09-05\n  tiers: [cadete, vba]\n---\n")

	entries := List([]string{root})
	if len(entries) != 3 {
		t.Fatalf("List returned %d entries, want 3", len(entries))
	}

	got := ScanCustomTiers(entries, []string{"universal", "vba", "web", "runtime"})
	want := []string{"cadete"}
	if !equalStringSlice(got, want) {
		t.Fatalf("ScanCustomTiers = %v, want %v", got, want)
	}
}

func TestScanCustomTiersEmptyCatalog(t *testing.T) {
	got := ScanCustomTiers(nil, []string{"universal", "vba"})
	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %v", got)
	}
}

func TestScanCustomTiersDeduplicates(t *testing.T) {
	root := t.TempDir()
	mk := func(name string, body string) {
		if err := os.MkdirAll(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, name, "SKILL.md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("a", "---\nname: a\ndescription: Trigger: x\nlicense: Apache-2.0\nmetadata:\n  author: a\n  version: \"1.0\"\n  last_verified: 2026-09-05\n  tiers: [dys-tui, vba]\n---\n")
	mk("b", "---\nname: b\ndescription: Trigger: x\nlicense: Apache-2.0\nmetadata:\n  author: a\n  version: \"1.0\"\n  last_verified: 2026-09-05\n  tiers: [dys-tui, engram]\n---\n")

	entries := List([]string{root})
	got := ScanCustomTiers(entries, []string{"universal", "vba", "web", "runtime"})
	want := []string{"dys-tui", "engram"}
	if !equalStringSlice(got, want) {
		t.Fatalf("dedup = %v, want %v", got, want)
	}
}

func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Helper used by the production code; the tests in tiers_test.go
// already exercise Filter / Group / CustomTiers. The scan function
// itself is what this file adds; the tier-merge step lives in tiers.go.

func mergeTiersForTest(discovered, canonical []string) []string {
	known := make(map[string]struct{}, len(canonical))
	for _, t := range canonical {
		known[strings.ToLower(strings.TrimSpace(t))] = struct{}{}
	}
	out := make([]string, 0)
	seen := make(map[string]struct{})
	for _, t := range discovered {
		key := strings.ToLower(strings.TrimSpace(t))
		if key == "" {
			continue
		}
		if _, ok := known[key]; ok {
			continue
		}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

var _ = mergeTiersForTest // keep the symbol referenced
