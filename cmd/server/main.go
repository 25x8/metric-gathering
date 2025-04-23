package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	"github.com/25x8/metric-gathering/internal/app"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	memProfile := flag.Bool("memprofile", false, "enable memory profiling")
	flag.Parse()

	// Инициализация логгера и обеспечение его синхронизации
	defer app.SyncLogger()

	h, addr, key := app.InitializeApp()

	defer h.CloseDB()

	// Если нужно профилирование, создаем файл профиля
	var profileFile *os.File
	if *memProfile {
		// Создаем директорию profiles, если её нет
		if err := os.MkdirAll("profiles", 0755); err != nil {
			log.Fatalf("Failed to create profiles directory: %v", err)
		}

		var err error
		profileFile, err = os.Create("profiles/base.pprof")
		if err != nil {
			log.Fatalf("Failed to create profile file: %v", err)
		}
		defer func() {
			// Записываем профиль перед закрытием файла
			log.Println("Writing memory profile...")
			if err := pprof.WriteHeapProfile(profileFile); err != nil {
				log.Printf("Failed to write profile: %v", err)
			}
			profileFile.Close()
			log.Println("Profile saved to profiles/base.pprof")
		}()
	}

	r := app.InitializeRouter(h, key)

	// Канал для перехвата сигналов завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("Server started at %s\n", addr)
		if err := http.ListenAndServe(addr, r); err != nil {
			log.Printf("Server error: %v\n", err)
			// Отправляем сигнал завершения в случае ошибки
			stop <- syscall.SIGTERM
		}
	}()

	// Ожидаем сигнал завершения
	<-stop
	log.Println("Server shutdown initiated...")
	log.Println("Server exited")
}
