---
name: cut-release
description: Cut a new tagged release of the stocksmith CLI — pre-flight green check, semver version decision, tag + push (which triggers GoReleaser), and post-publish verification of the GitHub release assets and Homebrew tap. Use when the user asks to cut, ship, publish, or tag a new version / release.
---

# Cut a Release

## Overview

Releases here are **tag-triggered**. Pushing a `vX.Y.Z` tag fires
`.github/workflows/release.yml`, which runs GoReleaser (`.goreleaser.yml`):
it cross-compiles the archives, publishes the GitHub release with
auto-generated notes, and updates the Homebrew tap formula. There is **no
manual "publish" button** — pushing the tag *is* the release.

This skill drives that end to end, safely: it refuses to tag a red commit,
proposes the right semver bump, confirms once, then tags/pushes/watches/verifies.

**Announce at start:** "I'm using the cut-release skill to cut a new version."

## Iron rules

- **Never tag a red commit.** Confirm the target commit's CI is fully green first.
- **One confirmation gate.** Get the user's OK on the version (and the green
  status) once, *before* pushing the tag. After that, run unattended through
  verification.
- **The tag push is irreversible-ish.** It publishes a public release and
  mutates the Homebrew tap that real users install from. Treat it like a
  deploy, not a commit.
- **Don't hardcode brand strings.** Owner/repo come from the git remote
  (`gh` resolves `:owner/:repo` from the cwd). The tap repo and binary name
  come from `.goreleaser.yml`. This repo is white-label-capable.

## Steps

### 1. Pre-flight

```bash
git fetch origin --tags
```

- Decide the **target ref** — almost always `origin/main`. Confirm with the
  user if they mean a different branch. Record the commit SHA you'll tag.
- **Confirm CI is green on that exact commit** (build/lint/test/verify all
  `success`). Never tag red:

  ```bash
  gh api repos/:owner/:repo/commits/<sha>/check-runs \
    --jq '.check_runs[] | "\(.name): \(.status) → \(.conclusion // "pending")"' | sort -u
  ```

  If anything is pending, wait. If anything failed, STOP and report — do not tag.
- Find the **last released tag** and what's shipped:

  ```bash
  git describe --tags --abbrev=0     # e.g. v0.2.0
  gh release list --limit 5
  ```

### 2. Decide the version (suggest, then confirm — the gate)

List the commits since the last tag and infer the bump from conventional
prefixes:

```bash
git log --oneline <last-tag>..origin/main
```

- any `feat:` (or a new command/user-facing capability) → **minor** bump (`X.(Y+1).0`)
- only `fix:` / `chore:` / `docs:` / `refactor:` / `test:` → **patch** bump (`X.Y.(Z+1)`)
- `feat!:` / a `BREAKING CHANGE:` footer → **major** bump (`(X+1).0.0`)

While pre-1.0 (`0.y.z`), still treat a feature as a minor bump (`0.2.0 → 0.3.0`)
and a fix as a patch (`0.2.0 → 0.2.1`). A new command is a **minor**, not a patch.

**Present the proposed version to the user with the commit list and the green
status, and get explicit confirmation.** This is the only gate. Everything
after runs unattended unless something fails.

### 3. Tag + push (after confirmation)

Make sure the tag is free, then create an annotated tag on the target commit
and push it:

```bash
git rev-parse "vX.Y.Z" 2>/dev/null && echo "tag exists — STOP" || \
  git tag -a "vX.Y.Z" -m "vX.Y.Z — <one-line summary of the headline change>" <sha>
git push origin "vX.Y.Z"
```

The push triggers `release.yml`. (Tags are pushed explicitly by name — a plain
`git push` does not push tags.)

### 4. Watch the release run

```bash
sleep 5
gh run list --workflow=release.yml --limit 3 \
  --json databaseId,headBranch,status,conclusion \
  --jq '.[] | "\(.databaseId) | \(.headBranch) | \(.status) → \(.conclusion // "running")"'
gh run watch <run-id> --exit-status
```

If the run **fails**, STOP, surface the failing job's logs
(`gh run view <run-id> --log-failed`), and report — do not proceed to verify.
(A benign "Node 20 is deprecated → forced to Node 24" annotation is expected
and is not a failure.)

### 5. Verify the publish

- **Release assets** — all platform archives + the checksums file are present:

  ```bash
  gh release view "vX.Y.Z" --json tagName,assets --jq '.tagName, (.assets[].name)'
  ```

  Expect the 5 GoReleaser archives (`darwin_amd64/arm64`, `linux_amd64/arm64`,
  `windows_amd64`) plus `*_checksums.txt`.

- **Homebrew tap bumped** — the formula's `version` and download URLs match the
  new tag. The tap repo is defined in `.goreleaser.yml` under
  `brews.repository` (currently `craftybase/homebrew-tap`, formula at the repo
  root):

  ```bash
  gh api repos/<tap-owner>/<tap-repo>/contents/stocksmith.rb \
    --jq '.content' | base64 -d | grep -E 'version|url' | head
  ```

- **Optional install smoke test** — prove the shipped binary actually runs and
  reports the new version (download the matching platform archive to a temp
  dir, extract, run `version`):

  ```bash
  D=$(mktemp -d); gh release download "vX.Y.Z" --repo <owner>/<repo> \
    -p "stocksmith_<X.Y.Z>_darwin_arm64.tar.gz" -D "$D"
  tar -xzf "$D"/stocksmith_*.tar.gz -C "$D" && "$D/stocksmith" version
  ```

### 6. Report

Summarize: release URL (`gh release view vX.Y.Z --json url`), the version,
the published assets, and the tap status. Note any follow-ups (e.g. "users on
an older install upgrade via `brew upgrade stocksmith`, `stocksmith update`,
or re-running the install script").

## Gotchas (hard-won)

- **Stale binary shadowing after release.** A user (or you) may have an old
  `~/.local/bin/stocksmith` that shadows the brew copy on `PATH`, so
  `stocksmith version` still shows the old version. Diagnose with
  `which -a stocksmith`; the fix is to remove the stale script-install copy or
  re-run the installer / `stocksmith update`.
- **The Homebrew tap must actually bump.** GoReleaser's `brews:` block pushes
  the formula to the tap repo using `HOMEBREW_TAP_TOKEN`. If the formula didn't
  update, check that secret and the release-run logs — the GitHub release can
  succeed while the tap push fails.
- **`brews` is deprecated** in GoReleaser (migrating to `homebrew_casks` is a
  separate, deliberate change — see the project memory). It still works
  (warning only); don't "fix" it as part of cutting a release.
- **A failed release is usually a build/config problem, not a tag problem.**
  If GoReleaser can't find the entrypoint or an archive, fix `.goreleaser.yml`
  / the source on `main`, then re-tag (delete and re-push the tag, or cut the
  next patch). Don't keep force-pushing a broken tag blindly.
