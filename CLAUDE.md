# Sheet Ops Public Package

This repository is the canonical public `sheet-ops` package and install source
for the same package topology used in the installed workspace tree.

## Canonical Package Contract

This repository root is the canonical public package.

The installed workspace tree is a materialized copy of that package under
`.codex/skills/sheet-ops`.

Load these first:

1. `README.md`
2. `AGENTS.md`
3. `policies/`
4. `contracts/`
5. `agents/`
6. `skills/`

Canonical package surfaces present here:

- `skills/`
- `agents/`
- `contracts/`
- `runtime/`
- `examples/`
- `internal/testfixtures/`
- `policies/`
- `knowledge/` skeleton and curated public seeds
- `install-skill.sh`
- `scripts/sheet-ops-runtime.sh`

Use `install-skill.sh` as the only install entry:

```bash
/absolute/path/to/this/repository/install-skill.sh --project /absolute/path/to/target/workspace
```

Do not manually copy files into `.codex/`. The installer writes
`.codex/skills/sheet-ops` into the target workspace, materializes the canonical
package there, and builds bundled `sheet-ops-agent` and `sheet-ops-codex`
runtime binaries from this repository.

Mutable local state and development-only reporting trees such as `.codex/`,
`.sheet-ops-state/`, `artifacts/`, `docs/`, and `tests/` are not public
package surfaces. Raw knowledge episode logs such as
`knowledge/**/records.jsonl` are excluded from this package.
