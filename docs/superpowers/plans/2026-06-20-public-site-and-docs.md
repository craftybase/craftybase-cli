# Public Site & Docs (cli.craftybase.dev) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stand up `https://cli.craftybase.dev` — landing page, `/install` one-liner, hand-written docs, an auto-generated command reference, and an LLM source — as an Astro + Starlight site in the CLI repo, hosted on GitHub Pages.

**Architecture:** A `website/` Astro+Starlight site lives in the repo alongside the Go CLI. A small Go program (`cmd/gen-docs`) renders the Cobra command tree to committed markdown that Starlight consumes. The install script and built site are served from GitHub Pages (deployed by GitHub Actions); Bunny.net DNS points the subdomain at Pages. Going public gates the live deploy.

**Tech Stack:** Go 1.26 + Cobra (`spf13/cobra/doc`), Astro + Starlight (Node 20 / npm), `starlight-llms-txt`, GitHub Pages + Actions, POSIX `sh`.

## Global Constraints

- Repo: `github.com/craftybase/craftybase-cli`; Go 1.26; Cobra-based CLI; GoReleaser already publishes cross-platform archives + the `craftybase/homebrew-tap` tap.
- Site canonical URL is `https://cli.craftybase.dev`; Astro config sets `site: 'https://cli.craftybase.dev'` with **no `base`** (custom domain serves at root). Static output (no SSR adapter).
- End-user install paths are **curl|bash, Homebrew, `go install`** — never npm/npx (the published artifact is a Go binary, not a Node package).
- Install asset names MUST match `.goreleaser.yml` `name_template`: archive `craftybase_<version>_<os>_<arch>.tar.gz`, checksums `craftybase_<version>_checksums.txt`. `<version>` has **no leading `v`**; `<os>`/`<arch>` are lowercase GOOS/GOARCH (`darwin`/`linux`, `amd64`/`arm64`).
- `/install` is served extensionless (GitHub Pages → `application/octet-stream`; fine for `curl … | bash`). A duplicate `/install.sh` is provided for browser viewing.
- Command reference is **auto-generated and committed**; CI fails on drift. Generation requires an exported `commands.RootCmd() *cobra.Command` accessor and `root.DisableAutoGenTag = true` for deterministic output.
- LLM source via `starlight-llms-txt` → `/llms.txt` + `/llms-full.txt`.
- Free GitHub Pages requires the repo to be **public**; the live deploy + DNS land in Phase 5 (go-live). Components are verified in isolation before then.
- License: **MIT** unless the user says otherwise (confirm in Task 1).
- Branch: `feat/public-site` (already created off `main`).
- This branch is based on `main`, which currently has commands `account, api, auth, materials, completion, version` (not `products`/`components` — those live on `feat/products-components`). The generated reference reflects whatever commands exist on this branch; the others appear automatically once merged to `main` and the site rebuilds. Use `materials list` / `materials show` as running examples in hand-written docs.

---

### Task 1: Repo hygiene — LICENSE + README (Phase 1)

**Files:**
- Create: `LICENSE`
- Create: `README.md`

**Interfaces:**
- Consumes: nothing.
- Produces: `LICENSE` and `README.md` at repo root (referenced by `.goreleaser.yml` `archives.files`).

- [ ] **Step 1: Confirm license, then create `LICENSE`**

Default to MIT (copyright `Craftybase`). If the user specified otherwise, use that. MIT text:

```
MIT License

Copyright (c) 2026 Craftybase

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

- [ ] **Step 2: Create `README.md`**

```markdown
# Craftybase CLI

