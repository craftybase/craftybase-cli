package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestFormatKeyValue_AlignsLabels(t *testing.T) {
	var buf bytes.Buffer
	FormatKeyValue(&buf, [][2]string{
		{"ID", "1042"},
		{"STATUS", "completed"},
		{"NOTES", "—"},
	}, false)
	out := buf.String()

	// Labels pad to the widest ("STATUS" = 6) + 2-space gap.
	for _, want := range []string{"ID      1042", "STATUS  completed", "NOTES   —"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing aligned line %q\n---\n%s", want, out)
		}
	}
	// No-color path must not emit ANSI escapes.
	if strings.Contains(out, "\033[") {
		t.Errorf("no-color output should contain no ANSI escapes, got %q", out)
	}
}
