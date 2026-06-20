---
title: Using with Agents & LLMs
description: Point models at the docs and use machine-readable output.
---

## Feed the docs to a model

The full documentation is available as plain text for LLM context:

- **Index:** [`/llms.txt`](/llms.txt) — a curated map of every page.
- **Full text:** [`/llms-full.txt`](/llms-full.txt) — all docs concatenated.

Any page is also available as raw markdown by appending `.md` to its URL.

## Machine-readable command output

Prefer `--json` (one envelope) or `--ndjson` (one object per line, auto-paginated) when an agent consumes CLI output:

```sh
craftybase materials list --json
craftybase materials list --ndjson
```

Exit codes are stable: `0` success, `3` auth error, `4` not found, `1` otherwise.
