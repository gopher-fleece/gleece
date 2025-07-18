package common

func Ptr[T any](v T) *T {
	return &v
}

func Map[T any, R any](in []T, f func(T) R) []R {
	out := make([]R, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}

func Flatten[T any](nested [][]T) []T {
	var flat []T
	for _, inner := range nested {
		flat = append(flat, inner...)
	}
	return flat
}
