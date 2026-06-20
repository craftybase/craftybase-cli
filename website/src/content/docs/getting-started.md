---
title: Getting Started
description: Install the Craftybase CLI, authenticate, and run your first command.
---

## Install

```sh
curl -fsSL https://cli.craftybase.dev/install | bash
```

This downloads the right binary for your OS/architecture, verifies its checksum, and installs `craftybase` to `/usr/local/bin` (or `~/.local/bin`). Homebrew (`brew install craftybase/tap/craftybase`) and `go install github.com/craftybase/craftybase-cli/cmd/craftybase@latest` also work.

## Authenticate

```sh
craftybase auth login
```

Paste your Craftybase API key when prompted (input is hidden). Credentials are saved to `~/.craftybase/config.toml`. See [Authentication](/authentication/).

## First command

```sh
craftybase materials list
```

Add `--json` for the raw API envelope or `--ndjson` to stream every page. See [Output Formats](/output-formats/).
