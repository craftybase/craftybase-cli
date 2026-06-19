# Branded Root Help Screen Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the default Cobra root help with a branded, Sentry-style landing screen (teal `CRAFTYBASE` logo, grouped command list, Flags/Env sections, `try:` hint, docs footer) that degrades gracefully on non-color/non-TTY/narrow terminals.

**Architecture:** A reusable ANSI styling layer in `internal/output` (palette + truecolor/256/no-color rendering). A pure screen renderer in `commands/roothelp.go` that takes the root `*cobra.Command` plus injected render options (so it is fully unit-testable). Wiring in `commands/root.go` that points Cobra's root help func and root `Run` at the renderer while leaving subcommand help untouched.

**Tech Stack:** Go 1.26, `github.com/spf13/cobra`, `github.com/spf13/pflag`, `golang.org/x/term`. Standard library: `fmt`, `io`, `os`, `strings`, `unicode/utf8`.

## Global Constraints

- Module path: `github.com/craftybase/craftybase-cli`.
- Scope is the **root help screen only** — `craftybase` (no args), `craftybase --help`/`-h`, `craftybase help`. Subcommand help and all data output stay as default Cobra.
- No new third-party dependencies. `golang.org/x/term` is already a direct dep; `github.com/spf13/pflag` is currently indirect and becomes direct (resolve with `go mod tidy`, which is offline — it is already in `go.sum`).
- Color must be suppressed when `output.ColorEnabled(flagNoColor)` is false (this already honors `--no-color`, `NO_COLOR`, and non-TTY stdout).
- Brand palette (24-bit), copied verbatim:
  - bright teal `#3EB1C1` = RGB(62,177,193)
  - light teal `#65C1CD` = RGB(101,193,205)
  - dim gray `#6F6F6F` = RGB(111,111,111)
  - gray `#837f7f` = RGB(131,127,127)
  - terracotta `#C48D81` = RGB(196,141,129)
- Fixed copy strings:
  - tagline: `The command-line interface for Craftybase`
  - try example: `craftybase materials list`
  - footer: `Learn more at https://craftybase.com/docs/api`
- Existing helpers to reuse, do not duplicate: `output.IsTTY()`, `output.ColorEnabled(bool)`, `brand.EnvTokenName` (`CRAFTYBASE_API_TOKEN`), `brand.EnvAPIURL` (`CRAFTYBASE_API_URL`).

---

### Task 1: ANSI styling primitives

**Files:**
- Create: `internal/output/style.go`
- Test: `internal/output/style_test.go`

**Interfaces:**
- Consumes: nothing (new leaf module).
- Produces:
  - `type RGB struct { R, G, B uint8 }`
  - Palette vars: `TealBright, TealLight, GrayDim, Gray, Terracotta RGB`
  - `func SupportsTrueColor() bool`
  - `type Styler struct { Color bool; TrueColor bool }`
  - `func (s Styler) Fg(c RGB, text string) string`
  - `func (s Styler) Bold(text string) string`
  - `func (s Styler) Underline(text string) string`
  - When `Color` is false, all three return `text` unchanged. When `Color` is true and `TrueColor` is false, `Fg` emits a 256-color (`\033[38;5;Nm`) sequence; otherwise 24-bit (`\033[38;2;R;G;Bm`).

- [ ] **Step 1: Write the failing test**

Create `internal/output/style_test.go`:

