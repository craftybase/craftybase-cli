---
title: "stocksmith manufactures list"
description: "Reference for the stocksmith manufactures list command."
---

## stocksmith manufactures list

List manufactures

### Synopsis

List production runs (manufactures) from your Stocksmith account.

Filter by product, status, or start-date range. Use --all to fetch all
pages, or --ndjson for streaming NDJSON output suitable for data pipelines.

```
stocksmith manufactures list [flags]
```

### Options

```
      --all                 Fetch all pages and render as a single table
      --from string         Filter by start date on or after (ISO 8601, e.g. 2026-01-01)
  -h, --help                help for list
      --page int            Page number (1-based)
      --per-page int        Items per page (server clamps to 100)
      --product-id string   Filter by product ID
      --status string       Filter by production status: not_started, work_in_progress, completed
      --to string           Filter by start date on or before (ISO 8601)
```

### Options inherited from parent commands

```
      --api-url string   API base URL (default: https://api.stocksmith.io)
      --json             Output raw API envelope (pretty-printed JSON)
      --ndjson           Output auto-paginated NDJSON stream
      --no-color         Disable ANSI color output
      --token string     API token (overrides stored credentials)
      --verbose          Show HTTP request/response detail (token redacted)
```

### SEE ALSO

* [stocksmith manufactures](/reference/stocksmith_manufactures/)	 - Manage manufactures

