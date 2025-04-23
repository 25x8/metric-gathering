package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateHashExtended(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		key      string
		expected string
	}{
		{
			name:     "Numeric input",
			data:     []byte("12345"),
			key:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			expected: "2fff8f4919567d5382d69dbb2baf68b14dbf6622db590e2bea01ef24e7a634bd",
		},
		{
			name:     "Special characters",
			data:     []byte("!@#$%^&*()"),
			key:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			expected: "875c23b72923bc391001715d04b99692293473bf1e0c0b5ddaaa26d07a12e790",
		},
		{
			name:     "Long input",
			data:     []byte("This is a longer string to test the hash calculation with more data"),
			key:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			expected: "475b0dd9a79e66f3bd612a46056b18bfb7f7ec4f20733ae4105a64d4c2d89bcc",
		},
		{
			name:     "Empty key",
			data:     []byte("test data"),
			key:      "",
			expected: "ed2abf5673fe90f2f5ce861e9a5c80bf9a419df4dcc392f8f603617e7eaa33be",
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

func TestCalculateHashBasic(t *testing.T) {
	testCases := []struct {
		name     string
		data     []byte
		key      string
		expected string
	}{
		{
			name:     "Empty data with empty key",
			data:     []byte(""),
			key:      "",
			expected: "b613679a0814d9ec772f95d778c35fc5ff1697c493715653c6c712144292c5ad",
		},
		{
			name:     "Some data with empty key",
			data:     []byte("test data"),
			key:      "",
			expected: "ed2abf5673fe90f2f5ce861e9a5c80bf9a419df4dcc392f8f603617e7eaa33be",
		},
		{
			name:     "Empty data with some key",
			data:     []byte(""),
			key:      "secret",
			expected: "f9e66e179b6747ae54108f82f8ade8b3c25d76fd30afde6c395822c530196169",
		},
		{
			name:     "Some data with some key",
			data:     []byte("test data"),
			key:      "secret",
			expected: "c66d73e3c4354ac8fa8c95dd1f3f79931d723bbc430030329a4de1fcb0993dc3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CalculateHash(tc.data, tc.key)
			assert.Equal(t, tc.expected, result)

			// Проверим результат вручную для подтверждения
			h := hmac.New(sha256.New, []byte(tc.key))
			h.Write(tc.data)
			expectedBytes := h.Sum(nil)
			expectedHex := hex.EncodeToString(expectedBytes)

			assert.Equal(t, expectedHex, result)
		})
	}
}
