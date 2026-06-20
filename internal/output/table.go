package output

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

const placeholder = "—"

// columnGap is the number of spaces between columns.
const columnGap = 2

// FormatTable writes headers and rows as an aligned, space-padded table.
//
// Column widths are measured on the visible (rune) width of each cell, and
// padding is always plain spaces — ANSI styling is applied to cell content
// only, never to the padding. This is why we can't lean on text/tabwriter:
// tabwriter counts the bytes of color escape sequences as visible width, which
// shifts styled headers out of alignment with the data rows.
func FormatTable(w io.Writer, headers []string, rows [][]string, useColor bool) {
	widths := columnWidths(headers, rows)

	writeRow(w, headers, widths, func(_ int, cell string) string {
		if useColor {
			return bold(cell)
		}
		return cell
	})

	for _, row := range rows {
		writeRow(w, row, widths, func(_ int, cell string) string {
			if useColor && cell == placeholder {
				return dim(cell)
			}
			return cell
		})
	}
}

// columnWidths returns the visible (rune) width of the widest cell per column.
func columnWidths(headers []string, rows [][]string) []int {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = utf8.RuneCountInString(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i >= len(widths) {
				break
			}
			if n := utf8.RuneCountInString(cell); n > widths[i] {
				widths[i] = n
			}
		}
	}
	return widths
}

// writeRow renders one row, padding each cell to its column width with plain
// spaces. render styles the cell content only — never the padding — so styling
// never affects alignment. The final column is left unpadded (no trailing space).
func writeRow(w io.Writer, cells []string, widths []int, render func(col int, cell string) string) {
	var sb strings.Builder
	for i, cell := range cells {
		sb.WriteString(render(i, cell))
		if i < len(cells)-1 {
			pad := columnGap
			if i < len(widths) {
				pad += widths[i] - utf8.RuneCountInString(cell)
			}
			sb.WriteString(strings.Repeat(" ", pad))
		}
	}
	fmt.Fprintln(w, sb.String())
}

func bold(s string) string {
	return "\033[1m" + s + "\033[0m"
}

func dim(s string) string {
	return "\033[2m" + s + "\033[0m"
}
