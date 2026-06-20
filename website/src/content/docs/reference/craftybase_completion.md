---
title: "craftybase completion"
description: "Reference for the craftybase completion command."
---

## craftybase completion

Generate shell completion scripts

### Synopsis

Generate shell completion scripts for craftybase.

To load completions:

Bash:
  $ source <(craftybase completion bash)

Zsh:
  $ source <(craftybase completion zsh)

Fish:
  $ craftybase completion fish | source

PowerShell:
  PS> craftybase completion powershell | Out-String | Invoke-Expression


```
craftybase completion [bash|zsh|fish|powershell]
```

### Options

```
  -h, --help   help for completion
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

