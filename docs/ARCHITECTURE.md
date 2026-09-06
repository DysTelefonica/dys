# Arquitectura de `dys`

## Vista general

```text
DysTelefonica/dys/
├── cmd/
│   ├── dys/                  # binario deployable: skills list + (skills tui fallback)
│   └── dys-tui/              # binario dev-only con bubbletea (//go:build tui)
├── internal/
│   ├── registry/             # scanner de SKILL.md + parser de frontmatter
│   ├── tiers/                # taxonomía universal/vba/web/runtime + custom discovery
│   └── tui/                  # bubbletea two-pane viewer (gated por //go:build tui)
├── go.mod
├── go.sum
├── Dockerfile               # multi-stage: golang:1.24-alpine → alpine:3.20
├── README.md
└── LICENSE
```

## Responsabilidades por directorio

| Directorio | Responsabilidad | Leído por | Escrito por |
|---|---|---|---|
| `cmd/dys/` | Entry point del binario deployable. Despacha subcomandos `skills list`, `skills tui` (fallback) y `version`. | Operadores, scripts, coolify-mcp deploy. | Mantenedores de CLI. |
| `cmd/dys-tui/` | Entry point del binario interactivo (con bubbletea). Build con `-tags tui`. | Developers locales. | Mantenedores de TUI. |
| `internal/registry/` | Scanner de SKILL.md (un nivel de profundidad) y parser de frontmatter YAML. Fuente única de datos para `Filter` y `ScanCustomTiers`. | `cmd/dys/`, `cmd/dys-tui/`. | Mantenedores de CLI. |
| `internal/tiers/` | Taxonomía `WellKnownTiers` + `Filter` + `Group` + `Normalize` + `IsWellKnown` + `CustomTiers` (estáticos). | `cmd/dys/`, `cmd/dys-tui/`. | Mantenedores de CLI. |
| `internal/tui/` | Vista bubbletea two-pane (Tiers + Skills). Detrás de `//go:build tui`. | `cmd/dys-tui/`. | Mantenedores de TUI. |
| `Dockerfile` | Multi-stage build para Coolify u otros orquestadores. `golang:1.24-alpine` (build) → `alpine:3.20` (runtime). | Coolify-mcp deploy. | Mantenedores de release. |
| `README.md` | Punto de entrada para developers y operadores. Castellano alan-style §4-§5. | Developers, operadores, IAs. | Mantenedores de docs. |
| `go.mod`, `go.sum` | Módulo Go `github.com/DysTelefonica/dys`. | Toda la toolchain. | Mantenedores de release. |

## Cómo se extiende cada responsabilidad

- `cmd/dys/`: añada un nuevo subcomando en `run(args, stdout, stderr)`; cree la función `runNUEVO` y wire en el switch de `run`.
- `cmd/dys-tui/`: añada una vista nueva en el `model` de bubbletea; respetela con `//go:build tui` para que no contamine el binario deployable.
- `internal/registry/`: añada un parser para campos de frontmatter nuevos (escriba el test antes, TDD discipline).
- `internal/tiers/`: añada un tier canónico solo si la mayoría de skills lo declaran explícitamente. Los tiers custom se descubren en runtime vía `ScanCustomTiers`.
- `internal/tui/`: añada vistas y filtros; respete el build tag `tui` para mantener el binario deployable libre de bubbletea.
- `Dockerfile`: bumpee la imagen base `golang:X.Y-alpine` solo cuando el toolchain arm64 de Coolify ofrezca X.Y. Hoy: 1.24.
- `README.md`: actualice la sección Núcleo invariantes o Tabla de puertas cuando cambie el contrato.

## Flujo de descubrimiento de skills

```text
operator
  │
  ▼ dys skills list --cwd <repo>
  │
  ├── List(roots)                      [internal/registry]
  │     │
  │     ▼ findSkillFiles(roots)        # one level deep, follows symlinks
  │     │
  │     ▼ ParseFrontmatter(src)        # inline + block scalars, metadata.tiers
  │
  ├── ScanCustomTiers(entries, canonical)
  │     │
  │     ▼ merge with WellKnownTiers → effective tier universe
  │
  └── Filter(entries, userTiers ∪ discovered)
        │
        ▼ Tier intersection match (case-insensitive)
        │
        ▼ emit TSV or JSON to stdout
```

## Núcleo invariantes

