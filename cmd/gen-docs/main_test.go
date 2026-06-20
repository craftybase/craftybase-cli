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
