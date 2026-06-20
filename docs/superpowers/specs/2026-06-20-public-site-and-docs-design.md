# Public Site & Docs (cli.craftybase.dev) — Design

**Date:** 2026-06-20
**Status:** Approved (pending spec review)
**Scope:** Public-facing website, install script, documentation, and LLM-source for the Craftybase CLI

## Goal

Stand up `https://cli.craftybase.dev` as the public home of the CLI: a landing page, a
one-line installer at `/install`, documentation (`/getting-started` and friends), an
auto-generated command reference, and an LLM-consumable copy of the docs. The repo
(`craftybase/craftybase-cli`) becomes public as part of going live.

Reference best practices: `https://cli.sentry.dev/` and `https://github.com/getsentry/cli`
(docs co-located with the CLI, command reference generated from command definitions,
one-line `curl … | bash` install, minimal nav). Our CLI is a **Go/Cobra binary** (Sentry's
is a Node package), so the install story is curl|bash + Homebrew + `go install` (no npm),
and the reference is generated via Cobra's `doc` package.

## Scope

**In scope:**
- Astro + Starlight static site in the CLI repo, hosted on GitHub Pages at `cli.craftybase.dev`.
- `/install` POSIX shell installer that fetches the right GoReleaser release asset.
- Hand-written guides + auto-generated command reference.
- LLM-friendly output (`llms.txt`, `llms-full.txt`, per-page markdown).
- Repo hygiene to support going public: `LICENSE`, `README.md`.
- CI to build/deploy the site and keep the generated reference in sync.

**Out of scope (YAGNI):**
- Docs versioning / multi-version selector (single "latest" only).
- Search beyond Starlight's built-in (Pagefind).
- Windows `curl|bash` install (documented via Homebrew/manual instead).
- Bunny CDN/edge in front of the site (Bunny is DNS only).
- Changing the API docs at `craftybase.com/docs/api` (separate property).

## Architecture & Repo Layout

Single repo (`craftybase-cli`), mirroring `getsentry/cli`. New top-level `website/` for the
site; a small Go generator for the reference; CI under `.github/workflows/`.

```
website/                              # Astro + Starlight site (Node toolchain)
  package.json
  astro.config.mjs                    # Starlight config: nav, sidebar, starlight-llms-txt
  src/
    content/
      docs/
        index.mdx                     # landing (custom hero) — Starlight splash template
        getting-started.md
        authentication.md
        output-formats.md
        configuration.md
        pagination.md
        agents.md                     # "Using with agents/LLMs"
        reference/                    # ← AUTO-GENERATED, committed
          craftybase.md
          craftybase_account.md
          craftybase_materials.md
          craftybase_materials_list.md
          ... (one per command/subcommand)
  public/
    install                           # POSIX installer (served at /install)
    install.sh                        # identical copy, browser-viewable path
    CNAME                             # "cli.craftybase.dev" (GitHub Pages custom domain)
    .nojekyll                         # bypass Jekyll on Pages
cmd/gen-docs/main.go                  # cobra/doc → website/src/content/docs/reference/*.md
Makefile                             # `make docs` target (regenerate reference)
.github/workflows/site.yml            # build + deploy site to Pages; reference-sync check
README.md                            # repo front page (install, quickstart, link to site)
LICENSE                              # also unblocks GoReleaser archives
```

`docs/` already exists for Superpowers specs/plans; the website is deliberately a separate
top-level dir to avoid collision.

## Tech Stack

- **Site:** Astro + Starlight (static output). Built with Node/npm. `site: "https://cli.craftybase.dev"`,
  no `base` (custom domain serves at root).
- **Reference generation:** Go program `cmd/gen-docs` using `github.com/spf13/cobra/doc`
  (`GenMarkdownTreeCustom`) with a frontmatter prepender (Starlight needs `title`/`description`)
  and a link handler that rewrites inter-command links to Starlight routes.
- **LLM output:** `starlight-llms-txt` plugin → `/llms.txt`, `/llms-full.txt`, per-page `.md`.
- **Hosting:** GitHub Pages, deployed by GitHub Actions (`actions/configure-pages`,
  `upload-pages-artifact`, `deploy-pages`).
- **DNS:** Bunny.net — a single `CNAME` record `cli` → `craftybase.github.io`. GitHub
  provisions HTTPS automatically. No Bunny hosting/CDN.