The official command-line interface for the [Craftybase](https://craftybase.com) Public API — manage your inventory from the terminal.

Full docs: **https://cli.craftybase.dev**

## Install

```sh
curl -fsSL https://cli.craftybase.dev/install | bash
```

Or with Homebrew:

```sh
brew install craftybase/tap/craftybase
```

Or with Go:

```sh
go install github.com/craftybase/craftybase-cli/cmd/craftybase@latest
```

## Quickstart

```sh
craftybase auth login        # paste your API key
craftybase materials list    # list materials
craftybase materials list --json | jq '.materials[].name'
```

See [Getting Started](https://cli.craftybase.dev/getting-started/) for more.

## License

[MIT](LICENSE)
```

- [ ] **Step 3: Verify GoReleaser config resolves (if `goreleaser` is available)**

Run: `goreleaser check` (if installed) — Expected: no errors about missing `LICENSE`/`README.md`. If `goreleaser` is not installed, instead verify both files exist: `test -f LICENSE && test -f README.md && echo OK` — Expected: `OK`.

- [ ] **Step 4: Commit**

```bash
git add LICENSE README.md
git commit -m "docs: add LICENSE (MIT) and README"
```

---

### Task 2: Astro + Starlight scaffold (Phase 2)

**Files:**
- Create: `website/` (Astro+Starlight project: `package.json`, `package-lock.json`, `astro.config.mjs`, `src/content.config.ts`, `src/content/docs/`, `public/`, `tsconfig.json`, `.gitignore`)
- Modify: `.gitignore` (root) — ignore `website/node_modules` and `website/dist`

**Interfaces:**
- Consumes: nothing.
- Produces: a buildable site (`cd website && npm run build` → `website/dist/`). `website/public/` exists for static files (used by Task 3). `astro.config.mjs` exports a Starlight config whose `sidebar` is extended by later tasks.

- [ ] **Step 1: Scaffold the Starlight template into `website/`**

Run (non-interactive):
```bash
npm create astro@latest website -- --template starlight --no-install --no-git --skip-houston --yes
cd website && npm install
```
This creates the standard Starlight project. (If `npm create` refuses because `website/` exists, scaffold into a temp dir and move everything into `website/`.)

- [ ] **Step 2: Replace `website/astro.config.mjs` with the Craftybase config**

```js
// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://cli.craftybase.dev',
  integrations: [
    starlight({
      title: 'Craftybase CLI',
      description: 'The command-line interface for Craftybase.',
      social: [
        { icon: 'github', label: 'GitHub', href: 'https://github.com/craftybase/craftybase-cli' },
      ],
      sidebar: [
        { label: 'Getting Started', slug: 'getting-started' },
        { label: 'Authentication', slug: 'authentication' },
        { label: 'Output Formats', slug: 'output-formats' },
        { label: 'Configuration', slug: 'configuration' },
        { label: 'Pagination', slug: 'pagination' },
        { label: 'Command Reference', autogenerate: { directory: 'reference' } },
        { label: 'Using with Agents & LLMs', slug: 'agents' },
      ],
    }),
  ],
});
```

- [ ] **Step 3: Remove template demo content, add Pages static files**

```bash
cd website
rm -f src/content/docs/index.mdx            # replaced by Task 4's landing
rm -rf src/content/docs/guides src/content/docs/reference  # template demo dirs
rm -f src/assets/houston.webp 2>/dev/null || true
mkdir -p public src/content/docs/reference
printf 'cli.craftybase.dev\n' > public/CNAME
: > public/.nojekyll
# placeholder so the sidebar 'reference' autogenerate has ≥1 page until Task 6 runs:
printf -- '---\ntitle: Command Reference\ndescription: Placeholder; replaced by generated reference.\n---\n\nGenerated by `make docs`.\n' > src/content/docs/reference/index.md
```

Add a temporary `src/content/docs/getting-started.md`, `authentication.md`, `output-formats.md`, `configuration.md`, `pagination.md`, `agents.md` each containing only frontmatter so the sidebar `slug:` references resolve and the build passes:
```bash
for s in getting-started authentication output-formats configuration pagination agents; do
  printf -- '---\ntitle: %s\n---\n\nComing soon.\n' "$s" > "src/content/docs/$s.md"
