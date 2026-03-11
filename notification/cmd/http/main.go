package main

import (
	"context"
	"log"
	"os"
	"winx-notification/api"
)

func main() {
	log.Println("starting server")

	err := api.NewServer(context.Background())
	if err != nil {
		log.Printf("notification worker stopped: %v", err)
		os.Exit(1)
	}
}
