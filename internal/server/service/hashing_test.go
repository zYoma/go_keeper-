package service

import (
	"crypto/sha256"
	"testing"
)

// TestGetHashStr проверяет функцию GetHashStr
func TestGetHashStr(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"password", string(hashBytes("password"))},
		{"123456", string(hashBytes("123456"))},
		{"", string(hashBytes(""))},
		{"a long password with spaces", string(hashBytes("a long password with spaces"))},
	}

	for _, test := range tests {
		result, err := GetHashStr(test.input)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result != test.expected {
			t.Errorf("For input '%s', expected '%x' but got '%x'", test.input, test.expected, result)
		}
	}
}

// hashBytes вычисляет SHA-256 хэш строки и возвращает его как байтовый срез
func hashBytes(s string) []byte {
	h := sha256.New()
	h.Write([]byte(s))
	return h.Sum(nil)
}
