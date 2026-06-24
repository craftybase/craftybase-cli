package output

import "testing"

func TestFormatBool(t *testing.T) {
	if got := FormatBool(true); got != "yes" {
		t.Errorf("true should render yes, got %q", got)
	}
	if got := FormatBool(false); got != "no" {
		t.Errorf("false should render no, got %q", got)
	}
}
