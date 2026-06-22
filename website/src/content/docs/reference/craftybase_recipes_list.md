---
title: "craftybase recipes list"
description: "Reference for the craftybase recipes list command."
---

## craftybase recipes list

List recipes

### Synopsis

List recipes (bills of materials) from your Craftybase account.

A recipe is the formulation for a product or variation — the materials and
quantities consumed per batch, with cost and COGS rollups. Filter by product,
variation, or change time. Use --all to fetch all pages, or --ndjson for
streaming NDJSON output suitable for data pipelines.

```
craftybase recipes list [flags]
```

### Options

```
      --all                    Fetch all pages and render as a single table
  -h, --help                   help for list
      --page int               Page number (1-based)
      --per-page int           Items per page (server clamps to 100)
      --product-id string      Filter by product ID
      --updated-since string   Return recipes updated on or after this time (ISO 8601, e.g. 2026-01-01)
      --variation-id string    Filter by variation ID
```

### Options inherited from parent commands

```
      --api-url string   API base URL (default: https://api.craftybase.com)
      --json             Output raw API envelope (pretty-printed JSON)
      --ndjson           Output auto-paginated NDJSON stream
      --no-color         Disable ANSI color output
      --token string     API token (overrides stored credentials)
      --verbose          Show HTTP request/response detail (token redacted)
```

### SEE ALSO

* [craftybase recipes](/reference/craftybase_recipes/)	 - Manage recipes

