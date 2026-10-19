package types

import "github.com/goravel/framework/support/carbon"

func Ptr[T any](v T) *T {
	return &v
}

// PtrFormat calls formatter on v if v is non-nil, returning a *string.
// Returns nil if v is nil. The caller controls all formatting logic.
func PtrFormat[T any](v *T, formatter func(T) string) *string {
	if v == nil {
		return nil
	}
	s := formatter(*v)
	return &s
}

// PtrDateTimeString converts a *carbon.DateTime to a formatted *string.
// Returns nil if dt is nil or is a zero date (0001-01-01 00:00:00),
// preventing zero-dates stored in the database from leaking into API responses.
func PtrDateTimeString(dt *carbon.DateTime) *string {
	if dt != nil && dt.IsZero() {
		return nil
	}
	return PtrFormat(dt, func(d carbon.DateTime) string { return d.ToDateTimeString() })
}
