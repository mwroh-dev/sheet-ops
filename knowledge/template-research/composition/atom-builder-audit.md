# Atom Builder Audit

이 문서는 code-gen assist 방향을 기록한다. 목적은 atom별로 반복되는 빌딩 방식을
정리하는 것이며, 새 runtime interface를 만들거나 capability support를 주장하지 않는다.

| Atom | Status | Input | Plan | Result | Verification |
| --- | --- | --- | --- | --- | --- |
| `group_summarize` | supported | source sheet, filters, group keys, metrics, target sheet, summary mode | inspect headers and rows, build summary rows or formula expectations | summary sheet with values or formulas | source preserved, summary rows match, formulas calculate |
| `highlight_threshold` | supported | source sheet, target column, operator, threshold, color | inspect target column, evaluate matching rows, apply highlight style | source sheet copy with highlighted rows/cells | source preserved, matching rows highlighted, color matches |
| `join_lookup` | supported | source sheet, lookup sheet, join key, source columns, lookup columns, target sheet | build key index, enrich source rows, create result sheet | joined result sheet | source preserved, appended columns match lookup rows |
| `write_values` | runtime primitive | target sheet and cell values | validate write boundary, write exact cells | workbook with written cells | source preserved, written values match |
| `append_structured_rows` | planned | table anchor, canonical headers, rows, propagation policy | detect table boundary, map headers, append rows, copy formulas and validation | appended row range with warnings | row count grows, appended rows join summaries, copied rules match |
| `extend_table_formulas` | planned | formula source range, extension target, relative/absolute reference policy | infer formula pattern, fill row/column/period target, record expectations | extended formula cells | formulas exist in target, references are compatible, calculated values match |
| `copy_period_sheet` | supported for explicit sheet copy | source sheet, target sheet | copy source sheet into target sheet | new target sheet | source workbook preserved, target sheet exists, values formulas and styles match source |
| `roll_forward_period` | supported for explicit mappings | source sheet, target sheet, closing-to-opening cell mappings | calculate declared closing cells and write declared opening cells | next-period opening cells updated | source preserved, closing equals next opening, non-mapped formulas unchanged |
| `normalize_headers` | supported for explicit mappings | source sheet, header row, explicit from/to mappings | locate declared headers and plan header-cell rewrites | normalized header cells | source preserved, mapped headers match, non-header rows and formulas unchanged |
| `create_pivot_summary` | planned | source table, dimensions, measures, output anchor | group rows into pivot-like summary, preserve source | summary table or pivot-like output | grouped totals match source, dimensions are complete |
| `add_data_validation` | supported for explicit list/dropdown rules | source sheet, target ranges, list rule type, allowed values, allow blank | build list validation rules for declared ranges | validation rules on input ranges | source workbook preserved, validation rule present, allowed values match |
| `protect_formula_cells` | supported for explicit ranges | source sheet, formula ranges, input ranges, optional password | separate locked formula cells from editable inputs | protected formula cells and sheet protection settings | source workbook preserved, formula cells locked, input cells unlocked, sheet protection enabled |
| `reconcile_tables` | planned | left/right sources, key mapping, match rules | normalize keys, match rows, bucket missing and mismatched records | reconciliation result table | matched/missing/mismatch buckets match expected cases |
| `generate_printable_form` | planned | metadata, line items, totals, output region | materialize reviewable document layout from structured data | printable/reviewable region | required fields populated, totals match source |

## 적용 규칙

새 organism이나 molecule을 만들 때 이 audit는 참고 자료다. 현재 supported status는
`contracts/capabilities`와 runtime verifier가 결정한다. Planned atom은 opportunity
registry에 남겨야 하며 supported capability enum에 추가하지 않는다.