done
```
(Real content lands in Tasks 4–8; these stubs keep the build green now.)

- [ ] **Step 4: Update the root `.gitignore`**

Append to repo-root `.gitignore`:
```
website/node_modules
website/dist
website/.astro
```

- [ ] **Step 5: Build to verify**

Run: `cd website && npm run build`
Expected: build succeeds; `website/dist/index.html` and `website/dist/CNAME` exist. Verify: `test -f dist/CNAME && test -f dist/getting-started/index.html && echo OK` → `OK`.

- [ ] **Step 6: Commit**

```bash
git add website .gitignore
git commit -m "feat(site): scaffold Astro + Starlight site"
```

---

### Task 3: Install script `/install` (Phase 1)

**Files:**
- Create: `website/public/install`
- Create: `website/public/install.sh` (identical copy)

**Interfaces:**
- Consumes: `website/public/` (from Task 2).
- Produces: a POSIX installer served at `/install`. Supports `CRAFTYBASE_DRY_RUN=1` printing `<archive>\n<asset-url>` for the resolved platform (used by its test); env overrides `CRAFTYBASE_VERSION`, `CRAFTYBASE_OS`, `CRAFTYBASE_ARCH`, `CRAFTYBASE_INSTALL_DIR`.

- [ ] **Step 1: Write the failing test**

Create `website/public/install` first as an empty file so the test can run, then write `test/install_test.sh` at repo root:

```sh
#!/bin/sh
# Verifies the installer resolves GoReleaser asset names without network I/O.
set -eu
out=$(CRAFTYBASE_DRY_RUN=1 CRAFTYBASE_OS=darwin CRAFTYBASE_ARCH=arm64 CRAFTYBASE_VERSION=v1.2.3 \
  sh website/public/install)
archive=$(printf '%s\n' "$out" | sed -n 1p)
url=$(printf '%s\n' "$out" | sed -n 2p)
[ "$archive" = "craftybase_1.2.3_darwin_arm64.tar.gz" ] || { echo "bad archive: $archive" >&2; exit 1; }
case "$url" in
  https://github.com/craftybase/craftybase-cli/releases/download/v1.2.3/craftybase_1.2.3_darwin_arm64.tar.gz) ;;
  *) echo "bad url: $url" >&2; exit 1 ;;
esac
echo OK
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `sh test/install_test.sh`
Expected: FAIL (the empty install script prints nothing → `bad archive:`).

- [ ] **Step 3: Write the installer**

Write `website/public/install`:

