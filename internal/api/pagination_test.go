package api_test

import (
	"encoding/json"
	"testing"

	"github.com/craftybase/craftybase-cli/internal/api"
)

func TestRawPageMeta_ParsesCurrentPage(t *testing.T) {
	body := []byte(`{"current_page":3,"total_pages":5,"total_count":10,"per_page":2}`)

	var m api.RawPageMeta
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatal(err)
	}

	if m.CurrentPage != 3 {
		t.Errorf("expected CurrentPage 3 (from \"current_page\"), got %d", m.CurrentPage)
	}
	if m.TotalPages != 5 {
		t.Errorf("expected TotalPages 5, got %d", m.TotalPages)
	}
}
