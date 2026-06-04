# Sheet Ops Release Entry

This checkout is a sanitized mirror of the canonical `sheet-ops` package. It
remains the public install source for the same package topology used in source
and in the installed workspace tree.

## Canonical Package Contract

The source repository root is the canonical package.

The release repository root is a sanitized mirror of the canonical package.

The installed workspace tree is a materialized copy of that package under
`.codex/skills/sheet-ops`.

Load these first:

1. `README.md`
2. `AGENTS.md`
3. `policies/`
4. `contracts/`
5. `agents/`
6. `skills/`

Canonical package surfaces present in this release:

- `skills/`
- `agents/`
- `contracts/`
- `runtime/`
- `examples/`
- `internal/testfixtures/`
- `policies/`
- `knowledge/` skeleton and sanitized seeds
- `install-skill.sh`
- `scripts/sheet-ops-runtime.sh`

Use `install-skill.sh` as the only install entry:

```bash
/absolute/path/to/this/repository/install-skill.sh --project /absolute/path/to/target/workspace
```

Do not manually copy files into `.codex/`. The installer writes
`.codex/skills/sheet-ops` into the target workspace, materializes the canonical
package there, and builds bundled `sheet-ops-agent` and `sheet-ops-codex`
runtime binaries from this checkout.

Mutable local state and development-only reporting trees such as `.codex/`,
`.sheet-ops-state/`, `artifacts/`, `docs/`, `release/`, and `tests/` are not
release package surfaces. Raw knowledge episode logs such as
`knowledge/**/records.jsonl` are excluded from the release mirror.