```go
package output_test

import (
	"strings"
	"testing"

	"github.com/craftybase/craftybase-cli/internal/output"
)

func TestStyler_ColorDisabled_NoANSI(t *testing.T) {
	s := output.Styler{Color: false, TrueColor: true}
	if got := s.Fg(output.TealBright, "hello"); got != "hello" {
		t.Errorf("Fg with color off should be plain, got %q", got)
	}
	if got := s.Bold("hi"); got != "hi" {
		t.Errorf("Bold with color off should be plain, got %q", got)
	}
	if got := s.Underline("hi"); got != "hi" {
		t.Errorf("Underline with color off should be plain, got %q", got)
	}
}

func TestStyler_TrueColor_Emits24Bit(t *testing.T) {
	s := output.Styler{Color: true, TrueColor: true}
	got := s.Fg(output.TealBright, "X")
	if !strings.Contains(got, "\033[38;2;62;177;193m") {
		t.Errorf("expected 24-bit teal sequence, got %q", got)
	}
	if !strings.HasSuffix(got, "\033[0m") {
		t.Errorf("expected reset suffix, got %q", got)
	}
	if !strings.Contains(got, "X") {
		t.Errorf("expected text preserved, got %q", got)
	}
}

func TestStyler_NoTrueColor_Emits256(t *testing.T) {
	s := output.Styler{Color: true, TrueColor: false}
	got := s.Fg(output.TealBright, "X")
	if strings.Contains(got, "38;2;") {
		t.Errorf("expected no 24-bit sequence in 256 mode, got %q", got)
	}
	if !strings.Contains(got, "\033[38;5;") {
		t.Errorf("expected 256-color sequence, got %q", got)
	}
}

func TestStyler_BoldUnderline_WhenColor(t *testing.T) {
	s := output.Styler{Color: true, TrueColor: true}
	if got := s.Bold("X"); got != "\033[1mX\033[0m" {
		t.Errorf("unexpected bold sequence: %q", got)
	}
	if got := s.Underline("X"); got != "\033[4mX\033[0m" {
		t.Errorf("unexpected underline sequence: %q", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/output/ -run TestStyler -v`
Expected: FAIL — `undefined: output.Styler` / `output.TealBright` (compile error).

- [ ] **Step 3: Write minimal implementation**

Create `internal/output/style.go`:

```go
package output

import (
	"fmt"
	"os"
)

// RGB is a 24-bit color.
type RGB struct{ R, G, B uint8 }

// Craftybase brand palette.
var (
	TealBright = RGB{62, 177, 193}  // #3EB1C1
	TealLight  = RGB{101, 193, 205} // #65C1CD
	GrayDim    = RGB{111, 111, 111} // #6F6F6F
	Gray       = RGB{131, 127, 127} // #837f7f
	Terracotta = RGB{196, 141, 129} // #C48D81
)

// SupportsTrueColor reports whether the terminal advertises 24-bit color.
func SupportsTrueColor() bool {
	ct := os.Getenv("COLORTERM")
	return ct == "truecolor" || ct == "24bit"
}

// Styler renders ANSI styling, honoring color and truecolor capabilities.
type Styler struct {
	Color     bool
	TrueColor bool
}

// Fg wraps text in a foreground color (24-bit when available, else 256-color).
func (s Styler) Fg(c RGB, text string) string {
	if !s.Color {
		return text
	}
	if s.TrueColor {
		return fmt.Sprintf("\033[38;2;%d;%d;%dm%s\033[0m", c.R, c.G, c.B, text)
	}
	return fmt.Sprintf("\033[38;5;%dm%s\033[0m", to256(c), text)
}

// Bold wraps text in the bold attribute.
func (s Styler) Bold(text string) string {
	if !s.Color {
		return text
	}
	return "\033[1m" + text + "\033[0m"
}

// Underline wraps text in the underline attribute.
func (s Styler) Underline(text string) string {
	if !s.Color {
		return text
	}
	return "\033[4m" + text + "\033[0m"
}

// to256 approximates an RGB color with an xterm-256 palette index.
func to256(c RGB) int {
	if c.R == c.G && c.G == c.B {
		switch {
		case c.R < 8:
			return 16
		case c.R > 248:
			return 231
		default:
			return 232 + (int(c.R)-8)*24/247
		}
	}
	r := int(c.R) * 5 / 255
	g := int(c.G) * 5 / 255
	b := int(c.B) * 5 / 255
	return 16 + 36*r + 6*g + b
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/output/ -run TestStyler -v`
Expected: PASS (all four tests).