## Hosting & Deploy

- A GitHub Actions workflow (`site.yml`) on push to the default branch and on release:
  1. **Reference-sync check:** run the Go generator into a temp dir (or `make docs`) and
     `git diff --exit-code` the committed `reference/` — fail if stale.
  2. Build: `cd website && npm ci && npm run build`.
  3. Deploy the `website/dist` artifact to GitHub Pages.
- Custom domain via the committed `public/CNAME`; HTTPS auto-provisioned by GitHub.
- **`/install` content-type:** GitHub Pages cannot set custom headers, so the extensionless
  `/install` is served as `application/octet-stream`. This does **not** affect
  `curl -fsSL https://cli.craftybase.dev/install | bash` (curl pipes bytes to bash regardless
  of content-type). `public/install.sh` is provided as a browser-viewable duplicate.
- **Repo-visibility gate:** free GitHub Pages and downloadable release assets require the repo
  to be **public**. Therefore the live site and a working `/install` come online only after the
  repo is public (Phase 5). Components are validated in isolation before then (shellcheck +
  dry-run for the script, local `npm run build` for the site).

## Install Script (`/install`)

POSIX `sh` (portable; runs under sh/bash/zsh). Behavior:

1. Detect platform: `uname -s` → `darwin`/`linux`; `uname -m` → map `x86_64`→`amd64`,
   `aarch64`/`arm64`→`arm64`. Unsupported OS (e.g. Windows/MINGW) → print Homebrew/manual
   instructions and exit non-zero.
2. Resolve version: latest release tag via
   `https://api.github.com/repos/craftybase/craftybase-cli/releases/latest`, unless
   `CRAFTYBASE_VERSION` is set.
3. Download the asset `craftybase_<version>_<os>_<arch>.tar.gz` and
   `craftybase_<version>_checksums.txt` from the release; verify SHA-256 against the checksum
   file (`sha256sum`/`shasum -a 256`); abort on mismatch.
4. Extract `craftybase`; install to `/usr/local/bin` if writable, else `~/.local/bin`
   (override with `CRAFTYBASE_INSTALL_DIR`). If the target isn't on `PATH`, print a warning
   with the line to add.
5. Print success + next step: `craftybase auth login`. Use a temp dir, clean up on exit,
   fail loudly with actionable messages.

**Coupling:** asset names must match `.goreleaser.yml` `name_template`
(`craftybase_{{.Version}}_{{.Os}}_{{.Arch}}`) and the checksums name
(`craftybase_{{.Version}}_checksums.txt`). CI shellchecks the script; a comment in both files
cross-references the coupling. `{{.Version}}` has no `v` prefix; `{{.Os}}`/`{{.Arch}}` are
lowercase GOOS/GOARCH.

## Site Content & Information Architecture

**Landing (`index.mdx`):** teal `CRAFTYBASE` hero + one-line value prop; copyable
`curl -fsSL https://cli.craftybase.dev/install | bash`; Homebrew (`brew install craftybase/tap/craftybase`)
and `go install` alternatives; a short feature showcase (manage inventory from the terminal ·
`--json`/`--ndjson` for pipelines & agents · authenticate once); CTAs to Getting Started + GitHub.

**Docs sidebar (hand-written, grounded in the actual CLI conventions):**
- **Getting Started** — install → `craftybase auth login` → `craftybase materials list`.
- **Authentication** — API token, `auth login`/`status`/`logout`, `CRAFTYBASE_API_TOKEN`,
  resolution precedence (flag → env → stored profile → default).
- **Output formats** — default table vs `--json` (full envelope) vs `--ndjson` (streamed,
  auto-paginated); `--no-color`/`NO_COLOR`; `jq` piping examples.
- **Configuration** — `~/.craftybase/config.toml`, `profiles.default`, atomic `0600` writes,
  `--api-url`/`CRAFTYBASE_API_URL`.
- **Pagination** — `--all`, `--ndjson`, `--page` (and their mutual exclusivity).
- **Command Reference** — auto-generated, one page per command (account, api, auth, materials,
  products, completion, version, + subcommands).
- **Using with agents/LLMs** — point models at `cli.craftybase.dev/llms.txt`; prefer
  `--json`/`--ndjson` for tool output.

