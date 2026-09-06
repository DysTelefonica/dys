# dys — Guía de uso del CLI

Ejemplos reales para los dos subcomandos que `dys` expone.
Complementa al [`README.md`](../README.md) (qué es) y a
[`docs/ARCHITECTURE.md`](../ARCHITECTURE.md) (cómo está organizado).

## Activación

Cargue este doc cuando:
- Vaya a usar `dys` por primera vez.
- Busque el flag correcto para un caso concreto (listar, filtrar, json).
- Necesite encadenar `dys` con otras herramientas (jq, scripts).

No cargue este doc para:
- Crear o auditar una skill (use la skill `skill-authoring`).
- Diagnosticar el binario (vea `docs/ARCHITECTURE.md`).

## Lo que es

- Una guía de uso del CLI con ejemplos copy-pasteables.
- Un cheat-sheet de flags por subcomando.

## Lo que no es

- Una API reference completa. Para eso use `dys skills list --help`
  o `dys skills tui --help`.
- Un tutorial de Go. Asume familiaridad con bash, git, jq.

## Listar skills

El subcomando por defecto. Escanea `skills/` en el directorio de
trabajo y emite una tabla TSV.

### Listar todo

```bash
$ cd /path/to/team-skills
$ dys skills list
NAME    SCOPE   TIERS   VERSION PATH
skill-propagation project universal,ops 0.2   /…/skill-propagation/SKILL.md
documentation-alan-style project universal,docs 2.0   /…/documentation-alan-style/SKILL.md
vba-binary-drift project vba,runtime 3.1.1  /…/vba-binary-drift/SKILL.md
...
```

### Filtrar por tier

```bash
$ dys skills list --tier vba
NAME    SCOPE   TIERS   VERSION PATH
access-vba-tdd-loop project vba,runtime 2.8.0  /…/access-vba-tdd-loop/SKILL.md
vba-binary-drift project vba,runtime 3.1.1  /…/vba-binary-drift/SKILL.md
...
```

`--tier` acepta:
- **Tiers canónicos**: `universal`, `vba`, `web`, `runtime`.
- **Tiers custom descubiertos del catálogo**: `cadete`, `gentle-ai`,
  `engram`, `dys-tui`, etc. `ScanCustomTiers` los descubre en runtime.

Un tier desconocido retorna cero rows con mensaje claro:

```bash
$ dys skills list --tier nonexistent
No skills match tiers [nonexistent].
```

### Múltiples tiers

```bash
$ dys skills list --tier vba,web
# (intersección OR: skills con cualquier tier de la lista)
```

### Salida JSON

```bash
$ dys skills list --json | jq '.[0]'
{
  "name": "skill-propagation",
  "scope": "project",
  "description": "Refresh de symlinks...",
  "path": "/…/skill-propagation/SKILL.md",
  "author": "Andrés Román",
  "version": "0.2",
  "tiers": ["universal", "ops"]
}
```

Use `--json` con `jq`, `gron`, o cualquier parser JSON. El objeto
siempre tiene 7 claves: `name`, `scope`, `description`, `path`, `author`,
`version`, `tiers`.

### Cambiar el directorio del catálogo

Por defecto, `dys` escanea `./skills/`. Use `--cwd` para apuntar a
otro repositorio:

```bash
$ dys skills list --cwd ~/projects/my-app
# (escanea ~/projects/my-app/skills/)
```

`--cwd` también funciona con paths absolutos. Combine con `--tier` y
`--json` para pipelines de CI:

```bash
$ dys skills list --cwd /var/lib/ci/team-skills --tier runtime --json \
    | jq -r '.[] | .name' \
    | xargs -I{} echo "skill: {}"
```

## Catálogo custom

Combine con grep para filtrar por descripción:

```bash
$ dys skills list --cwd /tmp/opencode/team-skills --json | \
    jq -r '.[] | select(.description | contains("dys")) | .name'
skill-authoring
```

O por autor:

```bash
$ dys skills list --cwd /tmp/opencode/team-skills --json | \
    jq -r '.[] | select(.author == "Andrés Román") | .name'
```

## Tubos comunes

