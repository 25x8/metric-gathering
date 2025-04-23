package utils

import (
	"log"
	"os"
	"runtime/pprof"
)

// StartMemProfiler запускает профилирование памяти и возвращает функцию для сохранения профиля
func StartMemProfiler(filename string) func() {
	// Создаем директорию profiles, если её нет
	if err := os.MkdirAll("profiles", 0755); err != nil {
		log.Fatalf("Failed to create profiles directory: %v", err)
	}

	f, err := os.Create("profiles/" + filename)
	if err != nil {
		log.Fatalf("Failed to create memory profile: %v", err)
	}

	log.Printf("Memory profiling started. Profile will be saved to profiles/%s", filename)

	// Возвращаем функцию для записи и закрытия профиля
	return func() {
		log.Println("Writing memory profile...")
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatalf("Failed to write memory profile: %v", err)
		}
		f.Close()
		log.Printf("Memory profile saved to profiles/%s", filename)
	}
}
