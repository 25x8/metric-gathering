package middleware

import (
	"net"
	"net/http"
	"strings"
)

// TrustedSubnetMiddleware проверяет IP-адрес клиента против доверенной подсети
func TrustedSubnetMiddleware(trustedSubnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if trustedSubnet == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Получаем IP из заголовка X-Real-IP
			realIP := r.Header.Get("X-Real-IP")
			if realIP == "" {
				// Если заголовок X-Real-IP отсутствует, пробуем получить IP из других источников
				realIP = getClientIP(r)
			}

			// Парсим доверенную подсеть
			_, trustedNet, err := net.ParseCIDR(trustedSubnet)
			if err != nil {
				// Если не удается распарсить CIDR, логируем и возвращаем 500
				http.Error(w, "Invalid trusted subnet configuration", http.StatusInternalServerError)
				return
			}

			// Парсим IP-адрес клиента
			clientIP := net.ParseIP(realIP)
			if clientIP == nil {
				http.Error(w, "Invalid client IP address", http.StatusBadRequest)
				return
			}

			// Проверяем, входит ли IP в доверенную подсеть
			if !trustedNet.Contains(clientIP) {
				http.Error(w, "Access denied from untrusted subnet", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP извлекает IP-адрес клиента из различных заголовков
func getClientIP(r *http.Request) string {
	// Проверяем стандартные заголовки в порядке приоритета
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For может содержать список IP, берем первый
		if parts := strings.Split(ip, ","); len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Fallback к RemoteAddr
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}

	return r.RemoteAddr
}
