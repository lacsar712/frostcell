package main

import (
	"log"
	"os"

	"github.com/lacsar712/frostcell/internal/app"
)

func main() {
	if err := app.RunFromEnv(); err != nil {
		log.Printf("frostcell stopped: %v", err)
		os.Exit(1)
	}
}
