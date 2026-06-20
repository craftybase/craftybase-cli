---
title: "craftybase products list"
description: "Reference for the craftybase products list command."
---

## craftybase products list

List products

```
craftybase products list [flags]
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

* [craftybase products](/reference/craftybase_products/)	 - Manage products