- [ ] **Step 5: Commit**

```bash
git add internal/output/style.go internal/output/style_test.go
git commit -m "feat: add ANSI styling primitives for branded output (CU-868k0t6gb)"
```

---

### Task 2: Root help screen renderer

**Files:**
- Create: `commands/roothelp.go`
- Test: `commands/roothelp_test.go`
- Modify: `go.mod` / `go.sum` (via `go mod tidy` — pflag becomes a direct dep)

**Interfaces:**
- Consumes (from Task 1): `output.Styler`, `output.RGB`, palette vars, `output.SupportsTrueColor`.
- Produces:
  - `type renderOpts struct { color bool; trueColor bool; width int }` (width ≤ 0 means "unknown / assume wide").
  - `func renderRootHelp(root *cobra.Command, w io.Writer, opts renderOpts)` — writes the full screen to `w`.
  - `const rootLogo string`, `func logoWidth() int`.
  - Tables `commandRows`, `flagRows`, `envVarRows` and constants `tagline`, `tryExample`, `docsURL`.
- Note: the test file uses `package commands` (internal) so it can reach these unexported names and the package's `rootCmd`. This coexists with the existing external `package commands_test` file in the same directory.

- [ ] **Step 1: Write the failing test**

Create `commands/roothelp_test.go`:

```go
package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

func renderToString(opts renderOpts) string {
	var buf bytes.Buffer
	renderRootHelp(rootCmd, &buf, opts)
	return buf.String()
}

func TestRenderRootHelp_PlainContent(t *testing.T) {
	out := renderToString(renderOpts{color: false, trueColor: false, width: 100})

	// no ANSI in plain mode
	if strings.Contains(out, "\033[") {
		t.Errorf("plain mode must not contain ANSI escapes")
	}
	// logo present
	if !strings.Contains(out, rootLogo) {
		t.Errorf("expected logo art in output")
	}
	// tagline, sections, footer
	for _, want := range []string{
		"The command-line interface for Craftybase",
		"$ craftybase materials list | show",
		"$ craftybase auth login | status | logout",
		"Flags:",
		"--json",
		"Environment Variables:",
		"CRAFTYBASE_API_TOKEN",
		"NO_COLOR",
		"try: craftybase materials list",
		"Learn more at https://craftybase.com/docs/api",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q", want)
		}
	}
}

func TestRenderRootHelp_ColorUsesTeal(t *testing.T) {
	out := renderToString(renderOpts{color: true, trueColor: true, width: 100})
	if !strings.Contains(out, "\033[38;2;62;177;193m") {
		t.Errorf("expected bright-teal (62;177;193) somewhere in colored output")
	}
	if !strings.Contains(out, "\033[38;2;196;141;129m") {
		t.Errorf("expected terracotta (196;141;129) for the try: command")
	}
	if !strings.Contains(out, "\033[4m") {
		t.Errorf("expected underline for the footer URL")
	}
}

func TestRenderRootHelp_NarrowDropsLogo(t *testing.T) {
	out := renderToString(renderOpts{color: false, trueColor: false, width: 40})
	if strings.Contains(out, rootLogo) {
		t.Errorf("logo should be omitted on a narrow terminal")
	}
	if !strings.Contains(out, "Craftybase") {
		t.Errorf("expected plain Craftybase heading fallback")
	}
	if !strings.Contains(out, "The command-line interface for Craftybase") {
		t.Errorf("tagline should still appear when logo is dropped")
	}
}

func TestRenderRootHelp_256Fallback(t *testing.T) {
	out := renderToString(renderOpts{color: true, trueColor: false, width: 100})
	if strings.Contains(out, "38;2;") {
		t.Errorf("256 mode must not emit 24-bit sequences")
	}
	if !strings.Contains(out, "\033[38;5;") {
		t.Errorf("expected 256-color sequences when truecolor unavailable")
	}
}

func TestCommandRowsCoverAllCommands(t *testing.T) {
	listed := map[string]bool{}
	for _, r := range commandRows {
		listed[r.name] = true
	}
	for _, c := range rootCmd.Commands() {
		if c.Hidden {
			continue
		}
		if !listed[c.Name()] {
			t.Errorf("command %q is not listed in commandRows (root help screen)", c.Name())
		}
	}
}

func TestFlagRowsCoverPersistentFlags(t *testing.T) {
	var joined strings.Builder
	for _, f := range flagRows {
		joined.WriteString(f.name)
		joined.WriteString("\n")
	}
	rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if !strings.Contains(joined.String(), "--"+f.Name) {
			t.Errorf("persistent flag --%s missing from flagRows", f.Name)
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./commands/ -run 'TestRenderRootHelp|TestCommandRowsCover|TestFlagRowsCover' -v`
Expected: FAIL — `undefined: renderOpts` / `renderRootHelp` / `rootLogo` / `commandRows` / `flagRows` (compile error). Also expect a missing-import error for `pflag` until `go mod tidy` runs in Step 4.

