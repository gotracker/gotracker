package util

func LerpFloat32(t float64, a, b float32) float32 {
	return float32(LerpFloat64(t, float64(a), float64(b)))
}

func LerpFloat64(t float64, a, b float64) float64 {
	if t <= 0 {
		return a
	} else if t >= 1 {
		return b
	}
	return a + (t * (b - a))
}

func LerpInt(t float64, a, b int) int {
	return int(LerpFloat64(t, float64(a), float64(b)))
}

func LerpUint(t float64, a, b uint) uint {
	return uint(LerpFloat64(t, float64(a), float64(b)))
}
