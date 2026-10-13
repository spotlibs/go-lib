package excel

import (
	"github.com/spotlibs/go-lib/stderr"
	"github.com/xuri/excelize/v2"
)

// DataValidation is a re-export of excelize.DataValidation.
// Callers use excelize.NewDataValidation() to build a *DataValidation,
// configure it, then pass it to AddDataValidation.
type DataValidation = excelize.DataValidation

// AddDataValidation adds a data validation rule to the given sheet.
// sheet: sheet name. dv: pointer to a configured DataValidation.
// Returns error if adding fails.
func (e *ExcelFile) AddDataValidation(sheet string, dv *DataValidation) error {
	if err := e.f.AddDataValidation(sheet, dv); err != nil {
		return stderr.ErrRuntime("failed to add data validation on sheet: " + sheet)
	}
	return nil
}

// DeleteDataValidation removes data validation rules from the given range on the sheet.
// sheet: sheet name. sqref: optional cell range strings e.g. "A1:A10".
// Pass no sqref to remove all validations on sheet.
// Returns error if deletion fails.
func (e *ExcelFile) DeleteDataValidation(sheet string, sqref ...string) error {
	if err := e.f.DeleteDataValidation(sheet, sqref...); err != nil {
		return stderr.ErrRuntime("failed to delete data validation on sheet: " + sheet)
	}
	return nil
}