- [ ] **Step 3: Write minimal implementation**

Create `commands/roothelp.go`:

```go
package commands

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/craftybase/craftybase-cli/internal/brand"
	"github.com/craftybase/craftybase-cli/internal/output"
	"github.com/spf13/cobra"
)

// rootLogo is the flat-teal CRAFTYBASE wordmark (48x3). Raw segments are joined
// around literal backticks so backslashes stay literal.
const rootLogo = ` __   __        ___ ___      __        __   ___
/  ` + "`" + ` |__)  /\  |__   |  \ / |__)  /\  /__` + "`" + ` |__  
\__, |  \ /~~\ |     |   |  |__) /~~\ .__/ |___ `

const (
	tagline    = "The command-line interface for Craftybase"
	tryExample = "craftybase materials list"
	docsURL    = "https://craftybase.com/docs/api"
)

const descCol = 48 // column (in the command list) where descriptions start

// cmdRow controls how one top-level command is displayed.
// desc empty => fall back to the command's Short from the live tree.
type cmdRow struct {
	name string
	args string   // e.g. "<METHOD> <path>", "<command>", "<shell>"; "" if none
	subs []string // shown joined by " | "; nil if none
	desc string
}

var commandRows = []cmdRow{
	{name: "help", args: "<command>", desc: "Display help for a command"},
	{name: "account"},
	{name: "api", args: "<METHOD> <path>"},
	{name: "auth", subs: []string{"login", "status", "logout"}, desc: "Authenticate with Craftybase"},
	{name: "materials", subs: []string{"list", "show"}},
	{name: "completion", args: "<shell>", desc: "Generate shell completion scripts"},
	{name: "version"},
}

type kv struct{ name, desc string }

var flagRows = []kv{
	{"    --json", "Output raw API envelope (pretty-printed JSON)"},
	{"    --ndjson", "Output auto-paginated NDJSON stream"},
	{"    --token <token>", "API token (overrides stored credentials)"},
	{"    --api-url <url>", "API base URL"},
	{"    --no-color", "Disable ANSI color output"},
	{"    --verbose", "Show HTTP request/response detail (token redacted)"},
	{"-h, --help", "Show help for a command"},
}

var envVarRows = []kv{
	{brand.EnvTokenName, "API token used for requests (CI, scripts)"},
	{brand.EnvAPIURL, "API base URL override"},
	{"NO_COLOR", "Disable colored output (no-color.org convention)"},
}

type renderOpts struct {
	color     bool
	trueColor bool
	width     int // <= 0 means unknown / assume wide
}

func logoWidth() int {
	max := 0
	for _, line := range strings.Split(rootLogo, "\n") {
		if n := utf8.RuneCountInString(line); n > max {
			max = n
		}
	}
	return max
}

func shortByName(root *cobra.Command, name string) string {
	for _, c := range root.Commands() {
		if c.Name() == name {
			return c.Short
		}
	}
	return ""
}

func renderRootHelp(root *cobra.Command, w io.Writer, opts renderOpts) {
	st := output.Styler{Color: opts.color, TrueColor: opts.trueColor}

	fmt.Fprintln(w)
	if opts.width <= 0 || opts.width >= logoWidth() {
		for _, line := range strings.Split(rootLogo, "\n") {
			fmt.Fprintln(w, st.Fg(output.TealBright, line))
		}
	} else {
		fmt.Fprintln(w, st.Bold("Craftybase"))
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, st.Bold(tagline))
	fmt.Fprintln(w)

	for _, r := range commandRows {
		fmt.Fprintln(w, renderCommandRow(st, root, r))
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, st.Bold("Flags:"))
	for _, f := range flagRows {
		fmt.Fprintln(w, "  "+pad(st.Fg(output.TealBright, f.name), f.name, 20)+st.Fg(output.Gray, f.desc))
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, st.Bold("Environment Variables:"))
	for _, e := range envVarRows {
		fmt.Fprintln(w, "  "+pad(st.Fg(output.TealBright, e.name), e.name, 25)+st.Fg(output.Gray, e.desc))
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, st.Fg(output.Gray, "try: ")+st.Fg(output.Terracotta, st.Bold(tryExample)))
	fmt.Fprintln(w)
	fmt.Fprintln(w, st.Fg(output.Gray, "Learn more at ")+st.Underline(st.Fg(output.TealBright, docsURL)))
}

// pad appends spaces after colored so that the underlying plain text reaches
// width columns (minimum two trailing spaces).
func pad(colored, plain string, width int) string {
	n := width - utf8.RuneCountInString(plain)
	if n < 2 {
		n = 2
	}
	return colored + strings.Repeat(" ", n)
}

func renderCommandRow(st output.Styler, root *cobra.Command, r cmdRow) string {
	desc := r.desc
	if desc == "" {
		desc = shortByName(root, r.name)
	}

	var plain, colored strings.Builder
	add := func(p, c string) { plain.WriteString(p); colored.WriteString(c) }

	add("$ ", st.Fg(output.GrayDim, "$ "))
	add("craftybase ", st.Fg(output.TealBright, "craftybase "))
	add(r.name, st.Fg(output.TealBright, r.name))
	if r.args != "" {
		add(" "+r.args, st.Fg(output.GrayDim, " "+r.args))
	}
	for i, s := range r.subs {
		sep := " "
		if i > 0 {
			sep = " | "
		}
		add(sep, st.Fg(output.GrayDim, sep))
		add(s, st.Fg(output.TealLight, s))
	}

	n := descCol - utf8.RuneCountInString(plain.String())
	if n < 2 {
		n = 2
	}
	return "  " + colored.String() + strings.Repeat(" ", n) + st.Fg(output.Gray, desc)
}
```

