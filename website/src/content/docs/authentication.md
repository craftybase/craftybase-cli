---
title: Authentication
description: How the Craftybase CLI resolves and stores your API token.
---

## Logging in

```sh
craftybase auth login          # interactive prompt (hidden input)
echo "$KEY" | craftybase auth login   # from stdin (CI)
craftybase auth login --token "$KEY"  # explicit flag
```

On success your account name + token are stored in `~/.craftybase/config.toml` (mode `0600`).

## Status & logout

```sh
craftybase auth status   # shows account, masked key, and API URL
craftybase auth logout   # removes stored credentials
```

## Token resolution precedence

The token is resolved in this order — first match wins:

1. `--token` flag
2. `CRAFTYBASE_API_TOKEN` environment variable
3. stored profile (`~/.craftybase/config.toml`)

A missing token yields a clear error and exit code `3`.
