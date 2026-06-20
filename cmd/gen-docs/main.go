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

// outDir is relative to the repo root; run gen-docs from there.
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
	entries, err := filepath.Glob(filepath.Join(outDir, "*.md"))
	if err != nil {
		return err
	}
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
	return fmt.Sprintf("---\ntitle: %q\ndescription: %q\n---\n\n", title, fmt.Sprintf("Reference for the %s command.", title))
}

// linkHandler maps a generated filename to its Starlight route.
func linkHandler(filename string) string {
	base := strings.TrimSuffix(filename, ".md")
	return "/reference/" + base + "/"
}
