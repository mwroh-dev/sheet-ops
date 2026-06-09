# Assist Orchestration Backlog

This backlog tracks gaps in Sheet Ops as an LLM-assisted workbook skill
harness. These items are not claims about missing Excel knowledge in the model.
They are places where the harness should better guide, constrain, verify, or
record model orchestration.

## Current Baseline

Sheet Ops now provides:

- single `sheet-ops` human-facing skill entry
- explicit planning guidance for goal, known facts, selected path, verifier
  focus, and stop condition
- supported atom capability registry and deterministic runtime execution
- advisory template research with atom/molecule/organism decision scaffolding
- compiler `template_class_plan` evidence for template-like requests
- schema-authorized `TaskSpec -> OperationIR -> Execute -> Verify -> Evidence`
  path
- curated research artifacts and cross-binding tests for advisory records

The remaining work is about making that guidance harder to ignore during live
skill use and easier to improve after repeated runs.

## Backlog

### 1. Skill Guidance Compliance Harness

Status: `planned`

Gap: The skill docs and prompts tell the model to record goal, known facts,
selected path, verifier focus, and stop condition, but not every live path is
required to emit those fields as a machine-checked artifact.

Needed:

- add a compact orchestration decision artifact schema
- require goal, known facts, missing facts, selected path, verifier focus, and
  stop condition before runtime handoff
- validate that blocked runs include a stop condition and repair direction
- validate that successful runs cite verifier evidence before final answer

Promotion gate: A public harness test fails if a representative skill run omits
the decision fields.

### 2. Live Usage Feedback Loop

Status: `planned`

Gap: Current tests prove deterministic fixtures and advisory artifact bindings,
but they do not measure whether real model-invoked skill runs consistently
follow the intended orchestration path.

Needed:

- collect redacted live-run decision artifacts
- classify deviations: unsupported claim, missing stop condition, weak verifier
  focus, skipped evidence review, or unnecessary workbook inspection
- promote recurring deviations into prompt, schema, or runtime tests

Promotion gate: A small redacted run set can be replayed or reviewed against
the orchestration decision schema.

### 3. Failure-To-Knowledge Promotion

Status: `planned`

Gap: Evidence exists, but the loop from failed workbook attempt to reusable
semantic knowledge is still mostly manual.

Needed:

- define when a failed run becomes an episodic failure record
- define when repeated failures become semantic repair knowledge
- link repair knowledge back to request-compiler ambiguity handling and
  verifier expectations

Promotion gate: A failed run can produce a sanitized knowledge candidate with
no local paths, private workbook content, or raw telemetry.

### 4. Template-Class Hint Adoption Checks

Status: `planned`

Gap: Template-class artifacts exist and are cross-bound, but live request
compiler behavior still needs stronger evidence that template-like requests
actually preserve and use those hints before narrowing to supported atoms.

Needed:

- add representative prompt-text cases for all high-priority organism classes
- assert compiler artifacts retain `template_class_plan` where applicable
- assert unsupported organism behavior stops as advisory instead of becoming a
  false capability claim

Promotion gate: Prompt-text request compiler tests cover the key organism
families, not only structured organism execution requests.

### 5. Evidence-Centered Final Answer Standard

Status: `planned`

Gap: Runtime emits verification and evidence, but the user-facing answer format
is not yet fully standardized around that evidence.

Needed:

- define final answer minimum fields for success, blocked, ambiguous, and
  failed states
- require success answers to cite verifier outcome and output artifact
- require blocked answers to cite the exact missing fact, unsupported
  capability, or policy/verifier limit

Promotion gate: Result-verifier or parent-orchestrator tests reject final
answers that claim success without evidence.

### 6. Real Workbook Regression Corpus

Status: `planned`

Gap: Deterministic fixtures are strong for bounded behavior, but the harness
needs more real-world workbook shape coverage to know where guidance breaks
down.

Needed:

- add sanitized workbook fixtures for messy headers, multiple tables, protected
  ranges, formula bands, and irregular summaries
- keep fixtures small and license-safe
- attach each fixture to the atom/template-class decision it is meant to test

Promotion gate: New real-shape fixtures must prove either supported execution,
blocked behavior, or advisory-only classification.

### 7. Stop-Condition Taxonomy

Status: `planned`

Gap: Stop conditions are described in prose, but they are not yet a compact
taxonomy that can be searched, counted, and improved.

Needed:

- define stable stop-condition IDs for missing facts, unsupported capability,
  ambiguous sheet boundary, insufficient verifier coverage, policy block, and
  advisory-only template class
- require blocked artifacts to use one or more IDs
- summarize stop-condition frequency in release diagnostics

Promotion gate: Blocked outcomes can be grouped by stop-condition ID without
reading free-form prose.

## Non-Goals

- Do not turn advisory atom/molecule/organism records into runtime authority.
- Do not claim arbitrary Excel template generation.
- Do not add broad model-calling runtime behavior inside the Go executor.
- Do not promote a capability without runtime contracts, fixtures, and
  verifier coverage.
