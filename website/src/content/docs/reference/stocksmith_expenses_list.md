---
title: "stocksmith expenses list"
description: "Reference for the stocksmith expenses list command."
---

## stocksmith expenses list

List expenses

### Synopsis

List expenses (purchases) from your Stocksmith account.

An expense is a supplier purchase — header totals plus the materials and costs
on each line. Filter by purchase-date range, change time, category, or supplier.
Use --all to fetch all pages, or --ndjson for streaming NDJSON output suitable
for data pipelines.

```
stocksmith expenses list [flags]
```

### Options

```
      --all                    Fetch all pages and render as a single table
      --category-id string     Filter by line-item category ID
      --from string            Filter by purchase date on or after (ISO 8601, e.g. 2026-01-01)
  -h, --help                   help for list
      --page int               Page number (1-based)
      --per-page int           Items per page (server clamps to 100)
      --supplier-id string     Filter by supplier ID
      --to string              Filter by purchase date on or before (ISO 8601)
      --updated-since string   Return expenses updated on or after this time (ISO 8601; includes line-item edits)
```

### Options inherited from parent commands

```
      --api-url string   API base URL (default: https://api.stocksmith.dev)
      --json             Output raw API envelope (pretty-printed JSON)
      --ndjson           Output auto-paginated NDJSON stream
      --no-color         Disable ANSI color output
      --token string     API token (overrides stored credentials)
      --verbose          Show HTTP request/response detail (token redacted)
```

### SEE ALSO

* [stocksmith expenses](/reference/stocksmith_expenses/)	 - Manage expenses