```sh
#!/bin/sh
# craftybase installer — https://cli.craftybase.dev/install
# Downloads the matching GoReleaser release asset, verifies its SHA-256, and
# installs the `craftybase` binary.
#
# Env overrides:
#   CRAFTYBASE_VERSION       specific version, e.g. v1.2.3 (default: latest release)
#   CRAFTYBASE_INSTALL_DIR   install dir (default: /usr/local/bin, else ~/.local/bin)
#   CRAFTYBASE_OS / _ARCH    force platform (mainly for testing)
#   CRAFTYBASE_DRY_RUN=1     print "<archive>\n<asset-url>" and exit (no download)
#
# Asset names MUST match .goreleaser.yml name_template:
#   craftybase_<version>_<os>_<arch>.tar.gz  and  craftybase_<version>_checksums.txt
set -eu

REPO="craftybase/craftybase-cli"
BINARY="craftybase"

err()  { printf 'error: %s\n' "$1" >&2; exit 1; }
info() { printf '%s\n' "$1" >&2; }
have() { command -v "$1" >/dev/null 2>&1; }

download_stdout() { if have curl; then curl -fsSL "$1"; elif have wget; then wget -qO- "$1"; else err "need curl or wget"; fi; }
download_file()   { if have curl; then curl -fsSL "$1" -o "$2"; elif have wget; then wget -qO "$2" "$1"; else err "need curl or wget"; fi; }
sha256_of()       { if have sha256sum; then sha256sum "$1" | awk '{print $1}'; elif have shasum; then shasum -a 256 "$1" | awk '{print $1}'; else err "need sha256sum or shasum"; fi; }

detect_os() {
  [ -n "${CRAFTYBASE_OS:-}" ] && { printf '%s' "$CRAFTYBASE_OS"; return; }
  case "$(uname -s)" in
    Linux)  printf 'linux' ;;
    Darwin) printf 'darwin' ;;
    *) err "unsupported OS '$(uname -s)'. On Windows, download a release from https://github.com/$REPO/releases or use WSL. See https://cli.craftybase.dev/getting-started/." ;;
  esac
}
detect_arch() {
  [ -n "${CRAFTYBASE_ARCH:-}" ] && { printf '%s' "$CRAFTYBASE_ARCH"; return; }
  case "$(uname -m)" in
    x86_64|amd64)  printf 'amd64' ;;
    arm64|aarch64) printf 'arm64' ;;
    *) err "unsupported architecture '$(uname -m)'." ;;
  esac
}
latest_version() {
  tag=$(download_stdout "https://api.github.com/repos/$REPO/releases/latest" \
        | grep '"tag_name"' | head -n1 | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
  [ -n "$tag" ] || err "could not determine latest version"
  printf '%s' "$tag"
}

OS=$(detect_os)
ARCH=$(detect_arch)
VERSION=${CRAFTYBASE_VERSION:-$(latest_version)}
VNUM=${VERSION#v}
ARCHIVE="${BINARY}_${VNUM}_${OS}_${ARCH}.tar.gz"
CHECKSUMS="${BINARY}_${VNUM}_checksums.txt"
BASE="https://github.com/$REPO/releases/download/$VERSION"
ASSET_URL="$BASE/$ARCHIVE"

if [ "${CRAFTYBASE_DRY_RUN:-}" = "1" ]; then
  printf '%s\n%s\n' "$ARCHIVE" "$ASSET_URL"
  exit 0
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
info "Downloading $ARCHIVE ($VERSION, $OS/$ARCH)…"
download_file "$ASSET_URL" "$tmp/$ARCHIVE"
download_file "$BASE/$CHECKSUMS" "$tmp/$CHECKSUMS"

info "Verifying checksum…"
want=$(grep " $ARCHIVE\$" "$tmp/$CHECKSUMS" | awk '{print $1}')
[ -n "$want" ] || err "no checksum for $ARCHIVE"
got=$(sha256_of "$tmp/$ARCHIVE")
[ "$want" = "$got" ] || err "checksum mismatch (want $want, got $got)"

tar -xzf "$tmp/$ARCHIVE" -C "$tmp"
[ -f "$tmp/$BINARY" ] || err "archive did not contain $BINARY"
chmod +x "$tmp/$BINARY"

dir=${CRAFTYBASE_INSTALL_DIR:-}
if [ -z "$dir" ]; then
  if [ -w /usr/local/bin ]; then dir=/usr/local/bin; else dir="$HOME/.local/bin"; fi
fi
mkdir -p "$dir"
mv "$tmp/$BINARY" "$dir/$BINARY"
info "Installed $BINARY to $dir/$BINARY"
case ":$PATH:" in
  *":$dir:"*) ;;
  *) info "note: $dir is not on PATH — add: export PATH=\"$dir:\$PATH\"" ;;
esac
info "Next: run '$BINARY auth login'"
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `sh test/install_test.sh`
Expected: `OK`.

- [ ] **Step 5: Lint and duplicate**

Run: `shellcheck website/public/install` — Expected: no warnings. (If shellcheck flags style-only items, fix them.)
Then duplicate for browser viewing: `cp website/public/install website/public/install.sh`.

- [ ] **Step 6: Verify the build still copies the script**

Run: `cd website && npm run build && test -f dist/install && test -f dist/install.sh && echo OK` → `OK`.

- [ ] **Step 7: Commit**

```bash
git add website/public/install website/public/install.sh test/install_test.sh
git commit -m "feat(site): add /install script with checksum verification"
```

---

### Task 4: Landing page (Phase 2)

**Files:**
- Create: `website/src/content/docs/index.mdx`

**Interfaces:**
- Consumes: Starlight (`template: splash`, `@astrojs/starlight/components`).
- Produces: the site root `/`.

- [ ] **Step 1: Write the landing page**

```mdx
---
title: Craftybase CLI
description: The command-line interface for Craftybase — manage your inventory from the terminal.
template: splash
hero:
  tagline: Manage your Craftybase inventory from the terminal. Built for humans and agents.
  actions:
    - text: Get started
      link: /getting-started/
      icon: right-arrow
    - text: View on GitHub
      link: https://github.com/craftybase/craftybase-cli
      icon: external
      variant: minimal
---

import { Card, CardGrid } from '@astrojs/starlight/components';

## Install

```sh
curl -fsSL https://cli.craftybase.dev/install | bash
```

With Homebrew:

```sh
brew install craftybase/tap/craftybase
```

With Go:

```sh
go install github.com/craftybase/craftybase-cli/cmd/craftybase@latest
```

<CardGrid>
  <Card title="Human-friendly" icon="laptop">Readable tables, color, and branded help out of the box.</Card>
  <Card title="Agent-friendly" icon="seti:json">`--json` and `--ndjson` for pipelines, scripts, and tools.</Card>
  <Card title="Authenticate once" icon="approve-check">`craftybase auth login`, then your token is stored locally.</Card>
