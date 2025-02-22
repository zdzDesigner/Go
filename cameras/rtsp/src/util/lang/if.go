package lang

// IF3 简单三元表达
func IF3[T any](ok bool, a, b T) T {
	if ok {
		return a
	}
	return b
}
