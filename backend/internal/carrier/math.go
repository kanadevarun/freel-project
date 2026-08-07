package carrier

import "math"

// math provides Round2 to round a float to 2 decimal places.
// Go's math package doesn't have a built-in Round-to-N-places helper,
// so we define a small one here.

// Round2 rounds f to n decimal places.
func Round2(f float64, n int) float64 {
	pow := math.Pow(10, float64(n))
	return math.Round(f*pow) / pow
}