Content reflects real behavior (enveloped responses, money as `{amount, currency_code}`
string pairs, exit codes 401→3/404→4, etc.) but the prose lives in the site, not duplicated
from code.

## LLM-Friendliness

`starlight-llms-txt` generates, at build time:
- `/llms.txt` — curated index (title, description, links to each doc) per llmstxt.org.
- `/llms-full.txt` — all docs concatenated into one file for direct context loading.
- Per-page markdown (the `.md` source behind each rendered page).

The "Using with agents/LLMs" page documents these endpoints. Because the command reference is
generated from the live command tree, the LLM source stays accurate by construction.

## Command-Reference Generation Pipeline

- `cmd/gen-docs/main.go` obtains the root `*cobra.Command` and calls
  `doc.GenMarkdownTreeCustom(root, outDir, prepender, linkHandler)`. Because `commands.rootCmd`
  is currently unexported, this requires adding a small exported accessor to the `commands`
  package (e.g. `func RootCmd() *cobra.Command` returning the assembled root) that `gen-docs`
  imports — the generator must not duplicate the command tree.
  - `prepender(filename)` emits Starlight frontmatter (`---\ntitle: …\ndescription: …\n---`)
    derived from the command's name/short.
  - `linkHandler(name)` rewrites cross-command links to Starlight routes under `/reference/`,
    using the filename stem as the slug (e.g. `craftybase_materials_list` →
    `/reference/craftybase_materials_list/`).
  - Disable Cobra's auto-generated timestamp footer (`cmd.DisableAutoGenTag = true`) so output
    is deterministic and diff-stable.
- Output dir: `website/src/content/docs/reference/`, committed.
- `make docs` runs the generator. CI re-runs it and fails on any diff (same sync philosophy as
  the existing `roothelp` `TestCommandRowsCoverAllCommands`/`TestFlagRowsCoverPersistentFlags`).

## Phasing

1. **Repo hygiene + install script.** Add `LICENSE` (MIT unless told otherwise) and
   `README.md` (badges, quick install, link to site) — also unblocks GoReleaser archives that
   already list both. Write `public/install` (+ `install.sh`), shellcheck it, add CI lint.
2. **Site scaffold + content.** Astro+Starlight project under `website/`, landing page, and the
   hand-written guides (getting-started, authentication, output-formats, configuration, pagination).
3. **Command reference generation.** `cmd/gen-docs`, `make docs`, the reference pages, the CI
   sync check, wire `reference/` into the Starlight sidebar.
4. **LLM + polish.** `starlight-llms-txt` (llms.txt/llms-full.txt), the agents page, OG/SEO
   metadata, a 404 page.
5. **Go live.** Flip the repo public → enable GitHub Pages (Actions deploy) → add the Bunny
   `CNAME` + GitHub custom domain → verify `cli.craftybase.dev`, `/getting-started`, `/install`
   end-to-end (real release download + checksum).

Each phase is independently shippable; go-live is last to honor "expose the repo eventually."
Going public earlier (to test live during build) is a valid alternative if desired — it only
moves Phase 5's repo-flip forward.

## Constraints, Dependencies, Assumptions

- Repo path `github.com/craftybase/craftybase-cli`; Go 1.26; Cobra-based; GoReleaser already
  publishes cross-platform archives + the `craftybase/homebrew-tap` tap.
- DNS for `craftybase.dev` is on Bunny.net (CNAME record addable).
- `brand.InstallScriptURL` already points at `https://cli.craftybase.dev/install`.
- **Open follow-up (flagged, not decided here):** `brand.DocsURL` currently points at
  `https://craftybase.com/docs/api` (the API docs). Once the CLI site is live, consider pointing
  the CLI's "Learn more" footer at `cli.craftybase.dev/getting-started`. Left as a Phase-5
  decision.
- License choice (MIT vs other) to be confirmed during Phase 1.

## Testing / Verification

- **Install script:** `shellcheck` in CI; a dry-run mode or local test against a public release;
  full end-to-end (download + checksum + PATH) verified in Phase 5 once the repo is public.
- **Reference sync:** CI runs the generator and asserts no diff vs committed `reference/`.
- **Site build:** `npm run build` must succeed in CI; link-check the built output.
- **Go live:** manual verification that `cli.craftybase.dev`, `/getting-started`, `/reference/*`,
  `/llms.txt`, and `curl …/install | bash` all work against a real release.
