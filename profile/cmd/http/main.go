package main

import (
	"context"
	"log"
	"os"
	"winx-profile/api"
)

func main() {
	log.Println("starting server")

	err := api.NewServer(context.Background())
	if err != nil {
		log.Printf("profile service stopped: %v", err)
		os.Exit(1)
	}
}
