# Failure Repair

Repair is failure-only advice in the current roadmap.

The deterministic runtime emits failure evidence and may emit a repair advice artifact. The repair advice artifact explains the likely correction, but it does not authorize automatic retry or workbook mutation.

If a run returns `BLOCKED` or `NEEDS_REVIEW`, report the failure and evidence path. Do not bypass schema validation, `OperationIR`, or deterministic verification.

