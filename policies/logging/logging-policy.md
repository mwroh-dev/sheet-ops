# Logging Policy

## Purpose

This repository treats orchestration logging as a mandatory operating rule, not as an optional debug aid.

## Rules

- Raw execution telemetry must be recorded for every orchestrated run.
- Telemetry must be append-only.
- Development runs keep sensitive values unmasked.
- Public skill defaults must use redacted retention unless the operator
  explicitly opts into full forensic retention.
- Telemetry, evidence, and reports must remain separate artifact layers.
- Evidence must link back to raw telemetry.
- Reports must link back to evidence.
- For the public package, raw telemetry stays internal retention only; curated
  evidence and reports are the public-facing records.

## Required Telemetry Classes

- model input and output
- agent dispatch and completion
- skill invocation
- file read, write, create, and replace events
- stream events
- final outcome events
