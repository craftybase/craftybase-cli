---
title: craftybase api
description: Reference for the craftybase api command.
---

## craftybase api

Make authenticated API requests

### Synopsis

Make authenticated requests to the Craftybase API.

The path must be the full API path starting with /api/v1/.

Examples:
  craftybase api GET /api/v1/account
  craftybase api GET /api/v1/materials
  craftybase api GET "/api/v1/materials?sku=WAX-001"

```
craftybase api <METHOD> <path> [flags]
```

### Options

```
  -h, --help   help for api
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

* [craftybase](/reference/craftybase/)	 - Official CLI for the Craftybase Public API

