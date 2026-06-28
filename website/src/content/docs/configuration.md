---
title: Configuration
description: The config file, profiles, and the API URL.
---

## Config file

Credentials live in `~/.stocksmith/config.toml`, keyed by profile (`profiles.default`), written atomically with `0600` permissions.

## API URL

The API base URL is resolved as: `--api-url` flag → `STOCKSMITH_API_URL` env → stored profile → default (`https://api.stocksmith.dev`).

## Global flags

| Flag | Purpose |
| --- | --- |
| `--json` | Raw API envelope (pretty-printed) |
| `--ndjson` | Auto-paginated NDJSON stream |
| `--no-color` | Disable ANSI color |
| `--token <token>` | Override stored credentials |
| `--api-url <url>` | Override the API base URL |
| `--verbose` | Show HTTP request/response detail (token redacted) |
