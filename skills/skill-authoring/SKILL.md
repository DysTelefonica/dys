---
name: skill-authoring
description: Trigger: creating a new SKILL.md, adding a skill to the catalog, integrating a skill with the dys binary or the team-skills reconciler, propagating a skill to runtime targets. Use when you need to author a new skill, audit an existing one, or wire a new skill into the dys index without breaking the build. Combines skill-style-guide and skill-propagation into one self-contained authoring flow.
metadata:
  author: ardelperal
  version: "1.0"
  last_verified: 2026-09-05
  scope: ['universal', 'dys']
  auto_invoke: ['authoring a new SKILL.md', 'integrating a skill with the dys binary', 'propagating a skill to runtime targets']
  tiers: ['universal', 'dys']
license: Apache-2.0
---

# skill-authoring

Guía única para crear, integrar y propagar una skill en el ecosistema
`DysTelefonica/dys` ↔ `DysTelefonica/team-skills`. Cubre desde el
frontmatter hasta el deploy del binario.

## Activación

Cargue esta skill cuando:
- Vaya a crear una skill nueva (`skills/<nombre>/SKILL.md`).
- Necesite añadir o cambiar los campos `metadata.tiers` o
  `metadata.auto_invoke` de una skill existente.
- Vaya a integrar una skill con el binario `dys` (descubrimiento,
  filtrado, o render TUI).
- Necesite propagar una skill a los runtimes (opencode/codex/claude/pi)
  vía el reconciler de `DysTelefonica/team-skills`.
- Audite una skill existente (frontmatter, secciones canónicas, contrato
  de auto-invoke).

No cargue esta skill para:
- Usar skills existentes (use `dys skills list` o `dys-tui`).
- Diagnosticar el binario `dys` (vea `docs/ARCHITECTURE.md`).

## Lo que es

- Una guía única que combina las convenciones de frontmatter
  (`skill-style-guide` de team-skills) y el contrato de propagación
  (`skill-propagation` de team-skills), específica para el flujo
  `DysTelefonica/dys ↔ team-skills`.
- Un test-driven workflow: cada paso tiene un comando de
  verificación concreto.

## Lo que no es

- Un duplicado de `skill-creator` y `skill-propagation` de team-skills.
  Esta skill los **absorbe** y los aplica específicamente al flujo
  `dys ↔ team-skills`.
- Una guía de cómo usar el binario `dys`. Para eso vea
  `docs/USAGE.md` (Sprint F+).

## Núcleo invariantes

- El frontmatter de la skill lleva `name`, `description`, `license`,
  `metadata.author`, `metadata.version`, `metadata.last_verified`,
  y al menos uno de `metadata.tiers` o `metadata.auto_invoke`.
- El nombre del directorio (`skills/<nombre>/`) coincide con el campo
  `name` del frontmatter, kebab-case.
- Las skills se versionan en `DysTelefonica/team-skills`, NO en `dys`.
  El binario `dys` solo las consume.
- La propagación a los runtimes se hace vía
  `scripts/refresh-personal-symlinks.sh overlays` en team-skills, NO
  desde el binario `dys`.

## Flujo de trabajo (6 pasos)

### Paso 1 — Decidir el tier

Antes de escribir frontmatter, decida a qué tier pertenece. La
taxonomía canónica de `dys` es:

```text
universal  vba  web  runtime
```

Tiers custom descubiertos del catálogo (`cadete`, `engram`, `gentle-ai`,
`ops`, `design`, `infra`, `dys-tui`, etc.) **se descubren en runtime**
vía `ScanCustomTiers` (commit `5b4df85`); no los hardcodee aquí.

| Si la skill aplica a… | Tier |
|---|---|
| Cualquier proyecto (`/dys` es universal) | `universal` |
| El runtime `/dys` específicamente | `dys` |
| Un proyecto VBA o Access | `vba` |
| Un proyecto web | `web` |
| Tooling cross-runtime | `runtime` |

### Paso 2 — Crear la skill en `team-skills`

Cree la skill en `DysTelefonica/team-skills/skills/<nombre>/SKILL.md`:

