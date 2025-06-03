package utils

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLocalIP(t *testing.T) {
	ip, err := GetLocalIP()

	// Проверяем, что функция не возвращает ошибку
	assert.NoError(t, err)

	// Проверяем, что IP не пустой
	assert.NotEmpty(t, ip)

	// Проверяем, что это валидный IP-адрес
	parsedIP := net.ParseIP(ip)
	assert.NotNil(t, parsedIP, "Returned IP should be valid")

	// Проверяем, что это не loopback адрес (обычно функция возвращает настоящий IP)
	assert.NotEqual(t, "127.0.0.1", ip, "Should not return loopback IP")

	// Проверяем, что это не unspecified адрес
	assert.NotEqual(t, "0.0.0.0", ip, "Should not return unspecified IP")

	t.Logf("Local IP: %s", ip)
}
