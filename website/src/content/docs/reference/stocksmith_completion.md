---
title: "stocksmith completion"
description: "Reference for the stocksmith completion command."
---

## stocksmith completion

Generate shell completion scripts

### Synopsis

Generate shell completion scripts for stocksmith.

To load completions:

Bash:
  $ source <(stocksmith completion bash)

Zsh:
  $ source <(stocksmith completion zsh)

Fish:
  $ stocksmith completion fish | source

PowerShell:
  PS> stocksmith completion powershell | Out-String | Invoke-Expression


```
stocksmith completion [bash|zsh|fish|powershell]
```

### Options

```
  -h, --help   help for completion
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

* [stocksmith](/reference/stocksmith/)	 - Official CLI for the Stocksmith Public API

