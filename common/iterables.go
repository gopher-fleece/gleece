package common

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

func AppendIfNotNil[T any](dst []T, item *T) []T {
	if item != nil {
		dst = append(dst, *item)
	}
	return dst
}

func ConcatConditional[T any](dst []T, items []T, condition func(item T) bool) []T {
	for _, item := range items {
		if condition(item) {
			dst = append(dst, item)
		}
	}
	return dst
}
