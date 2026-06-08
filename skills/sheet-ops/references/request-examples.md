# Request Examples

Examples map user language to capability families. They do not bypass contracts.

- "Summarize orders by region and total amount" -> `group_summarize`
- "Highlight rows where amount is greater than 100000" -> `highlight_threshold`
- "Append customer tier from the Customers sheet by customer_id" -> `join_lookup`
- "Append these new line-item rows to LineItems" -> `append_structured_rows`
- "Extend the formulas from row 2 into row 3 for columns D and E" -> `extend_table_formulas`
- "Add a status dropdown to D2:D50 with draft, sent, paid" -> `add_data_validation`

After mapping a request, create or select a `UseEnvelope` and use the
skill-owned runtime handoff. Do not type or reconstruct internal launcher
commands in the runner pane.