</CardGrid>
```

- [ ] **Step 2: Build to verify**

Run: `cd website && npm run build && grep -rq "cli.craftybase.dev/install" dist/index.html && echo OK` → `OK`.

- [ ] **Step 3: Commit**

```bash
git add website/src/content/docs/index.mdx
git commit -m "feat(site): landing page"
```

---

### Task 5: Core guides (Phase 2)

**Files:**
- Modify (replace stubs): `website/src/content/docs/getting-started.md`, `authentication.md`, `output-formats.md`, `configuration.md`, `pagination.md`

**Interfaces:**
- Consumes: Starlight markdown.
- Produces: the five sidebar guide pages with accurate CLI facts.

Replace each stub with the content below. Facts are authoritative — match them to the CLI exactly.

- [ ] **Step 1: `getting-started.md`**

```markdown
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
```

- [ ] **Step 2: `authentication.md`**

```markdown
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
```

- [ ] **Step 3: `output-formats.md`**

```markdown
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
```

- [ ] **Step 4: `configuration.md`**

```markdown
---
title: Configuration
description: The config file, profiles, and the API URL.
---

## Config file

Credentials live in `~/.craftybase/config.toml`, keyed by profile (`profiles.default`), written atomically with `0600` permissions.

## API URL

The API base URL is resolved as: `--api-url` flag → `CRAFTYBASE_API_URL` env → stored profile → default (`https://api.craftybase.com`).

## Global flags

| Flag | Purpose |
| --- | --- |
| `--json` | Raw API envelope (pretty-printed) |
| `--ndjson` | Auto-paginated NDJSON stream |
| `--no-color` | Disable ANSI color |
| `--token <token>` | Override stored credentials |
| `--api-url <url>` | Override the API base URL |
| `--verbose` | Show HTTP request/response detail (token redacted) |
```

- [ ] **Step 5: `pagination.md`**

```markdown
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
```

- [ ] **Step 6: Build to verify**

Run: `cd website && npm run build && grep -rq "Token resolution precedence" dist/authentication/index.html && echo OK` → `OK`.

- [ ] **Step 7: Commit**

```bash
git add website/src/content/docs/getting-started.md website/src/content/docs/authentication.md website/src/content/docs/output-formats.md website/src/content/docs/configuration.md website/src/content/docs/pagination.md
git commit -m "docs(site): getting-started, auth, output, config, pagination guides"
```

---

### Task 6: Command-reference generator (Phase 3)

**Files:**
- Modify: `commands/root.go` (add exported `RootCmd()` accessor)
- Create: `cmd/gen-docs/main.go`
- Create: `cmd/gen-docs/main_test.go`

**Interfaces:**
- Consumes: `commands` package (the assembled `rootCmd`); `github.com/spf13/cobra/doc`.
- Produces: `commands.RootCmd() *cobra.Command`; a `gen-docs` program writing `website/src/content/docs/reference/<command>.md` with Starlight frontmatter. Pure helpers `filePrepender(filename string) string` and `linkHandler(filename string) string`.

- [ ] **Step 1: Write the failing test**

Create `cmd/gen-docs/main_test.go`:

```go
package main

import (
	"strings"
	"testing"
)

func TestFilePrepender(t *testing.T) {
	got := filePrepender("/anywhere/craftybase_materials_list.md")
	if !strings.HasPrefix(got, "---\n") {
		t.Fatalf("expected YAML frontmatter, got %q", got)
	}
	if !strings.Contains(got, "title: craftybase materials list") {
		t.Errorf("expected derived title, got %q", got)
	}
}

