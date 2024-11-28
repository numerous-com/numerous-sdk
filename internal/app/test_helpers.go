package app

func ref[T any](v T) *T {
	return &v
}
