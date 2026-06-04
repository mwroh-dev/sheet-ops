# Install

This repository is currently a source-based local harness. To exercise the full
public flow, install the project-local Codex skill. The install command also
places the bundled agent runtime under the installed skill directory.

If a user gives Codex the path to this repository and asks to install the
skill, Codex should run the repository-local `install-skill.sh` and use
`--project` for the workspace that should receive `.codex/skills/sheet-ops`.
It should not copy skill files manually.

Official install flow:

1. Check whether Go is already visible in the current shell:
   `command -v go && go version`
2. If Go is missing or older than 1.25, install or upgrade Go from the
   official Go guide: https://go.dev/doc/install
3. Run the project-local install for the Codex skill and bundled runtime:
   `/path/to/sheet-ops/install-skill.sh --project /path/to/target/workspace`

`install-skill` checks Go before writing files. If Go 1.25 or newer is already
available, it does not prompt and installs without asking. If Go is missing, it
asks before sending the user to the official Go installation guide. If Go
exists but is older than 1.25, it asks before sending the user to upgrade.

If Go is installed but not visible in the current app or shell `PATH`, find the
actual Go executable first and pass it explicitly:

```bash
/path/to/sheet-ops/install-skill.sh --project /path/to/target/workspace --go-bin /absolute/path/to/go
```

Do not guess the Go path or install a second copy before checking the existing
installation.

This flow does not use a zip file, tarball, prebuilt platform binary, or
Homebrew dependency. It builds the skill-local agent binary during install.

## Local State And Evidence

The install surface is `./.codex/skills/sheet-ops`, including the bundled
runtime at `./.codex/skills/sheet-ops/agent-system`. Mutable runtime state
defaults to `./.sheet-ops-state/` in local testing mode. Artifacts and
knowledge are retained locally so the agent can inspect evidence and repair
failures without guessing. Delete `.sheet-ops-state/` to clear local case
state.

This state retention and cleanup model is for local source-install harness use.
It is not deployment-safe yet for hosted multi-tenant execution.
