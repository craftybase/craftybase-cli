---
title: "stocksmith"
description: "Reference for the stocksmith command."
---

## stocksmith

Official CLI for the Stocksmith Public API

### Synopsis

stocksmith is a command-line interface for the Stocksmith Public API.

Authenticate once, then manage your inventory from the terminal.

Documentation: https://cli.stocksmith.dev/getting-started

```
stocksmith [flags]
```

### Options

```
      --api-url string   API base URL (default: https://api.stocksmith.io)
  -h, --help             help for stocksmith
      --json             Output raw API envelope (pretty-printed JSON)
      --ndjson           Output auto-paginated NDJSON stream
      --no-color         Disable ANSI color output
      --token string     API token (overrides stored credentials)
      --verbose          Show HTTP request/response detail (token redacted)
```

### SEE ALSO

* [stocksmith account](/reference/stocksmith_account/)	 - Show account information
* [stocksmith api](/reference/stocksmith_api/)	 - Make authenticated API requests
* [stocksmith auth](/reference/stocksmith_auth/)	 - Manage authentication credentials
* [stocksmith completion](/reference/stocksmith_completion/)	 - Generate shell completion scripts
* [stocksmith components](/reference/stocksmith_components/)	 - Manage components
* [stocksmith expenses](/reference/stocksmith_expenses/)	 - Manage expenses
* [stocksmith manufactures](/reference/stocksmith_manufactures/)	 - Manage manufactures
* [stocksmith materials](/reference/stocksmith_materials/)	 - Manage materials
* [stocksmith products](/reference/stocksmith_products/)	 - Manage products
* [stocksmith recipes](/reference/stocksmith_recipes/)	 - Manage recipes
* [stocksmith update](/reference/stocksmith_update/)	 - Update stocksmith to the latest release
* [stocksmith version](/reference/stocksmith_version/)	 - Print version information

