package common

func MapKeys[TKey comparable, TValue any](m map[TKey]TValue) []TKey {
	out := make([]TKey, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func MapValues[TKey comparable, TValue any](m map[TKey]TValue) []TValue {
	out := make([]TValue, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}
