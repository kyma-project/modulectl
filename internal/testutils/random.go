package testutils

import "math/rand"

const (
	charset = "abcdefghijklmnopqrstuvwxyz"
)

func RandomName(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))] //nolint:gosec // this is not a security-sensitive context
	}
	return string(b)
}
