package common

// Ptr returns a pointer to the given value.
// Useful for constructing pointer literals inline.
func Ptr[T any](v T) *T {
	return &v
}

func Ternary[T any](condition bool, left, right T) T {
	if condition {
		return left
	}
	return right
}
