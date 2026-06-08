# Organism Preview Harness

This preview harness is not an organism runtime.

It demonstrates that supported atoms can be composed in sequence while the
atom/molecule/organism catalog remains advisory. The first preview target is
`invoice_line_item_billing` because it has a narrow, high-evidence composition:

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

Promotion rule: an organism preview can influence roadmap priority, but public
runtime support is still decided only by capability records, runtime contracts,
and operation-specific verifier coverage.
