package output

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// FormatKeyValue writes aligned "LABEL  value" rows: the label column is padded
// to the widest label plus a 2-space gap. Labels are bolded when useColor, and a
// placeholder "—" value is dimmed — matching FormatTable's styling.
func FormatKeyValue(w io.Writer, rows [][2]string, useColor bool) {
	labelWidth := 0
	for _, r := range rows {
		if n := utf8.RuneCountInString(r[0]); n > labelWidth {
			labelWidth = n
		}
	}
	for _, r := range rows {
		label, value := r[0], r[1]
		pad := strings.Repeat(" ", labelWidth-utf8.RuneCountInString(label)+columnGap)
		renderedLabel := label
		if useColor {
			renderedLabel = bold(label)
		}
		renderedValue := value
		if useColor && value == placeholder {
			renderedValue = dim(value)
		}
		fmt.Fprintln(w, renderedLabel+pad+renderedValue)
	}
}
