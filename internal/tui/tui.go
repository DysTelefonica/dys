// Package tui implements the bubbletea-based Sprint 1 MVP for
// `dys skills tui`. The MVP shows two views: a Tiers view that counts
// skills per tier and lists custom tiers, and a Skills view that renders
// all discovered skills. Switch via Tab; press 1-4 to filter by canonical
// tier; press 'a' or Esc to clear the filter; press 'q' to quit.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/DysTelefonica/dys/internal/registry"
	"github.com/DysTelefonica/dys/internal/tiers"
)

type view int

const (
	viewTiers view = iota
	viewSkills
)

type model struct {
	all     []tiers.SkillEntry
	view    view
	cursor  int
	filter  string
	grouped map[string][]tiers.SkillEntry
	custom  []string
	width   int
	height  int
	quitting bool
}

// New constructs a model from registry entries.
func New(entries []registry.SkillEntry) *model {
	proj := project(entries)
	return &model{
		all:     proj,
		view:    viewTiers,
		grouped: tiers.Group(proj),
		custom:  tiers.CustomTiers(proj),
	}
}

// project converts registry entries to tier entries with the universal
// default applied to entries without tiers.
func project(entries []registry.SkillEntry) []tiers.SkillEntry {
	out := make([]tiers.SkillEntry, len(entries))
	for i, e := range entries {
		out[i] = tiers.SkillEntry{
			Name:        e.Name,
			Path:        e.Path,
			Description: e.Description,
			Author:      e.Author,
			Version:     e.Version,
			Tiers:       e.Tiers,
		}
	}
	return out
}

// Init satisfies tea.Model.
func (m *model) Init() tea.Cmd { return nil }

// Update satisfies tea.Model.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "tab":
			if m.view == viewTiers {
				m.view = viewSkills
			} else {
				m.view = viewTiers
			}
			m.cursor = 0
			return m, nil
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			if m.cursor < m.listLen()-1 {
				m.cursor++
			}
			return m, nil
		case "esc", "a":
			m.filter = ""
			return m, nil
		default:
			if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
				r := msg.Runes[0]
				if r >= '1' && r <= '9' {
					i := int(r - '1')
					if i < len(tiers.WellKnownTiers) {
						m.filter = tiers.WellKnownTiers[i]
						m.view = viewSkills
						m.cursor = 0
						return m, nil
					}
				}
			}
		}
	}
	return m, nil
}

func (m *model) listLen() int {
	switch m.view {
	case viewTiers:
		return len(tiers.WellKnownTiers) + len(m.custom)
	case viewSkills:
		return len(m.filteredSkills())
	}
	return 0
}

func (m *model) filteredSkills() []tiers.SkillEntry {
	if m.filter == "" {
		return m.all
	}
	return tiers.Filter(m.all, []string{m.filter})
}

// View satisfies tea.Model.
func (m *model) View() string {
	if m.quitting {
		return "bye.\n"
	}
	if m.width == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(titleStyle.Render("dys skills tui (Sprint 1 MVP)") + "\n")
	sb.WriteString(tabStyle(m.view == viewTiers, "Tiers") + "  " + tabStyle(m.view == viewSkills, "Skills") + "\n")
	if m.filter != "" {
		sb.WriteString(filterStyle.Render(fmt.Sprintf("filter: tier=%s  (esc / a to clear)", m.filter)) + "\n")
	} else {
		sb.WriteString(filterStyle.Render("filter: none  (press 1-4 for tier, esc to clear)") + "\n")
	}
	sb.WriteString(strings.Repeat("─", minInt(m.width, 80)) + "\n")

	switch m.view {
	case viewTiers:
		sb.WriteString(m.renderTiers())
	case viewSkills:
		sb.WriteString(m.renderSkills())
	}
	return sb.String()
}

func (m *model) renderTiers() string {
	var sb strings.Builder
	for i, t := range tiers.WellKnownTiers {
		count := len(m.grouped[t])
		row := fmt.Sprintf("%s  %3d skill(s)", t, count)
		sb.WriteString(rowStyle(i == m.cursor, false, row) + "\n")
	}
	for i, t := range m.custom {
		count := len(m.grouped[t])
		row := fmt.Sprintf("%s (custom)  %3d skill(s)", t, count)
		sb.WriteString(rowStyle(i+len(tiers.WellKnownTiers) == m.cursor, true, row) + "\n")
	}
	return sb.String()
}

func (m *model) renderSkills() string {
	var sb strings.Builder
	header := fmt.Sprintf("%-32s  %-8s  %-20s  %s", "NAME", "VERSION", "TIERS", "PATH")
	sb.WriteString(headerStyle.Render(header) + "\n")
	rows := m.filteredSkills()
	if len(rows) == 0 {
		sb.WriteString(emptyStyle.Render("(no skills match the current filter)") + "\n")
		return sb.String()
	}
	for i, e := range rows {
		line := fmt.Sprintf("%-32s  %-8s  %-20s  %s",
			truncate(e.Name, 32),
			e.Version,
			strings.Join(e.Tiers, ","),
			e.Path,
		)
		sb.WriteString(rowStyle(i == m.cursor, false, line) + "\n")
	}
	return sb.String()
}

// styles ----------------------------------------------------------------

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			PaddingBottom(1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	tabStyle = func(active bool, name string) string {
		if active {
			return lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#7D56F4")).
				Padding(0, 1).
				Render(name)
		}
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Render(name)
	}

	rowStyle = func(selected, custom bool, text string) string {
		s := lipgloss.NewStyle()
		if selected {
			s = s.Bold(true).Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#3C3C3C"))
		} else if custom {
			s = s.Foreground(lipgloss.Color("#F78000"))
		}
		return s.Render(text)
	}

	filterStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("#888888"))

	emptyStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("#888888"))
)

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}
