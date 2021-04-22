package util

import "math"

func AccurancyFloat64(val float64, prec int, round bool) float64 {
	pow10 := math.Pow10(prec)

	if round {
		return math.Trunc((val + 0.5 / pow10) * pow10) / pow10
	}

	return math.Trunc(val * pow10) / pow10
}
