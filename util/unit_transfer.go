package util

import "math"

// ToCents 将元转为分
func ToCents(amount float64) int64 {
	return int64(math.Round(amount * 100))
}
