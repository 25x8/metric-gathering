package main

import (
	"fmt"
	"log"

	pattern "github.com/25x8/metric-gathering/examples/exit_pattern/pkg"
)

func main() {
	fmt.Println("Example showing return pattern instead of os.Exit")

	if err := pattern.Run(); err != nil {
		log.Println("Error:", err)

		if exitErr, ok := err.(*pattern.ExitError); ok {
			fmt.Printf("Would exit with code %d\n", exitErr.Code)
		}

		return
	}

	fmt.Println("Program completed successfully")
}
