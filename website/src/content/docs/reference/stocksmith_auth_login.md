---
title: "stocksmith auth login"
description: "Reference for the stocksmith auth login command."
---

## stocksmith auth login

Authenticate with the Stocksmith API

### Synopsis

Authenticate with the Stocksmith API using an API key.

The key can be provided via:
  - --token flag
  - stdin (when piped)
  - interactive prompt (when run in a terminal)

On success, credentials are saved to ~/.stocksmith/config.toml.

```
stocksmith auth login [flags]
```

### Options

```
  -h, --help           help for login
      --token string   API token to authenticate with
```

### Options inherited from parent commands

```
      --api-url string   API base URL (default: https://api.stocksmith.io)
      --json             Output raw API envelope (pretty-printed JSON)
      --ndjson           Output auto-paginated NDJSON stream
      --no-color         Disable ANSI color output
      --verbose          Show HTTP request/response detail (token redacted)
```

### SEE ALSO

* [stocksmith auth](/reference/stocksmith_auth/)	 - Manage authentication credentials

