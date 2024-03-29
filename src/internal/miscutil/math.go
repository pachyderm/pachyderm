package miscutil

import "golang.org/x/exp/constraints"

// Min returns the smaller of a and b.
func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Max returns the larger of a and b.
func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}