- [ ] **Step 4: Resolve the pflag dependency, then run tests to verify they pass**

Run: `go mod tidy && go test ./commands/ -run 'TestRenderRootHelp|TestCommandRowsCover|TestFlagRowsCover' -v`
Expected: PASS (all six tests). `go mod tidy` moves `github.com/spf13/pflag` from indirect to direct in `go.mod` (offline; already in `go.sum`).

- [ ] **Step 5: Commit**

```bash
git add commands/roothelp.go commands/roothelp_test.go go.mod go.sum
git commit -m "feat: add branded root help screen renderer (CU-868k0t6gb)"
```

---

### Task 3: Wire the renderer into Cobra

**Files:**
- Modify: `commands/root.go`
- Test: `commands/roothelp_test.go` (append integration tests; same `package commands`)

**Interfaces:**
- Consumes (from Task 2): `renderRootHelp(root, w, opts)`, `renderOpts`.
- Consumes (from Task 1): `output.ColorEnabled`, `output.SupportsTrueColor`.
- Produces: `func resolveRenderOpts() renderOpts`; Cobra wiring so the branded screen shows for the root command while subcommands keep default help.

- [ ] **Step 1: Write the failing test**

Append to `commands/roothelp_test.go`:

```go
func TestExecute_NoArgs_ShowsBrandedScreen(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "The command-line interface for Craftybase") {
		t.Errorf("no-args invocation should show the branded screen, got:\n%s", out)
	}
}

func TestExecute_RootHelpFlag_ShowsBrandedScreen(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"--help"})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(buf.String(), "The command-line interface for Craftybase") {
		t.Errorf("--help should show the branded screen")
	}
}

func TestExecute_SubcommandHelp_UsesDefault(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"materials", "--help"})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "The command-line interface for Craftybase") {
		t.Errorf("subcommand help must not show the branded root screen")
	}
	if !strings.Contains(out, "Usage:") {
		t.Errorf("subcommand help should be default Cobra help, got:\n%s", out)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./commands/ -run TestExecute -v`