func TestLinkHandler(t *testing.T) {
	if got := linkHandler("craftybase_materials.md"); got != "/reference/craftybase_materials/" {
		t.Errorf("got %q, want /reference/craftybase_materials/", got)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./cmd/gen-docs/`
Expected: FAIL — `undefined: filePrepender` / `undefined: linkHandler` (build error).

- [ ] **Step 3: Add the exported accessor to `commands/root.go`**

Add (e.g. just below the `Execute` function):

```go
// RootCmd returns the assembled root command with all subcommands registered.
// Exposed for documentation generation (see cmd/gen-docs).
func RootCmd() *cobra.Command {
	return rootCmd
}
```

- [ ] **Step 4: Write `cmd/gen-docs/main.go`**

```go
// Command gen-docs renders the CLI command tree as Starlight-flavored markdown
// into website/src/content/docs/reference/. Run via `make docs`.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/craftybase/craftybase-cli/commands"
	"github.com/spf13/cobra/doc"
)

const outDir = "website/src/content/docs/reference"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	// Remove stale generated files so deletions/renames don't linger.
	entries, _ := filepath.Glob(filepath.Join(outDir, "*.md"))
	for _, f := range entries {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	root := commands.RootCmd()
	root.DisableAutoGenTag = true // deterministic output (no timestamp footer)
	return doc.GenMarkdownTreeCustom(root, outDir, filePrepender, linkHandler)
}

// filePrepender returns Starlight frontmatter for a generated reference file.
func filePrepender(filename string) string {
	base := strings.TrimSuffix(filepath.Base(filename), ".md")
	title := strings.ReplaceAll(base, "_", " ")
	return fmt.Sprintf("---\ntitle: %s\ndescription: Reference for the %s command.\n---\n\n", title, title)
}

// linkHandler maps a generated filename to its Starlight route.
func linkHandler(filename string) string {
	base := strings.TrimSuffix(filename, ".md")
	return "/reference/" + base + "/"
}
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./cmd/gen-docs/`
Expected: PASS.

- [ ] **Step 6: Generate the reference and confirm output**

Run from repo root:
```bash
rm -f website/src/content/docs/reference/index.md   # drop the Task 2 placeholder
go run ./cmd/gen-docs
ls website/src/content/docs/reference/
head -5 website/src/content/docs/reference/craftybase.md
```
Expected: one `.md` per command (`craftybase.md`, `craftybase_materials.md`, `craftybase_materials_list.md`, …), each starting with `---\ntitle: …`.

- [ ] **Step 7: Confirm the site still builds with generated reference**

Run: `cd website && npm run build && test -d dist/reference && echo OK` → `OK`.

- [ ] **Step 8: Commit**

```bash
git add commands/root.go cmd/gen-docs/ website/src/content/docs/reference/
git commit -m "feat: generate command reference markdown from Cobra (cmd/gen-docs)"
```

---

### Task 7: `make docs`, CI verify workflow (Phase 3)

**Files:**
- Create: `Makefile`
- Create: `.github/workflows/site-ci.yml`

**Interfaces:**
- Consumes: `cmd/gen-docs` (Task 6), the install script (Task 3), the site (Task 2).
- Produces: a `make docs` target and a CI job that shellchecks the installer, asserts the committed reference is in sync, and builds the site.

- [ ] **Step 1: Create `Makefile`**

```makefile
.PHONY: docs build-site test-install

docs: ## Regenerate the committed command reference
	go run ./cmd/gen-docs

build-site: ## Build the website
	cd website && npm ci && npm run build

test-install: ## Unit-test the install script (no network)
	sh test/install_test.sh
```

- [ ] **Step 2: Create `.github/workflows/site-ci.yml`**

```yaml
name: Site CI
on:
  pull_request:
    paths: ['website/**', 'cmd/gen-docs/**', 'commands/**', '.goreleaser.yml', 'test/install_test.sh', 'Makefile']
  push:
    branches: [main]
jobs:
  verify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.26'
      - name: shellcheck installer
        run: shellcheck website/public/install
      - name: install script unit test
        run: sh test/install_test.sh
      - name: command reference is in sync
        run: |
          go run ./cmd/gen-docs
          git diff --exit-code website/src/content/docs/reference
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: npm
          cache-dependency-path: website/package-lock.json
      - run: npm ci
        working-directory: website
      - run: npm run build
        working-directory: website
```

- [ ] **Step 3: Verify the sync check locally**

Run from repo root:
```bash
make docs
git diff --exit-code website/src/content/docs/reference && echo "in sync"
```
Expected: `in sync` (Task 6 already committed up-to-date output; regenerating yields no diff).

- [ ] **Step 4: Verify the install unit test target**

Run: `make test-install` → Expected: `OK`.

- [ ] **Step 5: Commit**

```bash
git add Makefile .github/workflows/site-ci.yml
git commit -m "ci: shellcheck installer, reference-sync check, site build"
```

---

### Task 8: LLM source + agents page (Phase 4)

**Files:**
- Modify: `website/astro.config.mjs` (add `starlight-llms-txt` plugin)
- Modify: `website/package.json` (add the dependency via npm)
- Modify (replace stub): `website/src/content/docs/agents.md`

**Interfaces:**
- Consumes: Starlight plugin API.
- Produces: `/llms.txt`, `/llms-full.txt` at build; an "Using with Agents & LLMs" page.

- [ ] **Step 1: Install the plugin**

Run: `cd website && npm install starlight-llms-txt`

- [ ] **Step 2: Wire the plugin into `astro.config.mjs`**

Add the import and `plugins` entry inside the `starlight({...})` call:

```js
import starlightLlmsTxt from 'starlight-llms-txt';
// ...
    starlight({
      title: 'Craftybase CLI',
      // ...existing options...
      plugins: [starlightLlmsTxt()],
      // sidebar: [ ... unchanged ... ],
    }),
```

- [ ] **Step 3: Replace `agents.md`**

```markdown
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
```

- [ ] **Step 4: Build and verify the LLM files exist**

Run: `cd website && npm run build && test -f dist/llms.txt && test -f dist/llms-full.txt && echo OK` → `OK`.

- [ ] **Step 5: Commit**

```bash
git add website/astro.config.mjs website/package.json website/package-lock.json website/src/content/docs/agents.md
git commit -m "feat(site): llms.txt LLM source + agents guide"
```

---

### Task 9: Polish — SEO/OG, 404, favicon (Phase 4)

**Files:**
- Create: `website/src/content/docs/404.md`
- Modify: `website/astro.config.mjs` (favicon + social-card/OG defaults)
- Create: `website/public/favicon.svg` (teal mark)

**Interfaces:**
- Consumes: Starlight head/favicon config.
- Produces: a 404 page, favicon, and default OG metadata.

- [ ] **Step 1: Create a teal favicon `website/public/favicon.svg`**

```svg
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="6" fill="#3EB1C1"/><text x="16" y="22" font-family="monospace" font-size="18" font-weight="700" text-anchor="middle" fill="#ffffff">c</text></svg>
```

- [ ] **Step 2: Reference favicon + OG defaults in `astro.config.mjs`**

Inside `starlight({...})` add:
```js
      favicon: '/favicon.svg',
      head: [
        { tag: 'meta', attrs: { property: 'og:image', content: 'https://cli.craftybase.dev/favicon.svg' } },
        { tag: 'meta', attrs: { name: 'theme-color', content: '#3EB1C1' } },
      ],
```

- [ ] **Step 3: Create `website/src/content/docs/404.md`**

```markdown
---
title: Page not found
description: That page doesn't exist.
template: splash
editUrl: false
---

That page doesn't exist. Head to [Getting Started](/getting-started/) or the [command reference](/reference/craftybase/).
```

- [ ] **Step 4: Build and verify**

Run: `cd website && npm run build && test -f dist/404.html && test -f dist/favicon.svg && echo OK` → `OK`.

- [ ] **Step 5: Commit**

```bash
git add website/public/favicon.svg website/astro.config.mjs website/src/content/docs/404.md
git commit -m "feat(site): favicon, OG metadata, 404 page"
```

---

### Task 10: Deploy workflow + go-live runbook (Phase 5)

**Files:**
- Create: `.github/workflows/site-deploy.yml`
- Create: `docs/superpowers/runbooks/go-live.md`

**Interfaces:**
- Consumes: the built site (`website/dist`), GitHub Pages.
- Produces: an Actions workflow that builds + deploys to Pages on pushes to `main`; a runbook of the manual go-live steps (repo public, DNS, custom domain).

- [ ] **Step 1: Create `.github/workflows/site-deploy.yml`**

```yaml
name: Deploy site
on:
  push:
    branches: [main]
    paths: ['website/**', '.github/workflows/site-deploy.yml']
  workflow_dispatch:
permissions:
  contents: read
  pages: write
  id-token: write
concurrency:
  group: pages
  cancel-in-progress: true
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: npm
          cache-dependency-path: website/package-lock.json
      - run: npm ci
        working-directory: website
      - run: npm run build
        working-directory: website
      - uses: actions/configure-pages@v5
      - uses: actions/upload-pages-artifact@v3
        with:
          path: website/dist
  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    steps:
      - id: deployment
        uses: actions/deploy-pages@v4
```

- [ ] **Step 2: Validate workflow YAML**

Run: `python3 -c "import yaml,sys; yaml.safe_load(open('.github/workflows/site-deploy.yml')); print('valid')"` → Expected: `valid`. (No deploy actually runs until the steps in the runbook are done.)

- [ ] **Step 3: Write the go-live runbook `docs/superpowers/runbooks/go-live.md`**

```markdown
# Go-live runbook — cli.craftybase.dev

These are manual, one-time steps (require repo-admin + Bunny DNS access).

1. **Merge** `feat/public-site` into `main`.
2. **Make the repo public:** GitHub → Settings → General → Danger Zone → Change visibility → Public. (Free GitHub Pages requires this.)
3. **Enable Pages via Actions:** Settings → Pages → Build and deployment → Source = **GitHub Actions**.
4. **Trigger the deploy:** push to `main` (or run the "Deploy site" workflow via *Run workflow*). Confirm it succeeds and publishes `website/dist`.
5. **Add the custom domain:** Settings → Pages → Custom domain → `cli.craftybase.dev` → Save. GitHub writes/uses the `public/CNAME` value and begins HTTPS provisioning.
6. **DNS on Bunny.net:** add a record `cli` (host) → **CNAME** → `craftybase.github.io`. Wait for propagation; GitHub will show "DNS check successful" and issue the certificate.
7. **Verify end-to-end:**
   - `https://cli.craftybase.dev/` and `/getting-started/` load over HTTPS.
   - `https://cli.craftybase.dev/reference/craftybase/` and `/llms.txt` resolve.
   - `curl -fsSL https://cli.craftybase.dev/install | bash` installs the binary from a real release (requires at least one published GoReleaser release; cut one if needed).
8. **(Optional follow-up)** repoint the CLI's "Learn more" footer: change `brand.DocsURL` to `https://cli.craftybase.dev/getting-started` (currently `https://craftybase.com/docs/api`).
```

- [ ] **Step 4: Commit**

```bash
git add .github/workflows/site-deploy.yml docs/superpowers/runbooks/go-live.md
git commit -m "ci: GitHub Pages deploy workflow + go-live runbook"
```

---

## Self-Review

**Spec coverage:**
- Same-repo `website/` Astro+Starlight, `site` no `base` → Task 2. ✓
- `/install` curl|bash, GoReleaser asset naming, SHA-256, env overrides, Windows guidance → Task 3 (+ `/install.sh`). ✓
- Landing (install one-liner, brew/go, showcase, CTAs) → Task 4. ✓
- Guides (getting-started, auth, output, config, pagination) grounded in real CLI facts → Task 5. ✓
- Cobra-generated reference, committed, `commands.RootCmd()`, `DisableAutoGenTag`, frontmatter prepender, link handler → Task 6. ✓
- `make docs` + CI sync check + shellcheck + build → Task 7. ✓
- `starlight-llms-txt` → `/llms.txt` + `/llms-full.txt`; agents page → Task 8. ✓
- Polish (favicon/OG/404) → Task 9. ✓
- GitHub Pages deploy via Actions; Bunny CNAME; repo-public gating; brand.DocsURL follow-up → Task 10 (runbook). ✓
- Hosting on Pages, DNS on Bunny, `/install` content-type caveat (no header control) → handled by serving extensionless + `/install.sh`; documented. ✓
- LICENSE + README (also unblock GoReleaser) → Task 1. ✓

**Placeholder scan:** No TBD/TODO. The Task 2 stub pages and `reference/index.md` placeholder are explicitly temporary and replaced in Tasks 4–6 (called out in-step), not plan placeholders. License default is MIT with a confirm step.

**Type/identifier consistency:** `commands.RootCmd()` defined in Task 6 and consumed by `cmd/gen-docs` in the same task; `filePrepender`/`linkHandler` names match between Task 6's test and implementation; reference output dir `website/src/content/docs/reference` consistent across Tasks 2, 6, 7; install env vars (`CRAFTYBASE_DRY_RUN`/`_OS`/`_ARCH`/`_VERSION`) consistent between Task 3's script and its test; asset name format consistent with the Global Constraints and `.goreleaser.yml`.

**Note on ordering:** tasks are sequenced by build dependency (scaffold before the install script that lives in `website/public/`, generator before its CI sync check), so the phase numbers in headings are logical groupings rather than strict execution order.
