# dys — CLI de gestión del catálogo de skills de DysTelefonica

Binario Go portable para listar, filtrar y (opcionalmente) navegar
el catálogo de skills de `DysTelefonica/team-skills`. Implementa el
subconjunto CLI del runtime; el binario hermano `dys-tui` (con
`bubbletea`) se compila aparte cuando se necesita la UI interactiva.

## Activación

Cargue este README cuando:
- Arranque a usar `dys` por primera vez.
- Busque la flag correcta para un subcomando (lista, json, tui).
- Necesite entender por qué `dys` se compila en dos binarios separados.

No cargue este README para:
- Entender el catálogo de skills (eso vive en `DysTelefonica/team-skills/README.md`).
- Aprender las convenciones de frontmatter de skills (eso vive en
  `DysTelefonica/team-skills/skills/skill-style-guide/SKILL.md`).

## Lo que es

- Un binario Go estático y portable, sin dependencias en tiempo de
  ejecución más allá de `ca-certificates` (la imagen `alpine:3.20`).
- Una CLI de dos subcomandos (`dys skills list`, `dys skills tui`),
  con descubrimiento dinámico de tiers custom desde el frontmatter del
  catálogo.
- Una alternativa portable al script `bin/list-skills.py` de team-skills,
  apta para deploy en contenedores pequeños y para entornos sin Python.

## Lo que no es

- Un sub-módulo de `Gentleman-Programming/gentle-ai`. Este repositorio
  es un fork con marca propia (`dys`), no una copia literal.
- Una réplica del catálogo de skills. El catálogo vive en
  `DysTelefonica/team-skills`; este binario solo lo consume.
- Un binario monolítico. `dys` (sin bubbletea) y `dys-tui` (con
  bubbletea) son dos productos distintos. `dys` es el deployable;
  `dys-tui` es solo para developers.

## Núcleo invariantes

- El binario deployable `dys` NO depende de `bubbletea` ni de `lipgloss`.
  El build tag `tui` excluye el package `internal/tui` del grafo de
  dependencias de `cmd/dys`.
- El parser de frontmatter lee el campo `metadata.tiers` (canónico) y,
  cuando está presente, también acepta `metadata.scope` (legacy, 2
  skills del catálogo lo usan). La precedencia es `tiers` sobre
  `scope` cuando ambos están presentes.
- `--tier X` con un tier desconocido retorna cero matches. Un tier
  desconocido no se expande silenciosamente a «todos los tiers del
  catálogo» (eso sería un foot-gun).
- `dys` lee el catálogo con la misma profundidad (un nivel) que el
  reconciler `refresh-personal-symlinks.sh` de team-skills. No
  indexa sub-skills de `vendored/` ni archivos sueltos.

## Comandos

```bash
dys skills list [--tier X,Y] [--json] [--cwd DIR]
dys skills tui  [--cwd DIR]
dys version
```

| Subcomando | Flags | Salida |
|---|---|---|
| `skills list` | `--tier` filtra por tier (canónico o custom). `--json` emite JSON. `--cwd` cambia la raíz del catálogo (default: cwd). | TSV por defecto; columnas: `NAME`, `SCOPE`, `TIERS`, `VERSION`, `PATH`. |
| `skills tui` | `--cwd` cambia la raíz del catálogo. Requiere binario con `//go:build tui`. | UI bubbletea de dos paneles (Tiers + Skills). |
| `version` | — | `dys dev (Sprint 1 MVP)` o el SHA de GoReleaser. |

## Build

| Target | Comando | Binario | Notas |
|---|---|---|---|
| CLI deployable (default) | `go build ./cmd/dys` | `dys` | Sin `bubbletea`; imagen `alpine:3.20` apta. |
| TUI interactivo (dev) | `go build -tags tui ./cmd/dys-tui` | `dys-tui` | Requiere `bubbletea` y `lipgloss`. Solo para uso local. |
| Tests | `go test ./...` | — | Cubre `internal/registry` y `internal/tiers`. |

