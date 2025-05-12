package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateHash(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		key      string
		expected string
	}{
		{
			name:     "Empty data",
			data:     []byte{},
			key:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			expected: "081247dc68bb7fafbf13220013a0ab71db8b628d679161f87b5e5bd9e19b1494", // Актуальное значение хеша
		},
		{
			name:     "Simple string",
			data:     []byte("test data"),
			key:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			expected: "59397c822d58965c34f39db0138c3b2d75e1510cfaa9d9b9bacc49a943e3bd10", // Актуальное значение хеша
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateHash(tt.data, tt.key)
			if got != tt.expected {
				t.Errorf("CalculateHash() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCalculateHashEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		key  string
	}{
		{"Empty data and key", []byte{}, ""},
		{"Nil data", nil, "some-key"},
		{"Long key", []byte("data"), "very-long-key-that-exceeds-normal-key-length-for-testing-purposes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateHash(tt.data, tt.key)
			assert.NotEmpty(t, result, "Hash result should not be empty")
		})
	}
}