```yaml
---
name: <nombre>
description: Trigger: <keywords>. <what this skill does>.
metadata:
  author: <github-username>
  version: "0.1"
  last_verified: <YYYY-MM-DD>
  scope: ['<tier1>', '<tier2>']
  auto_invoke: ['<action1>', '<action2>']
license: Apache-2.0
---

# <name>

<opening paragraph: qué hace la skill, ≤ 200 caracteres>

## Activación

Cargue esta skill cuando:
- Vaya a crear una skill nueva en `DysTelefonica/team-skills/skills/<nombre>/`.
- Vaya a actualizar el frontmatter o el cuerpo de una skill existente.
- Necesite bumpear `metadata.version` o cambiar `metadata.tiers` o
  `metadata.auto_invoke` de una skill ya publicada.
- Audite el ciclo de vida completo (crear / actualizar / propagar) de
  una skill del catálogo.

## Lo que es / Lo que no es

## Núcleo invariantes

## Tabla de puertas

## Antipatrones

## Contrato de salida

## Revisor checklist

## Navegación
```

Mantén la regla `documentation-alan-style` §4 (Castellano, sentence
case headings, usted, párrafos <200 chars, no emojis, comillas «»).

### Paso 3 — Validar localmente

Antes de commitear, corra:

```bash
bash DysTelefonica/team-skills/testing/suites/frontmatter-validator/validate-frontmatter.sh skills
```

Salida esperada: `OK: N SKILL.md files pass all checks`. Si falla, lea
cuál campo falta y agregue.

### Paso 4 — Commitear y pushear al repo `team-skills`

```bash
cd DysTelefonica/team-skills
git add skills/<nombre>/
git commit -m "feat(skills): <short description>

<RD-style evidence: symptom + reproduction + acceptance criteria>"
git push origin main
```

### Paso 5 — Regenerar la tabla auto-invoke (en team-skills)

```bash
cd DysTelefonica/team-skills
bash bin/build-auto-invoke-table.sh skills > /tmp/auto-invoke.md
# Para cada target (opencode, codex, claude, pi):
bash bin/inject-auto-invoke.sh /tmp/auto-invoke.md ~/.config/opencode/AGENTS.md
bash bin/inject-auto-invoke.sh /tmp/auto-invoke.md ~/.codex/AGENTS.md
bash bin/inject-auto-invoke.sh /tmp/auto-invoke.md ~/.claude/CLAUDE.md
bash bin/inject-auto-invoke.sh /tmp/auto-invoke.md ~/.pi/agent/AGENTS.md
```

O más simple: corra el reconciler completo, que invoca ambos scripts:

```bash
cd DysTelefonica/team-skills
bash testing/suites/refresh-personal-symlinks/refresh-personal-symlinks.sh overlays
```

### Paso 6 — Confirmar integración con `dys`

Desde el repo `dys`:

```bash
go build -o dys ./cmd/dys
./dys skills list --cwd ../team-skills --tier <tu-tier> | grep <nombre>
```

Salida esperada: tu skill en la lista. Si no aparece, verifique:

- ¿El campo `name` del frontmatter coincide con el directorio?
- ¿`metadata.tiers` o `metadata.scope` incluyen tu tier?
- ¿El reconciler de team-skills corrió sin error?

## Actualizar una skill existente

Cuando cambie el comportamiento, el contrato o el frontmatter de una
skill ya publicada, siga este flujo de 4 pasos. La diferencia clave con
el flujo de creación: la skill ya existe, así que NO recrea el
directorio ni el archivo; solo edita, bumpea version, valida, regenera
auto-invoke.

### Paso U-1 — Decidir qué cambia

| Cambio | Acción |
|---|---|
| Bug en el contrato / output | Bump `metadata.version` (semver patch: 1.0.0 → 1.0.1). |
| Nueva feature / cambio breaking | Bump minor (1.0.x → 1.1.0). |
| Cambio de tier (e.g. añadir `dys`) | Bump patch + actualizar `metadata.tiers`. |
| Cambio en `auto_invoke` phrases | Bump patch + actualizar la lista. |
| Cambio cosmético (typo, formato) | Bump patch solo si cambia el contrato visible. |
| Refactor de la prosa (no funcional) | NO bump; bumpee solo cuando la prosa afecte el contrato. |