El repo incluye un `Dockerfile` multi-stage que produce `dys` (sin
`bubbletea`) listo para deploy en Coolify u otro orquestador.

## Tabla de puertas

| Condición | Acción |
|---|---|
| Cambia una skill del catálogo | Edita el frontmatter, corre `bin/build-auto-invoke-table.sh` y `bin/inject-auto-invoke.sh` para refrescar el AGENTS.md de cada runtime. |
| Cambia el parser de frontmatter | Edita `internal/registry/registry.go` y `internal/registry/registry_test.go`. TDD: añadir test antes de la implementación. |
| Cambia la taxonomía de tiers | Edita `internal/tiers/tiers.go` (`WellKnownTiers`). Los tiers custom se descubren en runtime. |
| Compila el binario deployable | `go build ./cmd/dys` y verifica que `dys` corra sin el package `internal/tui` importado. |

## Antipatrones

| Síntoma | Solución |
|---|---|
| `dys` falla con `error: no such tool "tui"` al ejecutar `dys skills tui` | El binario deployable no incluye el TUI. Compila `cmd/dys-tui` con `--tags tui` para builds con TUI. |
| `dys skills list --tier foo` retorna todas las skills | No debería pasar; verifica que la flag `--tier` se pase a `Filter` sin union con todos los custom tiers. |
| `go build` con `bubbletea` falla en CI | El `dys` deployable no debe importar `bubbletea`. El package `internal/tui` tiene `//go:build tui`. |
| `dys` retorna 0 skills para un tier custom del catálogo | Verifica que el tier esté en `metadata.tiers` de al menos un SKILL.md; la flag `--tier` solo matchea contra tiers presentes en el catálogo. |

## Contrato de salida

| Subcomando | Clave | Tipo | Descripción |
|---|---|---|---|
| `skills list` | `NAME` | string | Nombre kebab-case de la skill. |
| `skills list` | `SCOPE` | string | `project` (ruta bajo `--cwd`) o `user` (ruta fuera). |
| `skills list` | `TIERS` | string CSV | Lista de tiers separados por coma. Default: `universal` si la skill no declara. |
| `skills list` | `VERSION` | string | Versión del frontmatter `metadata.version`. |
| `skills list` | `PATH` | string | Ruta absoluta del SKILL.md. |
| `skills list --json` | `name`, `scope`, `description`, `path`, `author`, `version`, `tiers` | objeto JSON | Equivalente al TSV más `description` completa. |

## Revisor checklist

- [ ] `go build ./cmd/dys` produce un binario estático sin `bubbletea` en
  la lista de dependencias (`go mod why github.com/charmbracelet/bubbletea`
  retorna «(unused)»).
- [ ] `go test ./...` pasa localmente.
- [ ] `dys skills list --tier <tier-conocido>` retorna las skills esperadas.
- [ ] `dys skills list --tier <tier-inexistente>` retorna cero rows con
  mensaje claro.
- [ ] El `Dockerfile` produce una imagen `alpine:3.20` que contiene
  `/usr/local/bin/dys`.
- [ ] El `CHANGELOG.md` del repo `DysTelefonica/team-skills` menciona
  este binario si cambia el contrato CLI.

## Navegación

- `DysTelefonica/team-skills` — catálogo de skills y reconciler
  transaccional.
- `DysTelefonica/team-skills/skills/documentation-alan-style/SKILL.md` —
  convenciones de estilo (este README las aplica).
- `DysTelefonica/team-skills/docs/BOOTSTRAP-ARCHITECTURE.md` — cómo se
  distribuye el catálogo a runtimes.
- `DysTelefonica/team-skills/skills/skill-style-guide/SKILL.md` — contrato
  de frontmatter que `dys` consume.
- `cmd/dys/main.go` — dispatcher y flags.
- `internal/registry/registry.go` — parser de frontmatter.
- `internal/tiers/tiers.go` — taxonomía de tiers.
- `internal/tui/tui.go` — UI bubbletea (solo `//go:build tui`).
- `CHANGELOG.md` — cambios versionados de este binario.
