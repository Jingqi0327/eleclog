package util

import "math"

// ToCents 将元转为分
func ToCents(amount float64) int32 {
	return int32(math.Round(amount * 100))
}
