package commands

import (
	"net/url"
	"strings"
	"testing"
)

func TestProjectFilters_Apply(t *testing.T) {
	f := projectFilters{sku: "WAX-001", name: "bees", category: "Waxes", state: "active"}
	params := url.Values{}
	f.apply(params)
	want := map[string]string{
		"sku":           "WAX-001",
		"name":          "bees",
		"category_name": "Waxes", // category maps to category_name
		"state":         "active",
	}
	for k, v := range want {
		if params.Get(k) != v {
			t.Errorf("param %q: want %q, got %q", k, v, params.Get(k))
		}
	}

	empty := url.Values{}
	(&projectFilters{}).apply(empty)
	if len(empty) != 0 {
		t.Errorf("empty filters should set no params, got %v", empty)
	}
}

func TestValidateListFlags(t *testing.T) {
	cases := []struct {
		name                 string
		jsonOut, ndjson, all bool
		page                 int
		wantErr              string // substring; "" => expect no error
	}{
		{"no flags", false, false, false, 0, ""},
		{"json only", true, false, false, 0, ""},
		{"ndjson only", false, true, false, 0, ""},
		{"all only", false, false, true, 0, ""},
		{"page only", false, false, false, 2, ""},
		{"all+page", false, false, true, 2, "--all and --page"},
		{"all+ndjson", false, true, true, 0, "--all and --ndjson"},
		{"json+all", true, false, true, 0, "--json and --all are mutually exclusive"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateListFlags(tc.jsonOut, tc.ndjson, tc.all, tc.page)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}
