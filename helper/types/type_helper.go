package types

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
