package output

import (
	"fmt"
	"os"
)

// RGB is a 24-bit color.
type RGB struct{ R, G, B uint8 }

// Stocksmith brand palette.
var (
	TealBright = RGB{62, 177, 193}  // #3EB1C1
	TealLight  = RGB{101, 193, 205} // #65C1CD
	GrayDim    = RGB{111, 111, 111} // #6F6F6F
	Gray       = RGB{131, 127, 127} // #837f7f
	Terracotta = RGB{196, 141, 129} // #C48D81
)

// SupportsTrueColor reports whether the terminal advertises 24-bit color.
func SupportsTrueColor() bool {
	ct := os.Getenv("COLORTERM")
	return ct == "truecolor" || ct == "24bit"
}

// Styler renders ANSI styling, honoring color and truecolor capabilities.
type Styler struct {
	Color     bool
	TrueColor bool
}

// Fg wraps text in a foreground color (24-bit when available, else 256-color).
func (s Styler) Fg(c RGB, text string) string {
	if !s.Color {
		return text
	}
	if s.TrueColor {
		return fmt.Sprintf("\033[38;2;%d;%d;%dm%s\033[0m", c.R, c.G, c.B, text)
	}
	return fmt.Sprintf("\033[38;5;%dm%s\033[0m", to256(c), text)
}

// Bold wraps text in the bold attribute.
func (s Styler) Bold(text string) string {
	if !s.Color {
		return text
	}
	return "\033[1m" + text + "\033[0m"
}

// Underline wraps text in the underline attribute.
func (s Styler) Underline(text string) string {
	if !s.Color {
		return text
	}
	return "\033[4m" + text + "\033[0m"
}

// to256 approximates an RGB color with an xterm-256 palette index.
func to256(c RGB) int {
	if c.R == c.G && c.G == c.B {
		switch {
		case c.R < 8:
			return 16
		case c.R > 248:
			return 231
		default:
			return 232 + (int(c.R)-8)*24/247
		}
	}
	r := int(c.R) * 5 / 255
	g := int(c.G) * 5 / 255
	b := int(c.B) * 5 / 255
	return 16 + 36*r + 6*g + b
}
