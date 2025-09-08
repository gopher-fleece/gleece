package common

// Map applies the function f to each element of the input slice in
// order and returns a new slice containing the results. The output slice
// will always have the same length as the input.
func Map[T any, R any](in []T, f func(T) R) []R {
	out := make([]R, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}

// MapNonZero applies the function fn to each element of the input slice,
// and returns a new slice containing only the non-zero results.
// This is especially useful for filtering out zero values like nil errors.
func MapNonZero[T any, U comparable](in []T, fn func(T) U) []U {
	var out []U
	var zero U
	for _, v := range in {
		res := fn(v)
		if res != zero {
			out = append(out, res)
		}
	}
	return out
}

func Filter[T any](in []T, filter func(T) bool) []T {
	var out []T
	for _, v := range in {
		if filter(v) {
			out = append(out, v)
		}
	}
	return out
}

func FilterNil[T any](in []*T) []*T {
	var out []*T
	for _, v := range in {
		if v != nil {
			out = append(out, v)
		}
	}
	return out
}

// Flatten takes a slice of slices and flattens it into a single slice
// by concatenating all inner slices in order.
func Flatten[T any](nested [][]T) []T {
	var flat []T
	for _, inner := range nested {
		flat = append(flat, inner...)
	}
	return flat
}

func DereferenceSliceElements[T any](slice []*T) []T {
	dereferenced := make([]T, 0, len(slice))
	for _, item := range slice {
		if item != nil {
			dereferenced = append(dereferenced, *item)
		}
	}
	return dereferenced
}
