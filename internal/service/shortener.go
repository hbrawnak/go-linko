package service

import (
	"crypto/rand"
	"regexp"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var base62Regex = regexp.MustCompile("^[a-zA-Z0-9]+$")

const shortCodeLenMin = 7
const shortCodeLenMax = 10

func GenerateShotCode(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	for i := 0; i < n; i++ {
		b[i] = letters[int(b[i])%len(letters)]
	}

	return string(b), nil
}

func IsBase62(s string) bool {
	return base62Regex.MatchString(s)
}

func IsLengthOk(s string) bool {
	return len(s) >= shortCodeLenMin && len(s) <= shortCodeLenMax
}
