package utils

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"net/url"
	"regexp"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var base62Regex = regexp.MustCompile("^[a-zA-Z0-9]+$")

const ShortCodeLenMin = 7
const ShortCodeLenMax = 8

func ValidateOriginalURL(u string) error {
	if u == "" {
		return errors.New("url is required")
	}

	_, err := url.ParseRequestURI(u)
	if err != nil {
		return err
	}

	return nil
}

func ValidateShortCode(code string) error {
	if code == "" {
		return errors.New("code is required")
	}

	if !IsBase62(code) {
		return errors.New("code is invalid")
	}

	if !IsLengthOk(code) {
		return errors.New("code is invalid")
	}

	return nil
}

func IsBase62(code string) bool {
	return base62Regex.MatchString(code)
}

func IsLengthOk(code string) bool {
	return len(code) >= ShortCodeLenMin && len(code) <= ShortCodeLenMax
}

func ToBase62(num uint64) string {
	if num == 0 {
		return string(letters[0])
	}
	var encoded []byte
	for num > 0 {
		remainder := num % 62
		encoded = append([]byte{letters[remainder]}, encoded...)
		num = num / 62
	}
	return string(encoded)
}

func HashToBase62(input string) string {
	hash := sha256.Sum256([]byte(input))

	num := binary.BigEndian.Uint64(hash[:8])
	code := ToBase62(num)

	for len(code) < ShortCodeLenMin {
		code = "0" + code
	}

	if len(code) > ShortCodeLenMax {
		code = code[:ShortCodeLenMax]
	}

	return code
}
