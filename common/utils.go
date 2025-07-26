package common

// Ptr returns a pointer to the given value.
// Useful for constructing pointer literals inline.
func Ptr[T any](v T) *T {
	return &v
}

// Coalesce returns the first non-zero of the given arguments.
// Note that the meaning of 'zero' may change with the type
func Coalesce[T comparable](args ...T) T {
	var zero T
	for _, arg := range args {
		if arg != zero {
			return arg
		}
	}
	return zero
}