### Paso U-2 — Editar el SKILL.md

Edite el archivo existente, sin recrear el directorio:

```bash
cd DysTelefonica/team-skills
$EDITOR skills/<nombre>/SKILL.md
# Edite el frontmatter (cambie `metadata.version`, `metadata.last_verified`,
# `metadata.tiers`, `metadata.auto_invoke` o el cuerpo del skill).
```

Reglas duras:
- `metadata.last_verified` DEBE ser la fecha ISO del día del commit
  (no la fecha del cambio original).
- `metadata.version` DEBE bumpearse en cada cambio visible.
- NO cambie el `name` (es el identificador único; cambiarlo rompe
  discoverability).
- NO cambie la convención `description` (sigue empezando con
  `Trigger: <keywords>.`).

### Paso U-3 — Validar localmente

Igual que para skills nuevas:

```bash
bash testing/suites/frontmatter-validator/validate-frontmatter.sh skills
```

Salida esperada: `OK: N SKILL.md files pass all checks`. Si el SKILL.md
modificado no pasa, lea cuál campo falta y arregle.

### Paso U-4 — Commit, push, regenerar auto-invoke, confirmar

Igual que para skills nuevas, pero el mensaje de commit debe reflejar
que es una actualización:

```bash
cd DysTelefonica/team-skills
git add skills/<nombre>/SKILL.md
git commit -m "fix(<scope>): <short description of the change>

<old behaviour, new behaviour, evidence>"
git push origin main
```

Después:

```bash
cd DysTelefonica/team-skills
bash testing/suites/refresh-personal-symlinks/refresh-personal-symlinks.sh overlays
```

Y desde `dys`:

```bash
go build -o dys ./cmd/dys
./dys skills list --cwd ../team-skills --tier <su-tier> | grep <nombre>
```

Salida esperada: la versión nueva en la columna `VERSION`. Si la
versión no cambió, el push de auto-invoke no se enteró del update.

## Tabla de puertas (update)

| Condición | Acción |
|---|---|
| Cambia el contrato visible de la skill | Bumpea `metadata.version` antes de commitear. |
| Cambia `metadata.tiers` o `metadata.auto_invoke` | Edite, bumpee version, regenere auto-invoke, confirme con `dys`. |
| Cambia el nombre del directorio (`skills/<nombre>/`) | NO LO HAGA. El nombre está acoplado al frontmatter `name` y al `Filter` del binario. En su lugar, agregue un `redirect` (en el `metadata.tiers` del catálogo o en una skill de migration). |
| Cambia `metadata.last_verified` a una fecha pasada | El reconciler no falla pero la tabla auto-invoke lo refleja. Use la fecha del día. |

## Antipatrones (update)

| Síntoma | Solución |
|---|---|
| Update commiteado sin bumpear `metadata.version` | El push de auto-invoke no se enterá del cambio. Use semver discipline. |
| Update commiteado sin actualizar `metadata.last_verified` | El validador puede pasar pero el catálogo se queda con fecha vieja. Use fecha ISO del día. |
| Update commiteado sin regenerar la tabla auto-invoke | Los runtimes muestran el comportamiento viejo hasta que se corra el reconciler. |
| Renombrar el directorio de la skill | Rompe el `Filter` de `dys`. En su lugar cree una skill de migration. |
| Update commiteado sin `go test ./...` | El binario `dys` puede romperse (poco probable pero posible si la skill expone algo al `Filter`). |

## Tabla de puertas

| Condición | Acción |
|---|---|
| Crea una skill nueva | Pasos 1-5 en orden; no commitee sin haber corrido el validador (paso 3). |
| Cambia los tiers de una skill existente | Edite el frontmatter, corra el validador y el reconciler. |
| Integra una skill con el binario `dys` | `dys skills list --cwd ../team-skills --tier <tier> | grep <nombre>`. |
| Propaga una skill a los runtimes | Reconciler en team-skills (`scripts/refresh-personal-symlinks.sh overlays`). |
| Quita una skill | Borre el directorio. No se necesita más acción; el reconciler la quita de la tabla auto-invoke en el siguiente ciclo. |

## Antipatrones

