package output_test

import (
	"strings"
	"testing"

	"github.com/craftybase/stocksmith-cli/internal/output"
)

func TestStyler_ColorDisabled_NoANSI(t *testing.T) {
	s := output.Styler{Color: false, TrueColor: true}
	if got := s.Fg(output.AmberBright, "hello"); got != "hello" {
		t.Errorf("Fg with color off should be plain, got %q", got)
	}
	if got := s.Bold("hi"); got != "hi" {
		t.Errorf("Bold with color off should be plain, got %q", got)
	}
	if got := s.Underline("hi"); got != "hi" {
		t.Errorf("Underline with color off should be plain, got %q", got)
	}
}

func TestStyler_TrueColor_Emits24Bit(t *testing.T) {
	s := output.Styler{Color: true, TrueColor: true}
	got := s.Fg(output.AmberBright, "X")
	if !strings.Contains(got, "\033[38;2;225;142;45m") {
		t.Errorf("expected 24-bit amber sequence, got %q", got)
	}
	if !strings.HasSuffix(got, "\033[0m") {
		t.Errorf("expected reset suffix, got %q", got)
	}
	if !strings.Contains(got, "X") {
		t.Errorf("expected text preserved, got %q", got)
	}
}

func TestStyler_NoTrueColor_Emits256(t *testing.T) {
	s := output.Styler{Color: true, TrueColor: false}
	got := s.Fg(output.AmberBright, "X")
	if strings.Contains(got, "38;2;") {
		t.Errorf("expected no 24-bit sequence in 256 mode, got %q", got)
	}
	if !strings.Contains(got, "\033[38;5;") {
		t.Errorf("expected 256-color sequence, got %q", got)
	}
}

func TestStyler_BoldUnderline_WhenColor(t *testing.T) {
	s := output.Styler{Color: true, TrueColor: true}
	if got := s.Bold("X"); got != "\033[1mX\033[0m" {
		t.Errorf("unexpected bold sequence: %q", got)
	}
	if got := s.Underline("X"); got != "\033[4mX\033[0m" {
		t.Errorf("unexpected underline sequence: %q", got)
	}
}
