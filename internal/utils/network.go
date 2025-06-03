package utils

import (
	"net"
)

// GetLocalIP возвращает локальный IP-адрес хоста
func GetLocalIP() (string, error) {
	// Создаем UDP соединение для определения локального IP
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// Получаем локальный адрес соединения
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}
