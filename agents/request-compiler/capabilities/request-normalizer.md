# Request Normalizer Capability

Agent-owned capability for converting raw workbook request text into normalized
semantic fields.

Responsibilities:

- accept raw natural-language workbook request text
- emit constrained semantic fields and ambiguity markers
- stay within the request-compiler vocabulary boundary

Boundaries:

- this is not a root-level discoverable skill
- checkpoint ownership stays with `agents/request-compiler/`
- deterministic execution authority stays in `contracts/` and `runtime/`