Expected: FAIL — `TestExecute_NoArgs_ShowsBrandedScreen` and `TestExecute_RootHelpFlag_ShowsBrandedScreen` fail because the root still prints default Cobra usage (no branded tagline / wiring not present). `resolveRenderOpts` is undefined once referenced — that arrives in Step 3.

- [ ] **Step 3: Write minimal implementation**

In `commands/root.go`, update the imports block to add `os`, `golang.org/x/term`, and the `output` package:

```go
import (
	"fmt"
	"os"

	"github.com/craftybase/craftybase-cli/internal/api"
	"github.com/craftybase/craftybase-cli/internal/config"
	"github.com/craftybase/craftybase-cli/internal/output"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)
```

At the end of the existing `init()` in `commands/root.go` (after the `PersistentPreRunE` assignment), add the wiring:

```go
	defaultHelp := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		if c == rootCmd {
			renderRootHelp(c, c.OutOrStdout(), resolveRenderOpts())
			return
		}
		defaultHelp(c, args)
	})
	rootCmd.Run = func(c *cobra.Command, args []string) {
		renderRootHelp(c, c.OutOrStdout(), resolveRenderOpts())
	}
```

Add `resolveRenderOpts` to `commands/root.go` (new function at the bottom of the file):

```go
func resolveRenderOpts() renderOpts {
	color := output.ColorEnabled(flagNoColor)
	width := 0
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}
	return renderOpts{
		color:     color,
		trueColor: color && output.SupportsTrueColor(),
		width:     width,
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./commands/ -run TestExecute -v`
Expected: PASS (all three integration tests). Under `go test`, stdout is not a TTY, so `resolveRenderOpts` yields `color:false` and the captured output is the plain branded screen.

- [ ] **Step 5: Run the full suite and build**

Run: `go build ./... && go test ./...`
Expected: build succeeds; all packages pass.

- [ ] **Step 6: Commit**

```bash
git add commands/root.go commands/roothelp_test.go
git commit -m "feat: render branded help for root command, default for subcommands (CU-868k0t6gb)"
```

---

### Task 4: Manual verification & golden snapshot

**Files:**
- Create: `testdata/golden/roothelp_plain.golden`
- Test: `commands/roothelp_test.go` (append one snapshot test)

**Interfaces:**
- Consumes: `renderRootHelp`, `renderOpts` (Task 2).
- Produces: a committed plain-mode snapshot guarding against accidental layout/copy regressions.

- [ ] **Step 1: Eyeball the real binary (all three modes)**

