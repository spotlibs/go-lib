package excel

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/spotlibs/go-lib/stderr"
	"github.com/xuri/excelize/v2"
)

// parseCellFloat parses a raw cell value as a finite float64.
// ok is false when raw is empty (caller should return the zero value).
func parseCellFloat(raw, sheet, cell string) (value float64, ok bool, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false, nil
	}
	f, parseErr := strconv.ParseFloat(raw, 64)
	if parseErr != nil || math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, false, stderr.ErrRuntime("cell is not a number at " + sheet + "!" + cell)
	}
	return f, true, nil
}

// GetCellInt returns the cell value as an int.
// An empty cell returns 0 and nil error. A non-numeric or non-whole value returns error.
// Precision is exact up to 2^53.
func (e *ExcelFile) GetCellInt(sheet, cell string) (int, error) {
	raw, err := e.GetCellValueRaw(sheet, cell)
	if err != nil {
		return 0, err
	}
	f, ok, err := parseCellFloat(raw, sheet, cell)
	if err != nil || !ok {
		return 0, err
	}
	if f != math.Trunc(f) {
		return 0, stderr.ErrRuntime("cell is not an integer at " + sheet + "!" + cell)
	}
	return int(f), nil
}

// GetCellFloat returns the cell value as a float64.
// An empty cell returns 0 and nil error. A non-numeric value returns error.
func (e *ExcelFile) GetCellFloat(sheet, cell string) (float64, error) {
	raw, err := e.GetCellValueRaw(sheet, cell)
	if err != nil {
		return 0, err
	}
	f, ok, err := parseCellFloat(raw, sheet, cell)
	if err != nil || !ok {
		return 0, err
	}
	return f, nil
}

// GetCellTime returns the cell value as a time.Time.
// The cell must store a date/time serial number (not text). An empty cell returns
// the zero time.Time and nil error.
func (e *ExcelFile) GetCellTime(sheet, cell string) (time.Time, error) {
	raw, err := e.GetCellValueRaw(sheet, cell)
	if err != nil {
		return time.Time{}, err
	}
	f, ok, err := parseCellFloat(raw, sheet, cell)
	if err != nil || !ok {
		return time.Time{}, err
	}
	t, convErr := excelize.ExcelDateToTime(f, false)
	if convErr != nil {
		return time.Time{}, stderr.ErrRuntime("cell is not a date at " + sheet + "!" + cell)
	}
	return t, nil
}
