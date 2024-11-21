package main

import (
	"github.com/25x8/metric-gathering/cmd/app"
	"log"
	"net/http"
)

func main() {
	// Инициализация приложения и получение обработчика и адреса
	h, addr := app.InitializeApp()

	// Обеспечиваем синхронизацию логгера перед завершением работы
	defer app.SyncLogger()

	// Инициализация маршрутизатора
	r := app.InitializeRouter(h)

	// Запуск сервера
	log.Printf("Server started at %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
