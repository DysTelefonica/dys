//go:build tui

// Command dys-tui is the interactive bubbletea-backed browser for the
// DysTelefonica skill catalog. Build with `go build -tags tui ./cmd/dys-tui`.
//
// It is split out of cmd/dys so that the default deployable
// (`cmd/dys`) does not pull bubbletea into its dependency graph.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DysTelefonica/dys/internal/registry"
	"github.com/DysTelefonica/dys/internal/tui"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "dys-tui: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr *os.File) error {
	if len(args) == 0 {
		return runTUI(".", stdout)
	}
	return runTUI(args[0], stdout)
}

func runTUI(cwd string, stdout *os.File) error {
	roots := []string{cwd}
	entries := registry.List(roots)
	m := tui.New(entries)
	p := tea.NewProgram(m, tea.WithOutput(stdout))
	_, err := p.Run()
	return err
}
