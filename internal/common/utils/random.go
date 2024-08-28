package utils

import "math/rand"

const (
	charset = "abcdefghijklmnopqrstuvwxyz"
	length  = 10
)

func GetRandomName() string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))] //nolint:gosec // this is not a security-sensitive context
	}
	return string(b)
}
