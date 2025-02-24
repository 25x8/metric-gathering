package main

import (
	"log"
	"net/http"

	"github.com/25x8/metric-gathering/internal/app"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// Инициализация приложения и получение обработчика и адреса
	h, addr, key := app.InitializeApp()

	// Обеспечиваем синхронизацию логгера перед завершением работы
	defer app.SyncLogger()

	// Закрываем подключение к базе данных при выходе
	defer h.CloseDB()

	// Инициализация маршрутизатора
	r := app.InitializeRouter(h, key)

	// Запуск сервера
	log.Printf("Server started at %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
