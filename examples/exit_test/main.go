package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("This program uses a graceful shutdown approach")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Println("Working...")
		stop <- syscall.SIGTERM
	}()

	<-stop

	fmt.Println("Cleaning up resources...")

	fmt.Println("Exiting gracefully")
}
