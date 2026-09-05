# dys — DysTelefonica skill-management CLI

`dys` is a Go CLI for managing the DysTelefonica skill catalog. Sprint 1 MVP
ships two subcommands:

```bash
dys skills list [--tier X,Y] [--json] [--cwd DIR]   # tab-separated or JSON
dys skills tui [--cwd DIR]                          # bubbletea two-pane viewer
dys version
```

## Build

```bash
go build ./cmd/dys
./dys version
```

## Project layout

- `cmd/dys/` — main entry point and dispatcher
- `internal/tiers/` — tier taxonomy (`universal`, `vba`, `web`, `runtime`,
  custom), filter / group / classify helpers
- `internal/registry/` — SKILL.md scanner and frontmatter parser
- `internal/tui/` — bubbletea two-pane MVP (Tiers + Skills views)

## Tests

```bash
go test ./...
```

The registry and tiers packages have Go unit tests. The TUI package is
intentionally tested via the `internal/tiers` and `internal/registry`
indirect APIs.

## Origin

Ported from `gentleman-programming/gentle-ai`'s `internal/components/skills/`
into a standalone Go module owned by DysTelefonica, per the
DysTelefonica/team-skills Sprint C refactor.

The full audit that produced this code is in
`DysTelefonica/team-skills` under `docs/BOOTSTRAP-ARCHITECTURE.md` and
the Sprint A..D commits on `main`.
