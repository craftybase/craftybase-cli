---
title: "craftybase components list"
description: "Reference for the craftybase components list command."
---

## craftybase components list

List components

```
craftybase components list [flags]
```

### Options

```
      --all               Fetch all pages and render as a single table
      --category string   Filter by category name
  -h, --help              help for list
      --name string       Filter by name (substring match)
      --page int          Page number (1-based)
      --per-page int      Items per page (server clamps to 100)
      --sku string        Filter by SKU (exact match)
      --state string      Filter by state: active, archived, all
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

* [craftybase components](/reference/craftybase_components/)	 - Manage components