- El binario deployable `dys` NO depende de `bubbletea` ni de `lipgloss`. La build tag `tui` excluye el package `internal/tui` del grafo de dependencias de `cmd/dys`.
- `--tier X` con un tier desconocido retorna cero matches. Un tier desconocido no se expande silenciosamente a «todos los tiers del catálogo».
- `dys` lee el catálogo con la misma profundidad (un nivel) que el reconciler `refresh-personal-symlinks.sh` de team-skills. No indexa sub-skills de `vendored/` ni archivos sueltos.
- El parser de frontmatter lee `metadata.tiers` (canónico) y, cuando está presente, también acepta `metadata.scope` (legacy). Precedencia: `tiers` sobre `scope`.
- El binario `dys` no expone HTTP. El `ENTRYPOINT` del Dockerfile ejecuta `dys version` y luego `tail -f /dev/null` para mantener el contenedor vivo sin healthcheck HTTP.

## Tabla de puertas

| Condición | Acción |
|---|---|
| Cambia el parser de frontmatter | Edite `internal/registry/registry.go` + `registry_test.go`. TDD: test antes de la implementación. |
| Cambia la taxonomía de tiers | Edite `internal/tiers/tiers.go` (`WellKnownTiers`). Los tiers custom se descubren en runtime; no los hardcodee aquí. |
| Añade un nuevo subcomando | Edite `cmd/dys/main.go` `run(args, stdout, stderr)`. Mantenga el switch de dispatch y añada un test en `cmd/dys/main_test.go` si la lógica es no trivial. |
| Bump de Go | Edite `go.mod` (`go 1.24`), `Dockerfile` (`golang:1.24-alpine`), `nixpacks.toml` si existe. Verifique arm64 en Coolify antes de bumpear. |
| Compila el binario deployable | `go build ./cmd/dys` y verifique que `dys version` retorna exit 0 sin `bubbletea` en la salida de `go mod why`. |

## Antipatrones

| Síntoma | Solución |
|---|---|
| `dys` falla con `error: no such tool "tui"` al ejecutar `dys skills tui` | El binario deployable no incluye TUI. Compila `cmd/dys-tui` con `--tags tui` para builds con TUI. |
| `dys skills list --tier foo` retorna todas las skills | No debería pasar; verifique que la flag `--tier` se pase a `Filter` sin union con todos los custom tiers. |
| `go build ./cmd/dys` incluye `bubbletea` en la salida de `go mod why` | El package `internal/tui` debe tener `//go:build tui` y la dispatch de `runSkillsTUI` debe estar en un archivo separado con el mismo tag. |
| `dys` retorna 0 skills para un tier custom del catálogo | Verifique que el tier esté en `metadata.tiers` de al menos un SKILL.md. `ScanCustomTiers` lo descubre en runtime. |
| `dys skills tui` muestra solo tiers canónicos | El TUI lee `Group(entries)` que incluye custom tiers. Refresque el catálogo o reinicie `dys-tui`. |

## Contrato de salida

| Subcomando | Formato | Salida |
|---|---|---|
| `dys skills list` | TSV | Cabecera `NAME\tSCOPE\tTIERS\tVERSION\tPATH` + filas por skill que matchea `--tier` (o todas si se omite). |
| `dys skills list --json` | JSON | Array de objetos con `name`, `scope`, `description`, `path`, `author`, `version`, `tiers`. |
| `dys skills tui` | bubbletea | UI two-pane (Tiers + Skills) navegable con Tab, 1-4, q. |
| `dys version` | texto plano | `dys dev (Sprint 1 MVP)` o el SHA de GoReleaser. |
| `dys` (sin args) | texto plano | Mensaje de uso. Exit 0. |

## Revisor checklist

- [ ] `go build ./cmd/dys` produce un binario estático sin `bubbletea` en `go mod why`.
- [ ] `go test ./...` pasa localmente.
- [ ] `dys skills list --tier <tier-conocido>` retorna las skills esperadas.
- [ ] `dys skills list --tier <tier-inexistente>` retorna cero rows con mensaje claro.
- [ ] `Dockerfile` produce una imagen `alpine:3.20` con `/usr/local/bin/dys` ejecutable.
- [ ] El binario `dys` deployado responde a `dys version` (verificado vía logs de Coolify).

## Navegación

- [`README.md`](../README.md) — punto de entrada para usuarios.
- [`Dockerfile`](../Dockerfile) — multi-stage build.
- [`DysTelefonica/team-skills`](https://github.com/DysTelefonica/team-skills) — catálogo
  de skills que `dys` consume.
- [`DysTelefonica/team-skills/docs/BOOTSTRAP-ARCHITECTURE.md`](https://github.com/DysTelefonica/team-skills/blob/main/docs/BOOTSTRAP-ARCHITECTURE.md) — arquitectura del
  catálogo (este doc es su gemelo en miniatura).
- [`DysTelefonica/team-skills/docs/team-skills-yaml.md`](https://github.com/DysTelefonica/team-skills/blob/main/docs/team-skills-yaml.md) — schema
  del fichero `.team-skills.yaml` (Sprint D).
