---
title: Getting Started
description: Install the Stocksmith CLI, authenticate, and run your first command.
---

## Install

```sh
curl -fsSL https://cli.stocksmith.dev/install | bash
```

This downloads the right binary for your OS/architecture, verifies its checksum, and installs `stocksmith` to `/usr/local/bin` (or `~/.local/bin`). Homebrew (`brew install craftybase/tap/stocksmith`) and `go install github.com/craftybase/stocksmith-cli/cmd/stocksmith@latest` also work.

## Authenticate

```sh
stocksmith auth login
```

Paste your Stocksmith API key when prompted (input is hidden). Credentials are saved to `~/.stocksmith/config.toml`. See [Authentication](/authentication/).

## First command

```sh
stocksmith materials list
```

Add `--json` for the raw API envelope or `--ndjson` to stream every page. See [Output Formats](/output-formats/).
