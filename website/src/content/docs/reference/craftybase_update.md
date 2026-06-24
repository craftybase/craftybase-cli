---
title: "craftybase update"
description: "Reference for the craftybase update command."
---

## craftybase update

Update craftybase to the latest release

### Synopsis

Update the craftybase binary in place to the latest GitHub release.

Downloads the release archive for your platform, verifies its SHA-256 checksum,
and atomically replaces the running binary. Homebrew installs should use
'brew upgrade craftybase'; Windows users download the release zip manually.

```
craftybase update [flags]
```

### Options

```
      --check   Check for a newer release without installing
  -h, --help    help for update
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

