package miscutil

// Min returns the smaller of a and b.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MinInt64 returns the smaller of a and b.
func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
