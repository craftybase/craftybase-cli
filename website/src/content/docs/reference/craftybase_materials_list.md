---
title: "craftybase materials list"
description: "Reference for the craftybase materials list command."
---

## craftybase materials list

List materials

### Synopsis

List materials from your Craftybase account.

Filter by SKU, name, category, or state. Use --all to fetch all pages,
or --ndjson for streaming NDJSON output suitable for data pipelines.

```
craftybase materials list [flags]
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

* [craftybase materials](/reference/craftybase_materials/)	 - Manage materials