```bash
go build -o craftybase ./cmd/craftybase
COLORTERM=truecolor ./craftybase            # full color (if your terminal supports truecolor)
./craftybase --no-color | cat               # plain, piped: no ANSI, logo still present
./craftybase | cat                          # piped: not a TTY -> plain
```
Expected: teal `CRAFTYBASE` logo, grouped commands, Flags/Env sections, terracotta `try:` line, underlined teal footer URL; piped output has no escape codes; logo present in all.

- [ ] **Step 2: Add `os` to the test imports, then write the snapshot test**

First update the import block at the top of `commands/roothelp_test.go` to include `os`:

```go
import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)
```

Then append the snapshot test (with an update toggle so the golden can be regenerated intentionally):

```go
func TestRenderRootHelp_PlainGolden(t *testing.T) {
	got := renderToString(renderOpts{color: false, trueColor: false, width: 100})
	const path = "../testdata/golden/roothelp_plain.golden"

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden (run with UPDATE_GOLDEN=1 to create): %v", err)
	}
	if got != string(want) {
		t.Errorf("root help plain output changed; re-run with UPDATE_GOLDEN=1 if intentional")
	}
}
```

- [ ] **Step 3: Create the golden file and verify the test passes**

Run:
```bash
UPDATE_GOLDEN=1 go test ./commands/ -run TestRenderRootHelp_PlainGolden
go test ./commands/ -run TestRenderRootHelp_PlainGolden -v
```
Expected: first command writes `testdata/golden/roothelp_plain.golden`; second PASSES against it.

- [ ] **Step 4: Full suite, then commit**

Run: `go test ./...`
Expected: all pass.

```bash
git add commands/roothelp_test.go testdata/golden/roothelp_plain.golden
git commit -m "test: golden snapshot for branded root help (CU-868k0t6gb)"
```

---

## Self-Review

**Spec coverage:**
- Logo (flat teal, 48×3, exact art) → Task 2 `rootLogo` + render. ✓
- Layout (logo→tagline→commands→Flags→Env→try→footer) → Task 2 `renderRootHelp`. ✓
- Color palette (all 5 brand colors + bold/underline) → Task 1 palette + Styler; Task 2 usage. ✓
- Modes: full color / 256 fallback / no-color / narrow → Task 1 (`SupportsTrueColor`, 256 path) + Task 2 (`renderOpts`, narrow branch) + Task 3 (`resolveRenderOpts`). Tests cover all four. ✓
- Command list derived from tree with override table + coverage test → Task 2 `commandRows`, `shortByName`, `TestCommandRowsCoverAllCommands`. ✓
- Flag table + persistent-flag coverage test → Task 2 `flagRows`, `TestFlagRowsCoverPersistentFlags`. ✓
- Env vars from `brand` constants → Task 2 `envVarRows`. ✓
- Scope: root only, subcommands default → Task 3 `SetHelpFunc` guard + `TestExecute_SubcommandHelp_UsesDefault`. ✓
- No new deps → only `pflag` promoted from indirect via `go mod tidy`. ✓
- Width via `term.GetSize`, assume-wide on failure → Task 3 `resolveRenderOpts` + `width <= 0` branch. ✓

**Placeholder scan:** No TBD/TODO; every code step shows complete code; every run step states expected output. ✓

**Type consistency:** `renderOpts{color,trueColor,width}` used identically in Tasks 2 & 3. `Styler{Color,TrueColor}` and `Fg/Bold/Underline` signatures match between Task 1 definition and Task 2 use. `renderRootHelp(root *cobra.Command, w io.Writer, opts renderOpts)` consistent across Tasks 2 & 3 and all tests. `cmdRow`/`kv` field names consistent. ✓

**Deviations from the approved mockup (intentional, noted for the implementer):** subcommands render in the order given by `commandRows` (`login | status | logout`, `list | show`) exactly as approved; app-command descriptions come from each command's `Short` (account/api/materials/version match the mockup verbatim); `help`, `completion`, and `auth` descriptions are explicit overrides to match the mockup. No behavioral deviation.
