package commands

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"github.com/craftybase/craftybase-cli/internal/brand"
	"github.com/craftybase/craftybase-cli/internal/output"
)

// rootLogo is the flat-teal CRAFTYBASE wordmark (48x3). Raw segments are joined
// around literal backticks so backslashes stay literal.
//
// NOTE: this art literally spells the product name, so — unlike the other
// brand-bearing strings, which route through internal/brand — it cannot be
// centralized. Regenerate it as part of the planned Stocksmith rebrand.
const rootLogo = ` __   __        ___ ___      __        __   ___
/  ` + "`" + ` |__)  /\  |__   |  \ / |__)  /\  /__` + "`" + ` |__
\__, |  \ /~~\ |     |   |  |__) /~~\ .__/ |___ `

const (
	tagline    = "The command-line interface for " + brand.ProductName
	tryExample = brand.BinaryName + " materials list"
	docsURL    = brand.DocsURL
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
	{name: "auth", subs: []string{"login", "status", "logout"}, desc: "Authenticate with " + brand.ProductName},
	{name: "materials", subs: []string{"list", "show"}},
	{name: "products", subs: []string{"list", "show"}},
	{name: "components", subs: []string{"list", "show"}},
	{name: "manufactures", subs: []string{"list", "show"}},
	{name: "recipes", subs: []string{"list", "show"}},
	{name: "expenses", subs: []string{"list", "show"}},
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
		fmt.Fprintln(w, st.Bold(brand.ProductName))
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, st.Bold(tagline))
	fmt.Fprintln(w)

	for _, r := range commandRows {
		fmt.Fprintln(w, renderCommandRow(st, root, r))
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, st.Bold("Flags:"))
	flagCol := columnWidth(flagRows)
	for _, f := range flagRows {
		fmt.Fprintln(w, "  "+pad(st.Fg(output.TealBright, f.name), f.name, flagCol)+st.Fg(output.Gray, f.desc))
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, st.Bold("Environment Variables:"))
	envCol := columnWidth(envVarRows)
	for _, e := range envVarRows {
		fmt.Fprintln(w, "  "+pad(st.Fg(output.TealBright, e.name), e.name, envCol)+st.Fg(output.Gray, e.desc))
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

// columnWidth returns the description-column offset for a set of rows: the
// longest name plus a two-space minimum gap. Deriving it (rather than
// hardcoding) keeps every description in a section aligned even as rows change.
func columnWidth(rows []kv) int {
	max := 0
	for _, r := range rows {
		if n := utf8.RuneCountInString(r.name); n > max {
			max = n
		}
	}
	return max + 2
}

func renderCommandRow(st output.Styler, root *cobra.Command, r cmdRow) string {
	desc := r.desc
	if desc == "" {
		desc = shortByName(root, r.name)
	}

	var plain, colored strings.Builder
	add := func(p, c string) { plain.WriteString(p); colored.WriteString(c) }

	add("$ ", st.Fg(output.GrayDim, "$ "))
	add(brand.BinaryName+" ", st.Fg(output.TealBright, brand.BinaryName+" "))
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
