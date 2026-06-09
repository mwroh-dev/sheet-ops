# Capabilities

Capability records live under `contracts/capabilities/records`. This reference is guidance over that registry, not a second authority.

Operating model:

`LLM is planner/compiler. Go is executor/verifier. Schema is authority.`

The skill routes execution through the skill-owned runtime handoff, not through
a runner-pane command.

## Public agent capabilities

- `group_summarize`: use when the request asks to summarize rows by keys, such as "summarize revenue by region".
- `highlight_threshold`: use when the request asks to highlight rows or cells over, under, or equal to a threshold.
- `join_lookup`: use when the request asks to merge or append lookup-sheet data by key.

These records have `status: supported` and `exposure: public_agent_capability`.
Before selecting a capability, read the matching machine-readable record in
`contracts/capabilities/records`.

## Internal runtime primitives

- `write_values`: `status: supported`, `exposure: runtime_primitive`. This is
  a deterministic runtime primitive for tests and lower-level execution paths.
  It is not selected directly by the public agent flow.

## Agent decision table

| Signal | Classification | Agent action |
| --- | --- | --- |
| User asks to summarize, group, aggregate, or create a summary sheet | Supported public capability | Select `group_summarize` after reading the registry record. |
| User asks to highlight values over/under/equal to a threshold | Supported public capability | Select `highlight_threshold` after reading the registry record. |
| User asks to merge lookup data by key | Supported public capability | Select `join_lookup` after reading the registry record. |
| Runtime needs to write literal cells inside a verified lower-level path | Internal primitive | Use `write_values` only through runtime-owned flows, not as a public request capability. |
| Deterministic smoke uses fixture-backed specialist decisions | Deterministic smoke | Treat as the required public harness gate, not live LLM delegation. |
| Live Codex is available and the caller opts in | Live smoke | Run only as a non-blocking diagnostic. |
| Hosted deployment, visual quality assertions, or public `write_values` is requested | Preview limitation | Do not claim support; report the limitation or require additional hardening/tests. |

Install contract reminder:

```bash
/path/to/sheet-ops/install-skill.sh --project /path/to/target/workspace
```

Go 1.25 or newer is required. If Go is ready, install-skill does not prompt. If Go is missing or outdated, it asks before continuing. The install flow does not use a zip file, tarball, or prebuilt platform binary.
