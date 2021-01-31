package posmath

// Lerp returns the linear interpolation of two values
func Lerp(t, lhs, rhs float64) float64 {
	return lhs + t*(rhs-lhs)
}
