package tiers

import (
	"reflect"
	"strings"
	"testing"
)

func mk(name string, tiers []string) SkillEntry {
	return SkillEntry{
		Name:  name,
		Path:  "/skills/" + name + "/SKILL.md",
		Tiers: tiers,
	}
}

func TestFilterCaseInsensitive(t *testing.T) {
	all := []SkillEntry{
		mk("a", []string{"VBA"}),
		mk("b", []string{"web"}),
	}
	got := Filter(all, []string{"vba"})
	if len(got) != 1 || got[0].Name != "a" {
		t.Fatalf("expected only a, got %v", got)
	}
}

func TestFilterDefaultsToUniversal(t *testing.T) {
	all := []SkillEntry{
		mk("legacy", nil),
		mk("explicit", []string{"universal"}),
	}
	got := Filter(all, []string{"universal"})
	if len(got) != 2 {
		t.Fatalf("expected 2 (legacy defaults to universal), got %d", len(got))
	}
}

func TestFilterEmptyTiersReturnsAll(t *testing.T) {
	all := []SkillEntry{mk("a", []string{"x"}), mk("b", []string{"y"})}
	got := Filter(all, nil)
	if len(got) != 2 {
		t.Errorf("empty filter should return all, got %d", len(got))
	}
}

func TestGroupMultiTier(t *testing.T) {
	all := []SkillEntry{
		mk("a", []string{"x", "y"}),
		mk("b", []string{"y"}),
		mk("c", nil),
	}
	got := Group(all)
	if len(got["x"]) != 1 || got["x"][0].Name != "a" {
		t.Errorf("x group wrong: %v", got["x"])
	}
	if len(got["y"]) != 2 {
		t.Errorf("y group should have 2 entries, got %d", len(got["y"]))
	}
	if len(got["universal"]) != 1 || got["universal"][0].Name != "c" {
		t.Errorf("c should fall into universal, got %v", got["universal"])
	}
}

func TestCustomTiers(t *testing.T) {
	all := []SkillEntry{
		mk("a", []string{"experimental"}),
		mk("b", []string{"vba", "experimental"}),
		mk("c", []string{"vba"}),
	}
	got := CustomTiers(all)
	if len(got) != 1 || got[0] != "experimental" {
		t.Errorf("CustomTiers = %v, want [experimental]", got)
	}
}

func TestIsWellKnownCaseInsensitive(t *testing.T) {
	for _, w := range WellKnownTiers {
		if !IsWellKnown(w) {
			t.Errorf("IsWellKnown(%q) = false", w)
		}
		if !IsWellKnown(strings.ToUpper(w)) {
			t.Errorf("IsWellKnown(%q) = false (case-insensitive)", strings.ToUpper(w))
		}
	}
	if IsWellKnown("not-a-tier") {
		t.Error("IsWellKnown(\"not-a-tier\") = true, want false")
	}
}

func TestNormalizeDedupesAndSorts(t *testing.T) {
	got := Normalize([]string{"z", "a", "A", "b", "", " a "})
	want := []string{"a", "b", "z"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Normalize = %v, want %v", got, want)
	}
}
