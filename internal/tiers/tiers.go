// Package tiers classifies SKILL.md files into project-type buckets.
//
// Skills declare tiers in their `metadata.tiers:` frontmatter field. Skills
// without tiers default to `universal`. Custom tiers are allowed.
package tiers

import (
	"sort"
	"strings"
)

// WellKnownTiers is the canonical tier list. The first entry is the
// implicit default for skills without metadata.tiers.
var WellKnownTiers = []string{
	"universal",
	"vba",
	"web",
	"runtime",
}

// DefaultTier is assigned to skills whose frontmatter omits metadata.tiers.
const DefaultTier = "universal"

// SkillEntry is the tier-aware projection of a registry SkillEntry.
type SkillEntry struct {
	Name        string
	Path        string
	Description string
	Author      string
	Version     string
	Tiers       []string
}

// ProjectSkillDirs lists the candidate directories the dys CLI scans for
// skills. Order matters: earlier directories win in case of name collision
// (matching gentle-ai / DysTelefonica/team-skills convention).
func ProjectSkillDirs() []string {
	return []string{
		"skills",
		".dysflow/runtime/skills",
		".pi/skills",
	}
}

// Filter returns the subset of entries whose Tiers field intersects the
// wantedTiers list (case-insensitive). Empty wantedTiers returns the input
// unchanged. Skills whose Tiers field is empty are treated as
// []string{DefaultTier}.
func Filter(entries []SkillEntry, wantedTiers []string) []SkillEntry {
	if len(wantedTiers) == 0 || len(entries) == 0 {
		return entries
	}
	wanted := make(map[string]struct{}, len(wantedTiers))
	for _, w := range wantedTiers {
		w = strings.ToLower(strings.TrimSpace(w))
		if w == "" {
			continue
		}
		wanted[w] = struct{}{}
	}
	if len(wanted) == 0 {
		return entries
	}
	out := make([]SkillEntry, 0, len(entries))
	for _, e := range entries {
		tiers := e.Tiers
		if len(tiers) == 0 {
			tiers = []string{DefaultTier}
		}
		for _, t := range tiers {
			if _, ok := wanted[strings.ToLower(strings.TrimSpace(t))]; ok {
				out = append(out, e)
				break
			}
		}
	}
	return out
}

// Group buckets entries by lowercased tier name. A skill with multiple
// tiers appears once per tier.
func Group(entries []SkillEntry) map[string][]SkillEntry {
	out := make(map[string][]SkillEntry, len(WellKnownTiers)+2)
	for _, e := range entries {
		tiers := e.Tiers
		if len(tiers) == 0 {
			tiers = []string{DefaultTier}
		}
		for _, t := range tiers {
			key := strings.ToLower(strings.TrimSpace(t))
			if key == "" {
				continue
			}
			out[key] = append(out[key], e)
		}
	}
	return out
}

// CustomTiers returns sorted tier names present in entries that are not in
// WellKnownTiers.
func CustomTiers(entries []SkillEntry) []string {
	seen := make(map[string]struct{})
	for _, e := range entries {
		for _, t := range e.Tiers {
			if IsWellKnown(t) {
				continue
			}
			key := strings.ToLower(strings.TrimSpace(t))
			if key == "" {
				continue
			}
			seen[key] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for t := range seen {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

// IsWellKnown reports whether a tier name is canonical (case-insensitive).
func IsWellKnown(tier string) bool {
	t := strings.ToLower(strings.TrimSpace(tier))
	for _, w := range WellKnownTiers {
		if strings.EqualFold(w, t) {
			return true
		}
	}
	return false
}

// Normalize trims, dedupes (case-insensitive), and sorts tier strings.
func Normalize(tiers []string) []string {
	if len(tiers) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(tiers))
	out := make([]string, 0, len(tiers))
	for _, t := range tiers {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		key := strings.ToLower(t)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}
