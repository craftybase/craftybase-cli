---
title: "stocksmith api"
description: "Reference for the stocksmith api command."
---

## stocksmith api

Make authenticated API requests

### Synopsis

Make authenticated requests to the Stocksmith API.

The path must be the full API path starting with /api/v1/.

Examples:
  stocksmith api GET /api/v1/account
  stocksmith api GET /api/v1/materials
  stocksmith api GET "/api/v1/materials?sku=WAX-001"

```
stocksmith api <METHOD> <path> [flags]
```

### Options

```
  -h, --help   help for api
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

* [stocksmith](/reference/stocksmith/)	 - Official CLI for the Stocksmith Public API

