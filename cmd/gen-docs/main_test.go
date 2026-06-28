package main

import (
	"strings"
	"testing"
)

func TestFilePrepender(t *testing.T) {
	got := filePrepender("/anywhere/stocksmith_materials_list.md")
	if !strings.HasPrefix(got, "---\n") {
		t.Fatalf("expected YAML frontmatter, got %q", got)
	}
	if !strings.Contains(got, `title: "stocksmith materials list"`) {
		t.Errorf("expected quoted derived title, got %q", got)
	}
}

func TestLinkHandler(t *testing.T) {
	if got := linkHandler("stocksmith_materials.md"); got != "/reference/stocksmith_materials/" {
		t.Errorf("got %q, want /reference/stocksmith_materials/", got)
	}
}
