---
title: "craftybase auth login"
description: "Reference for the craftybase auth login command."
---

## craftybase auth login

Authenticate with the Craftybase API

### Synopsis

Authenticate with the Craftybase API using an API key.

The key can be provided via:
  - --token flag
  - stdin (when piped)
  - interactive prompt (when run in a terminal)

On success, credentials are saved to ~/.craftybase/config.toml.

```
craftybase auth login [flags]
```

### Options

```
  -h, --help           help for login
      --token string   API token to authenticate with
```

### Options inherited from parent commands

```
      --api-url string   API base URL (default: https://api.craftybase.com)
      --json             Output raw API envelope (pretty-printed JSON)
      --ndjson           Output auto-paginated NDJSON stream
      --no-color         Disable ANSI color output
      --verbose          Show HTTP request/response detail (token redacted)
```

### SEE ALSO

* [craftybase auth](/reference/craftybase_auth/)	 - Manage authentication credentials

