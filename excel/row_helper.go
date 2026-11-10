package excel

import "strings"

// PadRow returns row extended with empty strings up to width cells.
// If row is already at least width long, it is returned unchanged.
// When padding is needed a new slice is returned; the input is not modified.
func PadRow(row []string, width int) []string {
	if len(row) >= width {
		return row
	}
	padded := make([]string, width)
	copy(padded, row)
	return padded
}

// PadRows applies PadRow to every row and returns a new outer slice.
func PadRows(rows [][]string, width int) [][]string {
	out := make([][]string, len(rows))
	for i, row := range rows {
		out[i] = PadRow(row, width)
	}
	return out
}

// IsBlankRow reports whether every cell in the row is empty or whitespace only.
// A nil or empty row is blank.
func IsBlankRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

// RemoveBlankRows returns a new slice with all blank rows (see IsBlankRow) removed.
// The order of the remaining rows is preserved.
func RemoveBlankRows(rows [][]string) [][]string {
	out := make([][]string, 0, len(rows))
	for _, row := range rows {
		if !IsBlankRow(row) {
			out = append(out, row)
		}
	}
	return out
}
