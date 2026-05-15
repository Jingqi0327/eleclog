package util

import (
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"
const alphanum = "abcdefghijklmnopqrstuvwxyz0123456789"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

func RandomStringwithDigits(n int) string {
	var sb strings.Builder
	k := len(alphanum)

	for i := 0; i < n; i++ {
		c := alphanum[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

func RandomName(n int) string {
	return RandomString(n)
}

func RandomCode(n int) string {
	return RandomStringwithDigits(n)
}

func RandomBalance(min, max int64) int64 {
	return RandomInt(min, max)
}

func RandomTime() time.Time {
	min := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC).Unix()
	delta := max - min

	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0)
}

func RandomEmail() string {
	return RandomString(6) + "@" + RandomString(4) + ".com"
}