| Caso | Comando |
|---|---|
| Contar skills por tier | `dys skills list --cwd /tmp/opencode/team-skills --json \| jq -r '.[] \| .tiers \| .[]' \| sort \| uniq -c` |
| Listar skills sin tier declarado | `dys skills list --cwd /tmp/opencode/team-skills --json \| jq -r '.[] \| select(.tiers \| not) \| .name'` (siempre vacío: el default `universal` se aplica) |
| Buscar skill por nombre | `dys skills list --cwd /tmp/opencode/team-skills --json \| jq -r '.[] \| select(.name \| test("dys")).name'` |
| Validar frontmatter en CI | `dys skills list --cwd ./skills --json \| jq -e '. \| length > 50'` (assert al menos 50) |

## La skill `skill-authoring`

Para crear o integrar una skill con `dys`, cargue la skill
[`skill-authoring`](../skills/skill-authoring/SKILL.md) en su agente
IA. Esa skill cubre:

- §Núcleo invariantes del frontmatter.
- §Cómo crear una skill nueva en `DysTelefonica/team-skills`.
- §Cómo regenerar la tabla «Auto-invoke» tras cambios.
- §Cómo confirmar la integración con `dys skills list`.

## Nucleo invariantes

- El subcomando `list` no modifica el catálogo ni los runtimes. Es
  solo lectura.
- `--tier` filtra por intersección (OR): una skill con `tiers: [vba, web]`
  matchea `--tier vba` o `--tier web`.
- `--cwd` afecta solo el subcomando actual; no persiste.

## Tabla de puertas

| Condición | Acción |
|---|---|
| Filtra el output del catálogo | `--tier X,Y` o `--tier X` (uno solo). `--json` para machine-readable. |
| Cambia el directorio del catálogo | `--cwd DIR`. Default: `./skills/`. |
| Quieres que la skill aparezca en la tabla «Auto-invoke» de los runtimes | Después de añadir el `metadata.auto_invoke` en el frontmatter, corra `bash bin/build-auto-invoke-table.sh skills` y reinyecte en cada target. La skill `skill-authoring` cubre el flujo. |
| Auditas el catálogo antes de un commit | `bash DysTelefonica/team-skills/testing/suites/frontmatter-validator/validate-frontmatter.sh skills` debe retornar `OK: N SKILL.md files pass all checks`. |

## Antipatrones

| Síntoma | Solución |
|---|---|
| `dys skills list` no retorna nada | Verifique que existe `./skills/` o pase `--cwd DIR` con la raíz del catálogo. |
| `--tier foo` retorna cero rows | Tier desconocido; revise la ortografía o use la salida de la tabla «Auto-invoke» en el AGENTS.md de un runtime. |
| JSON no parsea | `dys skills list --json` siempre emite un array JSON válido. Si no parsea, hay un bug en `dys` (reporte). |
| `dys` crashea con segfault | Reporte con el output de `go build -o dys ./cmd/dys && ./dys skills list --cwd /tmp/opencode/team-skills 2>&1 \| head -5`. |

## Contrato de salida

| Bandera | Tipo | Default | Descripción |
|---|---|---|---|
| `--tier` | CSV de strings | (none) | Filtra por intersección con `metadata.tiers` (case-insensitive). |
| `--json` | bool | `false` | Emite JSON array en lugar de TSV. |
| `--cwd` | string | `.` | Raíz del catálogo (busca `<cwd>/skills/<nombre>/SKILL.md`). |
| `--help`, `-h` | bool | — | Muestra el mensaje de uso. |

## Revisor checklist

- [ ] `dys skills list` retorna al menos las 65 skills esperadas.
- [ ] `dys skills list --tier vba` retorna ~22 skills (todas las vba).
- [ ] `dys skills list --json \| jq '.[0]'` parsea sin error.
- [ ] `dys skills list --tier nonexistent` retorna 0 rows con mensaje claro.
- [ ] `dys` corre con `LD_LIBRARY_PATH` vacío y `ca-certificates` disponible
  (en la imagen `alpine:3.20`).

## Navegación

- [`README.md`](../README.md) — entry point para usuarios.
- [`docs/ARCHITECTURE.md`](../ARCHITECTURE.md) — cómo está organizado.
- [`skills/skill-authoring/SKILL.md`](../skills/skill-authoring/SKILL.md) —
  cómo crear / integrar / propagar una skill.
- [`DysTelefonica/team-skills`](https://github.com/DysTelefonica/team-skills) — el
  catálogo que `dys` consume.
- [`DysTelefonica/team-skills/docs/team-skills-yaml.md`](https://github.com/DysTelefonica/team-skills/blob/main/docs/team-skills-yaml.md) — schema
  del `.team-skills.yaml` (Sprint D).