| Síntoma | Solución |
|---|---|
| Frontmatter sin `metadata.last_verified` | El validador falla. Agregue la fecha ISO de hoy. |
| `description` no empieza con `Trigger:` | El validador falla loud. Prefije con `Trigger: <keywords>.` y luego la prosa. |
| Tier custom inventado (`mid-tier`, `temp`) | Mejor declare `universal` o use el `auto_invoke` para refinar. Los custom se descubren; inventar tiers sin skills que los usen no aporta. |
| Skill versionada en `dys` | Las skills viven en `team-skills`. `dys` las lee, no las contiene. |
| No regeneré la tabla auto-invoke tras añadir la skill | La skill aparecerá en `dys skills list` pero NO en la tabla «Auto-invoke» de los runtimes. Corra el paso 5. |
| Headers en Title Case en skills Castellanas | skill-style-guide §4.4: sentence case en ES, Title Case solo en EN. |

## Contrato de salida

### Crear

| Artefacto | Ubicación | Formato |
|---|---|---|
| Skill en el catálogo | `DysTelefonica/team-skills/skills/<nombre>/SKILL.md` | Frontmatter YAML + Markdown. |
| Tabla «Auto-invoke» | `~/.config/opencode/AGENTS.md` (y 3 targets más) | Markdown con header `<!-- auto-invoke-table -->`. |
| Index del catálogo | `DysTelefonica/team-skills/.atl/skill-registry.md` | Markdown con tabla de skills. |
| Binario `dys` | `DysTelefonica/dys` | `dys skills list --tier <tier>`. |

### Actualizar

| Artefacto | Cambio esperado | Comando |
|---|---|---|
| `SKILL.md` (frontmatter) | `metadata.version` bumpeada, `metadata.last_verified` actualizado, `metadata.tiers` / `metadata.auto_invoke` ajustados | Editor + `validate-frontmatter.sh`. |
| Tabla «Auto-invoke» | Recalculada desde el frontmatter nuevo | `bash refresh-personal-symlinks.sh overlays`. |
| `CHANGELOG.md` (team-skills) | Entrada bajo `[Unreleased]` con el delta | Editor + `git commit`. |
| Binario `dys` | Compila con el binario deployable actualizado | `go build ./cmd/dys`. |

## Revisor checklist

### Crear

- [ ] `bash testing/suites/frontmatter-validator/validate-frontmatter.sh skills` pasa.
- [ ] Las suites de team-skills pasan: `for s in testing/suites/*/test-*.sh; do bash "$s"; done`.
- [ ] `dys skills list --cwd ../team-skills --tier <tier> | grep <nombre>` retorna la skill.
- [ ] El `CHANGELOG.md` de team-skills menciona la nueva skill bajo `[Unreleased]`.
- [ ] El reconciler de team-skills emite `PHASE_DONE` por fase.
- [ ] La tabla «Auto-invoke» de cada target incluye la skill en una fila.

### Actualizar

- [ ] `metadata.version` bumpeada según semver (patch / minor / major).
- [ ] `metadata.last_verified` actualizada a la fecha ISO del día.
- [ ] El validador de frontmatter pasa después del cambio.
- [ ] El reconciler de team-skills regeneró la tabla «Auto-invoke».
- [ ] `dys skills list` muestra la nueva versión de la skill.
- [ ] El `CHANGELOG.md` menciona el cambio bajo `[Unreleased]`.

## Navegación

- `DysTelefonica/team-skills/skills/documentation-alan-style/SKILL.md` — convenciones
  de prosa para todas las skills.
- `DysTelefonica/team-skills/skills/skill-propagation/SKILL.md` — política
  de propagación (esta skill la absorbe).
- `DysTelefonica/team-skills/AGENTS.md` — índice del catálogo.
- `DysTelefonica/team-skills/docs/team-skills-yaml.md` — schema del fichero
  `.team-skills.yaml` (Sprint D).
- `DysTelefonica/dys/README.md` — entry point del binario.
- `DysTelefonica/dys/docs/ARCHITECTURE.md` — arquitectura del binario.
- `DysTelefonica/dys/docs/USAGE.md` — ejemplos de uso del CLI.
