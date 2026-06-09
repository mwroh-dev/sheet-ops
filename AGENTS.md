# Sheet Ops Public Package

This repository is the canonical public Sheet Ops package and install source
for `sheet-ops`, not a separate installer-only tree.

## Canonical Package Contract

This repository root is the canonical public package.

The installed workspace tree is a materialized copy of that package under
`.codex/skills/sheet-ops`.

`agents/` and `skills/` are required package/runtime surfaces, not optional
docs.

## Repository Shape

- `skills/` provides the public skill entry and shared references.
- `agents/` provides open-layer agent prompts, manifests, and contracts.
- `contracts/`, `runtime/`, and `policies/` stay authoritative in this package.
- `examples/` provides public request and operation examples.
- `internal/testfixtures/` provides shared verification helpers.
- `knowledge/` carries namespace skeleton and curated public seeds only.
- `install-skill.sh` is the public install entry.
- `scripts/sheet-ops-runtime.sh` is the source-repo development adapter.
- the installed skill wrapper executes an install-time bundled CLI binary.

Do not treat `.sheet-ops-state/`, `.codex/`, `artifacts/`, `docs/`, or
`tests/` as public package surfaces. Raw knowledge episode logs such as
`knowledge/**/records.jsonl` do not belong in this package.

## Install Rule

When a user points Codex at this repository and asks to install the Sheet Ops
skill, use exactly one install path:

`/absolute/path/to/this/repository/install-skill.sh --project /absolute/path/to/target/workspace`

Do not manually copy files from this repository into `.codex/`. Do not install
from a different repository when this repository is the path the user provided.
The repository path is the installer source; `--project` is the workspace that
should receive `.codex/skills/sheet-ops`.

## Runtime Boundary

After installation, humans invoke the `sheet-ops` Codex skill. The installed
skill calls its bundled `sheet-ops-agent use --envelope` launcher internally.
That launcher is not a second human-facing entry.
