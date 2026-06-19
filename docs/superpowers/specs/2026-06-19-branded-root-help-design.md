# Branded Root Help Screen — Design

**Date:** 2026-06-19
**Status:** Approved (pending spec review)
**Scope:** Root landing/help screen only

## Goal

Replace the default Cobra help output shown for the root command with a branded,
Sentry-style landing screen: an ASCII logo, a tagline, a grouped command list with
right-aligned descriptions, a Flags section, an Environment Variables section, a
`try:` hint, and a "Learn more" footer — all in the Craftybase teal palette, degrading
gracefully on non-color / non-TTY / narrow terminals.

This is purely presentational. No command behavior, flags, or API calls change.

## Scope

**In scope** — the screen rendered when the user runs:
- `craftybase` (no arguments)
- `craftybase --help` / `craftybase -h`
- `craftybase help`

**Out of scope** (unchanged, keep default Cobra rendering):
- Subcommand help (e.g. `craftybase materials --help`)
- Data output (tables, JSON, NDJSON)
- Error/usage output on bad input

## The Logo

Figlet-style uppercase wordmark (matches the existing Craftybase logo aesthetic),
48 columns × 3 rows, rendered in **flat bright teal** (`#3EB1C1`):

```
 __   __        ___ ___      __        __   ___
/  ` |__)  /\  |__   |  \ / |__)  /\  /__` |__
\__, |  \ /~~\ |     |   |  |__) /~~\ .__/ |___
```

The art is stored as a constant (a raw string literal so backslashes are preserved).

## Layout

```
<logo, flat teal>

The command-line interface for Craftybase          <- tagline, bold

  $ craftybase help <command>                  Display help for a command
  $ craftybase account                         Show account information
  $ craftybase api <METHOD> <path>             Make authenticated API requests
  $ craftybase auth login | status | logout    Authenticate with Craftybase
  $ craftybase materials list | show           Manage materials
  $ craftybase completion <shell>              Generate shell completion scripts
  $ craftybase version                         Print version information

Flags:
      --json            Output raw API envelope (pretty-printed JSON)
      --ndjson          Output auto-paginated NDJSON stream
      --token <token>   API token (overrides stored credentials)
      --api-url <url>   API base URL
      --no-color        Disable ANSI color output
      --verbose         Show HTTP request/response detail (token redacted)
  -h, --help            Show help for a command

Environment Variables:
  CRAFTYBASE_API_TOKEN  API token used for requests (CI, scripts)
  CRAFTYBASE_API_URL    API base URL override
  NO_COLOR              Disable colored output (no-color.org convention)

try: craftybase materials list

Learn more at https://craftybase.com/docs/api
```

- The description column in the command list is right-padded to a fixed start column
  (~48) computed from the visible (ANSI-stripped) length of the left side.
- Tagline, footer URL, and `try:` example are fixed strings (below).

## Color Palette

Adapted from the Craftybase brand (tailwind `cb-bright-blue` teal + warm terracotta).
Primary path uses 24-bit truecolor.

| Element | Color | Hex | RGB |
|---|---|---|---|
| Logo, `craftybase` + command names, flag names, env var names, footer URL | bright teal | `#3EB1C1` | 62,177,193 |
| Subcommand tokens (`login`, `list`, …) | light teal | `#65C1CD` | 101,193,205 |
| `$` prompt, `\|` separators, `<args>` placeholders | dim gray | `#6F6F6F` | 111,111,111 |
| Descriptions (right column, flags, env vars) | gray | `#837f7f` | 131,127,127 |
| `try:` example command | terracotta | `#C48D81` | 196,141,129 |
| Tagline, section headers (`Flags:`, `Environment Variables:`) | bold (default fg) | — | — |
| Footer URL | bright teal + underline | `#3EB1C1` | 62,177,193 |

## Rendering Modes & Degradation

Decision inputs: `colorEnabled` (existing `output.ColorEnabled(flagNoColor)`, which already
honors `NO_COLOR` and TTY), truecolor support, and terminal width.

1. **Full color** — stdout is a TTY, color enabled, truecolor supported:
   logo (flat teal) + full palette as above.
2. **256-color** — color enabled but truecolor not detected
   (`COLORTERM` not `truecolor`/`24bit`): render with nearest 256-color approximations
   of teal / gray / terracotta so the screen stays on-brand. (Exact 256 codes finalized
   in implementation.)
3. **No color** — color disabled (`--no-color`, `NO_COLOR`, or not a TTY / piped):
   plain text, no ANSI, but the **logo is still shown** if width allows (it is plain ASCII).
   When not a TTY, width is treated as wide enough.
4. **Narrow terminal** — detected width < logo width (48 cols): omit the logo, show a plain
   bold `Craftybase` heading + tagline instead; the rest of the screen renders normally.

Width is obtained via `term.GetSize(int(os.Stdout.Fd()))` (the `golang.org/x/term`
dependency is already used by `commands/auth.go`). If size lookup fails, assume wide.

## Data Sources (stay in sync with the command tree)

The command list is **derived from the live Cobra tree**, not hand-duplicated, so adding a
command surfaces it automatically:

- Iterate `rootCmd.Commands()`, skipping commands marked `Hidden`.
- **Name** and **description** come from each command's `Use`/`Short`.
- **Subcommands** come from each command's immediate children's names, joined by ` | `.
- A small **override table** handles special cases:
  - `completion` is **shown** as a row, but its args are displayed as `<shell>` rather than
    expanding its children (`bash | zsh | fish | powershell`).
  - controls **display order** to match the layout above.
  - `help` is shown first with a `<command>` arg.
- A unit test asserts every non-hidden root command appears in the rendered screen, so a new
  command added without consideration fails CI rather than being silently omitted.

The **flag rows are a curated table** (the layout above is the source of truth for which
flags appear, their wording, and order — we do not auto-derive them, to keep `<token>`/`<url>`
arg hints and grouping under our control). A unit test asserts the curated list covers every
persistent flag registered on `rootCmd`, so a newly added persistent flag fails CI until listed.

Environment variables, tagline, `try:` example, and footer URL are constants:
- Tagline: `The command-line interface for Craftybase`
- Env vars: `CRAFTYBASE_API_TOKEN`, `CRAFTYBASE_API_URL`, `NO_COLOR`
  (names sourced from `internal/brand` where they already exist)
- `try:` example: `craftybase materials list`
- Footer: `Learn more at https://craftybase.com/docs/api`

## Implementation Approach

Chosen: **`rootCmd.SetHelpFunc` + a root `Run`**, with styling primitives in
`internal/output` and screen composition alongside the commands.

Rejected alternatives:
- Cobra `SetUsageTemplate` (Go `text/template`): per-character/gradient-free flat color is
  doable but TTY/width/truecolor branching and ANSI in templates get ugly and untestable.
- Intercepting `os.Args` before Cobra: fragile, duplicates arg parsing.

### Files

**New: `internal/output/style.go`**
- Brand palette constants (RGB triples / hex).
- `SupportsTrueColor() bool` — inspects `COLORTERM`.
- Helpers: foreground-color, bold, underline wrappers that **no-op when color is disabled**,
  and a truecolor→256 fallback for the brand colors.
- Keeps existing `color.go` (`IsTTY`, `ColorEnabled`) and `table.go` (`bold`/`dim`) intact;
  may consolidate the ad-hoc `bold`/`dim` here later (not required for this change).

**New: `commands/roothelp.go`**
- `renderRootHelp(cmd *cobra.Command, w io.Writer, opts renderOpts)` — composes the screen.
  `renderOpts` carries `color bool`, `trueColor bool`, `width int` so tests can inject them.
- The logo constant, command override table, env-var list, tagline, `try:`, footer.
- Wiring (in `init()` or an exported setup called from `root.go`):
  ```go
  defaultHelp := rootCmd.HelpFunc()
  rootCmd.SetHelpFunc(func(c *cobra.Command, args []string) {
      if c == rootCmd {
          renderRootHelp(c, c.OutOrStdout(), resolveRenderOpts())
          return
      }
      defaultHelp(c, args) // subcommands keep default Cobra help
  })
  rootCmd.Run = func(c *cobra.Command, args []string) {
      renderRootHelp(c, c.OutOrStdout(), resolveRenderOpts())
  }
  ```
- `resolveRenderOpts()` reads `flagNoColor`, `output.ColorEnabled`, `output.SupportsTrueColor`,
  and terminal width.

**Modified: `commands/root.go`**
- Call the help-func/Run wiring during `init()` (or `Execute`).
- `rootCmd.Long` description becomes unused for the root screen (kept for safety / `help`
  fallback but not displayed by the branded renderer).

## Testing

Match the existing style (`bytes.Buffer` capture + content/ANSI assertions; tests live in
`commands` and `internal/output`). Render via `renderRootHelp` with injected `renderOpts`:

1. **Plain mode** (`color:false`, wide): exact-content assertions — logo present, all command
   names present, all section headers, env vars, `try:`, footer; assert **no** `\033[` sequences.
2. **Color mode** (`color:true, trueColor:true`, wide): assert teal escape for the logo and
   command names, terracotta for `try:`, underline for the URL.
3. **Narrow mode** (`width:40`): assert logo art is **absent** and the bold `Craftybase`
   heading is present.
4. **Command-tree sync**: assert every non-hidden `rootCmd.Commands()` entry's name appears in
   the plain-mode output (guards against forgetting to surface a new command).
5. **256-color fallback** (`color:true, trueColor:false`): assert ANSI present and no 24-bit
   (`38;2;`) sequences.

Optionally store the plain-mode screen as a golden file under `testdata/golden/` (directory
already exists) for a full-screen regression snapshot.

## Edge Cases

- **No subcommand + flags only** (e.g. `craftybase --json`): still shows the branded screen
  (root `Run` fires). Acceptable.
- **`help` subcommand for a child** (`craftybase help materials`): routes through the help
  func with `c != rootCmd` → default Cobra help. Correct.
- **Output redirected** (`craftybase > file`): not a TTY → no color; logo still written as
  plain ASCII (consistent with `--help | less`).
- **Very wide flag/desc lines**: the longest natural line is ~82 cols; no special handling —
  matches Sentry, which also lets the screen assume a reasonably wide terminal.

## Non-Goals / YAGNI

- No theming / configurable palette.
- No animation or spinners.
- No restyling of subcommand help or data output (explicitly out of scope).
- No new dependencies (`golang.org/x/term` already present).
