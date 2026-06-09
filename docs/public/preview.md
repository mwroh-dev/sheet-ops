# Preview Status

The claim source of truth is `contracts/public_preview/claims.json`.

## Supported today

- experimental local workflow surface for Codex/Claude workbook agents
- single human-facing `sheet-ops` skill entry
- currently implemented public path compositions: `group_summarize`,
  `highlight_threshold`, and `join_lookup`
- schema-authorized `TaskSpec -> OperationIR -> Execute -> Verify -> Evidence`
  runtime path
- deterministic fixture-backed harness smoke with checked-in frozen demo
  evidence
- render artifact emission when renderer tools exist, or explicit unavailable
  evidence when they do not

## Preview limitations

- live Codex delegation smoke is a non-blocking opt-in diagnostic, not a
  required public harness gate
- render artifact emission records file-open and preview artifact properties;
  it is not visual quality verification
- Sheet Ops is not hosted or multi-tenant ready
- `write_values` is not a public agent capability; it remains an internal
  runtime primitive until public orchestration and e2e tests exist

## Follow-up

- harden the optional live smoke with stronger timeout/process supervision
- add cross-platform visual quality assertions only after renderer support is
  proven
- harden state/privacy boundaries before any hosted claim
- promote `write_values` only after public orchestration and e2e coverage exist
