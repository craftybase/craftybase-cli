---
title: Pagination
description: Fetching one page, all pages, or a stream.
---

List commands are paginated.

```sh
craftybase materials list                 # first page (default size)
craftybase materials list --page 2        # a specific page
craftybase materials list --all           # fetch every page into one table
craftybase materials list --ndjson        # stream every page as NDJSON
```

`--all` is mutually exclusive with both `--ndjson` and `--page`. Both `--all` and `--ndjson` walk pages automatically using the response `meta.total_pages`.
