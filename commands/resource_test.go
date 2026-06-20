package commands

import (
	"strings"
	"testing"
)

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
		{"json+all", true, false, true, 0, "--json and --all"},
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
