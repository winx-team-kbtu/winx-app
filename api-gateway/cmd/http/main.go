package main

import (
	"context"
	"log"
	"os"

	"winx-api-gateway/api"
)

func main() {
	log.Println("starting api gateway")

	if err := api.NewServer(context.Background()); err != nil {
		log.Printf("api gateway stopped: %v", err)
		os.Exit(1)
	}
}
