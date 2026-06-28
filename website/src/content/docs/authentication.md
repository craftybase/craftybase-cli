---
title: Authentication
description: How the Stocksmith CLI resolves and stores your API token.
---

## Logging in

```sh
stocksmith auth login          # interactive prompt (hidden input)
echo "$KEY" | stocksmith auth login   # from stdin (CI)
stocksmith auth login --token "$KEY"  # explicit flag
```

On success your account name + token are stored in `~/.stocksmith/config.toml` (mode `0600`).

## Status & logout

```sh
stocksmith auth status   # shows account, masked key, and API URL
stocksmith auth logout   # removes stored credentials
```

## Token resolution precedence

The token is resolved in this order — first match wins:

1. `--token` flag
2. `STOCKSMITH_API_TOKEN` environment variable
3. stored profile (`~/.stocksmith/config.toml`)

A missing token yields a clear error and exit code `3`.
