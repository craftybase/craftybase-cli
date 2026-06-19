# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

`craftybase` is a Go CLI for the Craftybase Public API. Module path: `github.com/craftybase/craftybase-cli` (Go 1.26).

## Commands

```bash
go build -o craftybase ./cmd/craftybase   # build the binary (gitignored)
go run ./cmd/craftybase materials list     # run without building
go test ./...                              # all tests
go test ./... -v -race                     # tests as CI runs them
go test ./commands -run TestRenderRootHelp_PlainContent   # single test
go vet ./...
golangci-lint run                          # lint (config in .golangci.yml)
```

CI (`.github/workflows/ci.yml`) runs `go test ./... -v -race`, `go vet ./...`, and golangci-lint (enables errcheck, gosimple, govet, ineffassign, staticcheck, unused, gofmt, goimports; goimports local-prefix is the module path). Releases are built by GoReleaser (`.goreleaser.yml`): cross-compiles darwin/linux/windows, publishes a Homebrew tap, and injects version info via ldflags into `main.version/commit/buildDate`.

## Architecture

Three layers:

- **`cmd/craftybase/main.go`** — tiny entry point. Receives ldflag-injected version vars, calls `commands.SetVersion(...)` then `commands.Execute(...)`.
- **`commands/`** (package `commands`) — Cobra command definitions, one file per command group (`root`, `auth`, `account`, `api`, `materials`, `version`, `completion`, `roothelp`). Each registers itself onto the shared `rootCmd` in its own `init()`.
- **`internal/`** — reusable libraries: `brand`, `config`, `api`, `output`.

### How a data command works (the standard flow)

Every command that hits the API follows the same shape — match it:

1. `client, _, err := requireAuth()` (or `resolveClient()` if auth is optional). These live in `commands/root.go` and assemble an `api.Client` from resolved token + base URL.
2. Build an `http.Request` against `client.BaseURL + "/api/v1/..."`.
3. Call `client.Do(req)` — **not** the raw HTTP client. `Do` handles retries, backoff, 429 waits, and error→`APIError` mapping.
4. Branch on the global output flags: `flagJSON` → `output.PrintJSON` (whole envelope); `flagNDJSON` → `api.WalkPages` + `output.WriteNDJSONLine`; default → unmarshal the envelope into a typed struct and `output.FormatTable`.

### Key conventions (the non-obvious parts)

- **Brand constants are centralized in `internal/brand`.** Do not hardcode `"craftybase"`, the config path, env var names, or the default API URL — reference `brand.BinaryName`, `brand.ConfigDir`, `brand.EnvTokenName`, `brand.DefaultAPIURL`, etc. The codebase is deliberately white-label capable.
- **API responses are enveloped**, e.g. `{"account": {...}}` and `{"materials": [...], "meta": {...}}`. Unmarshal into an anonymous struct keyed by the envelope name; `--json` prints the full envelope, table mode extracts the inner object(s).
- **Money is a string pair**, `{"amount": "8.75", "currency_code": "USD"}` — never a float. Use `output.Money` / `output.FormatMoney`. There are "contract check" tests asserting fixtures keep this shape (and that `category` stays a flat string).
- **Output-mode flags are global persistent flags** on `rootCmd` (`--json`, `--ndjson`, `--no-color`, `--verbose`, `--token`, `--api-url`), read directly as package-level `flag*` vars. `--json`/`--ndjson` are mutually exclusive (enforced in `PersistentPreRunE`).
- **Token / API-URL resolution precedence** (in `internal/config`): flag → env var (`CRAFTYBASE_API_TOKEN` / `CRAFTYBASE_API_URL`) → stored profile → default. Config is TOML at `~/.craftybase/config.toml`, profile-keyed (`profiles.default`), written atomically (temp file + rename) with enforced `0600` perms.
- **Errors return `*api.APIError` from `RunE`.** `Execute()` maps status codes to process exit codes via `api.ExitCode` (401→3, 404→4, others→1). `rootCmd` sets `SilenceUsage`/`SilenceErrors` so Cobra does not double-print; `Execute` prints the error itself.
- **Color/TTY**: gate any styled output on `output.ColorEnabled(flagNoColor)`, which honors `--no-color`, `NO_COLOR`, and non-TTY stdout. `output.Styler` renders 24-bit truecolor with an automatic 256-color fallback; the brand palette lives in `internal/output/style.go`.
- **Pagination**: `--all` and `--ndjson` use `api.WalkPages`, which reads `meta.total_pages` from page 1 and walks the rest with per-page retry/backoff. `--all` is mutually exclusive with both `--ndjson` and `--page`.

### Branded root help (`commands/roothelp.go`)

`renderRootHelp` is a pure, fully unit-tested renderer (takes the root `*cobra.Command` + injected `renderOpts`). It is wired into Cobra in `root.go` via `SetHelpFunc` (branded screen for the root command; default Cobra help delegated for every subcommand) plus a root `Run` for the no-args invocation; `resolveRenderOpts` resolves color/truecolor/terminal-width from the real stdout at call time. Column widths in the Flags and Environment Variables sections are derived from the longest entry (`columnWidth`), not hardcoded. Two invariant tests enforce coverage: `TestCommandRowsCoverAllCommands` (every visible root command must appear in `commandRows`) and `TestFlagRowsCoverPersistentFlags` (every persistent flag must appear in `flagRows`). **If you add a top-level command or a persistent flag, update `roothelp.go` or these tests fail.**

## Testing

- Tests mock the API with `httptest.NewServer` + route maps (`setupMockServer`); no real network calls. Fixtures are inline JSON helpers.
- The `commands` package mixes internal tests (`package commands`, e.g. `roothelp_test.go`) and external tests (`package commands_test`, e.g. `commands_test.go`).
- `testdata/golden/` holds golden files; `*.got` outputs are gitignored.
