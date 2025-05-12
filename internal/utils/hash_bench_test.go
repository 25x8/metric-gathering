package utils

import (
	"crypto/rand"
	"testing"
)

// BenchmarkCalculateHash_Small тестирует производительность хеширования для малых данных
func BenchmarkCalculateHash_Small(b *testing.B) {
	// Создаем небольшой набор данных
	data := []byte("This is a small amount of data for benchmarking hash calculation")
	key := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" // 64-символьный ключ для SHA-256

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateHash(data, key)
	}
}

// BenchmarkCalculateHash_Medium тестирует производительность хеширования для средних данных
func BenchmarkCalculateHash_Medium(b *testing.B) {
	// Создаем средний набор данных (10 КБ)
	data := make([]byte, 10*1024) // 10 КБ
	_, err := rand.Read(data)
	if err != nil {
		b.Fatal(err)
	}

	key := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateHash(data, key)
	}
}

// BenchmarkCalculateHash_Large тестирует производительность хеширования для больших данных
func BenchmarkCalculateHash_Large(b *testing.B) {
	// Создаем большой набор данных (1 МБ)
	data := make([]byte, 1024*1024) // 1 МБ
	_, err := rand.Read(data)
	if err != nil {
		b.Fatal(err)
	}

	key := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateHash(data, key)
	}
}

// BenchmarkCalculateHash_DifferentKeys тестирует производительность хеширования с разными ключами
func BenchmarkCalculateHash_DifferentKeys(b *testing.B) {
	// Фиксированные данные
	data := []byte("Fixed data for benchmarking with different keys")

	// Генерируем разные ключи для каждой итерации
	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keyBytes := make([]byte, 32) // 32 байта = 64 символа в hex
		_, err := rand.Read(keyBytes)
		if err != nil {
			b.Fatal(err)
		}

		// Конвертируем байты в hex-строку
		key := ""
		for _, b := range keyBytes {
			key += string(hexChars[b>>4]) + string(hexChars[b&0x0F])
		}
		keys[i] = key
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateHash(data, keys[i])
	}
}

// Вспомогательные константы
var hexChars = []byte("0123456789abcdef")
