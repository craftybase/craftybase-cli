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
