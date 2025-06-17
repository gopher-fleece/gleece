package common

func Ptr[T any](v T) *T {
	return &v
}
