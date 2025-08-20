package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestValidateOriginalURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"empty url", "", true},
		{"valid url", "https://example.com", false},
		{"invalid url", "not-a-url", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOriginalURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateShortCode(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		message string
		wantErr bool
	}{
		{"empty code", "", "code is required", true},
		{"too short", "fwmxe", "code is invalid", true},
		{"invalid characters", "fwmx-eRGu", "code is invalid", true},
		{"valid code", "fwmxeRGu", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			err := ValidateShortCode(tt.code)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsBase62(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"empty string", "a", true},
		{"valid lowercase", "abcdef", true},
		{"valid uppercase", "ABCDEF", true},
		{"valid numbers", "123456", true},
		{"valid mixed", "aB3Def", true},
		{"invalid hyphen", "abc-def", false},
		{"invalid underscore", "abc_def", false},
		{"invalid space", "abc def", false},
		{"invalid special chars", "abc@def", false},
		{"invalid symbols", "abc!#$%", false},
		{"single valid char", "a", true},
		{"single invalid char", "-", false},
		{"all base62 chars", "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBase62(tt.code)
			assert.Equal(t, tt.expected, result, "Expected %v for input: %s", tt.expected, tt.code)
		})
	}
}

func TestIsLengthOk(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"empty string", "", false},
		{"too short - 6 chars", "abcdef", false},
		{"minimum valid - 7 chars", "abcdefg", true},
		{"valid - 8 chars", "abcdefgh", true},
		{"too long - 9 chars", "abcdefghi", false},
		{"too long - 10 chars", "abcdefghij", false},
		{"single char", "a", false},

		{"exactly min length", strings.Repeat("a", ShortCodeLenMin), true},
		{"exactly max length", strings.Repeat("a", ShortCodeLenMax), true},
		{"one less than min", strings.Repeat("a", ShortCodeLenMin-1), false},
		{"one more than max", strings.Repeat("a", ShortCodeLenMax+1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsLengthOk(tt.code)
			assert.Equal(t, tt.expected, result, "Expected %v for length %d", tt.expected, len(tt.code))
		})
	}
}

func TestToBase62(t *testing.T) {
	tests := []struct {
		name     string
		num      uint64
		expected string
	}{
		{"zero", 0, "a"},
		{"one", 1, "b"},
		{"small number", 61, "9"},
		{"base boundary", 62, "ba"},
		{"larger number", 3844, "baa"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToBase62(tt.num)
			assert.Equal(t, tt.expected, result, "Expected %s for number %d", tt.expected, tt.num)
		})
	}
}

func TestHashToBase62(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"simple string", "hello"},
		{"url string", "https://example.com"},
		{"long string", "this is a very long string to test hashing"},
		{"special characters", "!@#$%^&*()"},
		{"unicode", "こんにちは世界"},
		{"numbers", "1234567890"},
		{"mixed", "Hello123!@#"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HashToBase62(tt.input)

			// Check length constraints
			assert.GreaterOrEqual(t, len(result), ShortCodeLenMin, "Result should be at least %d characters", ShortCodeLenMin)
			assert.LessOrEqual(t, len(result), ShortCodeLenMax, "Result should be at most %d characters", ShortCodeLenMax)

			// Check that result is valid base62
			assert.True(t, IsBase62(result), "Result should contain only base62 characters")

			// Check length is valid
			assert.True(t, IsLengthOk(result), "Result should have valid length")
		})
	}
}

func TestHashToBase62Consistency(t *testing.T) {
	inputs := []string{"test1", "test2", "https://example.com", ""}

	for _, input := range inputs {
		t.Run(fmt.Sprintf("consistency_%s", input), func(t *testing.T) {
			result1 := HashToBase62(input)
			result2 := HashToBase62(input)

			assert.Equal(t, result1, result2, "Same input should always produce same output")
		})
	}
}

func TestHashToBase62Uniqueness(t *testing.T) {
	inputs := []string{
		"input1", "input2", "input3", "different", "test", "hello", "world",
		"https://example.com", "https://google.com", "https://github.com",
	}

	results := make(map[string]string)

	for _, input := range inputs {
		result := HashToBase62(input)

		// Check if we've seen this result before
		if prevInput, exists := results[result]; exists {
			t.Errorf("Hash collision: inputs '%s' and '%s' both produced '%s'", prevInput, input, result)
		}

		results[result] = input
	}

	assert.Equal(t, len(inputs), len(results), "All inputs should produce unique outputs")
}

// Benchmark test (optional)
func BenchmarkHashToBase62(b *testing.B) {
	input := "https://example.com/very/long/path/to/test/performance"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HashToBase62(input)
	}
}
