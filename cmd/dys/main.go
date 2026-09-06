// Command dys is the DysTelefonica skill-management CLI. Sprint 1 MVP
// ships `dys skills list` (with optional --tier filter and --json output)
// in the default build. The interactive `dys skills tui` view is gated
// behind the `tui` build tag and pulls in github.com/charmbracelet/bubbletea;
// the default deployable build excludes it so the binary has no
// bubbletea / lipgloss dependency and compiles on every Go toolchain.
//
// Build the TUI locally with `go build -tags tui ./cmd/dys`.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DysTelefonica/dys/internal/registry"
	"github.com/DysTelefonica/dys/internal/tiers"
)

const version = "dev (Sprint 1 MVP)"

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "dys: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr *os.File) error {
	if len(args) == 0 {
		return usage(stdout)
	}
	switch args[0] {
	case "--help", "-h", "help":
		return usage(stdout)
	case "version", "--version", "-v":
		fmt.Fprintf(stdout, "dys %s\n", version)
		return nil
	case "skills":
		return runSkills(args[1:], stdout, stderr)
	default:
		return fmt.Errorf("unknown command %q (try `dys help`)", args[0])
	}
}

func usage(w *os.File) error {
	fmt.Fprintln(w, `Usage: dys <command> [flags]

Commands:
  dys skills list [--tier X,Y] [--json] [--cwd DIR]
  dys skills tui [--cwd DIR]
  dys version

dys is the DysTelefonica skill-management CLI. Sprint 1 MVP.`)
	return nil
}

func runSkills(args []string, stdout, stderr *os.File) error {
	if len(args) == 0 {
		return skillsUsage(stdout)
	}
	if args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		return skillsUsage(stdout)
	}
	switch args[0] {
	case "list":
		return runSkillsList(args[1:], stdout, stderr)
	case "tui":
		return runSkillsTUI(args[1:], stdout, stderr)
	default:
		return fmt.Errorf("unknown skills command %q (want list or tui)", args[0])
	}
}

func skillsUsage(w *os.File) error {
	fmt.Fprintf(w, `Usage: dys skills <list|tui> [flags]

Commands:
  dys skills list [--tier X,Y] [--json] [--cwd DIR]
  dys skills tui [--cwd DIR]
`)
	return nil
}

func runSkillsList(args []string, stdout, stderr *os.File) error {
	cwd, opts, err := parseSkillsListArgs(args)
	if err != nil {
		return err
	}
	if opts.help {
		fmt.Fprintln(stdout, `Usage: dys skills list [--tier X,Y] [--json] [--cwd DIR]

Lists every SKILL.md found under skills/ in --cwd (default: cwd).
With --tier, filters by metadata.tiers intersection. Comma-separated
list supported. With --json, emits a JSON document instead of TSV.`)
		return nil
	}

	roots, err := resolveSkillRoots(cwd)
	if err != nil {
		return err
	}
	entries := registry.List(roots)
	proj := project(entries)
	var filtered []tiers.SkillEntry
	if len(opts.tiers) > 0 {
		filtered = tiers.Filter(proj, opts.tiers)
	} else {
		filtered = proj
	}

	if opts.json {
		return emitJSON(stdout, filtered, cwd)
	}

	if len(filtered) == 0 {
		if len(opts.tiers) > 0 {
			fmt.Fprintf(stdout, "No skills match tiers %v.\n", opts.tiers)
		} else {
			fmt.Fprintln(stdout, "No skills found.")
		}
		return nil
	}

	fmt.Fprintln(stdout, "NAME\tSCOPE\tTIERS\tVERSION\tPATH")
	for _, e := range filtered {
		scope := scopeForPath(cwd, e.Path)
		fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\t%s\n",
			e.Name,
			scope,
			strings.Join(e.Tiers, ","),
			e.Version,
			e.Path,
		)
	}
	return nil
}

// runSkillsTUI handles `dys skills tui` in the default (no-bubbletea)
// build. It prints a hint and falls back to the text listing so the
// operator still gets a usable view. Build with `-tags tui` for the real
// bubbletea-backed browser (it lives in cmd/dys-tui as a separate binary
// to keep bubbletea out of this package's dependency graph).
func runSkillsTUI(args []string, stdout, stderr *os.File) error {
	fmt.Fprintln(stdout, "dys skills tui: this binary was built without the bubbletea-backed browser.")
	fmt.Fprintln(stdout, "Falling back to `dys skills list --json` for a machine-readable view.")
	fmt.Fprintln(stdout, "To get the interactive browser, install `dys-tui` (cmd/dys-tui, built with -tags tui).")
	return runSkillsList(args, stdout, stderr)
}

type listOpts struct {
	tiers []string
	json  bool
	help  bool
}

func parseSkillsListArgs(args []string) (string, *listOpts, error) {
	cwd := ""
	opts := &listOpts{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--help", "-h":
			opts.help = true
		case "--cwd":
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("--cwd requires a value")
			}
			cwd = args[i+1]
			i++
		case "--json":
			opts.json = true
		case "--tier":
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("--tier requires a value (comma-separated)")
			}
			opts.tiers = append(opts.tiers, strings.Split(args[i+1], ",")...)
			i++
		default:
			return "", nil, fmt.Errorf("unknown argument %q", args[i])
		}
	}
	return cwd, opts, nil
}

func resolveSkillRoots(cwd string) ([]string, error) {
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	out := make([]string, 0, len(tiers.ProjectSkillDirs()))
	for _, r := range tiers.ProjectSkillDirs() {
		out = append(out, filepath.Join(cwd, r))
	}
	return out, nil
}

func scopeForPath(cwd, path string) string {
	if cwd != "" && strings.HasPrefix(path, cwd) {
		return "project"
	}
	return "user"
}

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

type jsonRow struct {
	Name        string   `json:"name"`
	Scope       string   `json:"scope"`
	Description string   `json:"description"`
	Path        string   `json:"path"`
	Author      string   `json:"author"`
	Version     string   `json:"version"`
	Tiers       []string `json:"tiers"`
}

func emitJSON(w *os.File, entries []tiers.SkillEntry, cwd string) error {
	rows := make([]jsonRow, len(entries))
	for i, e := range entries {
		rows[i] = jsonRow{
			Name:        e.Name,
			Scope:       scopeForPath(cwd, e.Path),
			Description: e.Description,
			Path:        e.Path,
			Author:      e.Author,
			Version:     e.Version,
			Tiers:       e.Tiers,
		}
	}
	data, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}
