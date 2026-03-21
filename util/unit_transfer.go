package util

import (
	"fmt"
	"math"
)

// ToCents 将元转为分
func ToCents(amount float64) int64 {
	return int64(math.Round(amount * 100))
}

// ToYuan 将分转为元
func ToYuan(cents int64) float64 {
	return float64(cents) / 100
}

// FormatCentsToYuan 格式化分为元字符串，保留两位小数
func FormatCentsToYuan(cents int64) string {
	return fmt.Sprintf("%.2f", ToYuan(cents))
}
