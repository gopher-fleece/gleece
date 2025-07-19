package common

// Ptr returns a pointer to the given value.
// Useful for constructing pointer literals inline.
func Ptr[T any](v T) *T {
	return &v
}
