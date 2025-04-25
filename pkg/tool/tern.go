package tool

// Tern 三目运算
func Tern[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}
