# Organism Preview Harness

This preview harness is not an organism runtime.

It demonstrates that supported atoms can be composed in sequence while the
atom/molecule/organism catalog remains advisory.

## Invoice Preview

The first preview target is `invoice_line_item_billing` because it has a narrow,
high-evidence composition:

1. append structured line-item rows
2. extend line-item total and tax formulas into the appended rows
3. add explicit list/dropdown validation to the SKU input range
4. protect explicit formula cells while leaving declared input cells editable
5. verify the source workbook remains preserved at each atom boundary

The current fixture-backed preview covers row-oriented line item growth, formula
extension, explicit list/dropdown validation on a declared input range, and
explicit formula/input range protection. It
does not claim printable invoice generation, metadata header synthesis, formula
region auto-discovery, numeric/date validation, formula-backed validation lists,
or full invoice document generation.

## Timesheet Preview

The second preview target is `timesheet_hours_log` because it is the narrowest
p0 organism that exercises period-copy behavior without requiring carry-forward
continuity:

1. copy the `Week1` timesheet sheet to a new `Week2` sheet
2. verify the copied sheet preserves values and formulas
3. verify the source workbook remains preserved at the atom boundary

This preview demonstrates explicit sheet-copy mechanics only. It does not claim
week label rewriting, cleared time-entry inputs, payroll carry-forward, or
roll-forward continuity.

Promotion rule: an organism preview can influence roadmap priority, but public
runtime support is still decided only by capability records, runtime contracts,
and operation-specific verifier coverage.
