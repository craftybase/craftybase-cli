---
title: Output Formats
description: Tables for humans, JSON and NDJSON for machines.
---

By default, list/show commands print an aligned, colored table.

## JSON

```sh
craftybase materials list --json
```

Prints the full API envelope (pretty-printed). Pipe into `jq`:

```sh
craftybase materials list --json | jq '.materials[] | {id, name}'
```

## NDJSON (streaming, auto-paginated)

```sh
craftybase materials list --ndjson
```

Emits one JSON object per line across all pages — ideal for pipelines. `--json` and `--ndjson` are mutually exclusive.

## Color

Color is on for interactive terminals and off when piped. Disable explicitly with `--no-color` or the `NO_COLOR` environment variable.
