package main

import (
	"auth/api"
	"context"
	"log"
	"os"
)

func main() {
	log.Println("starting server")

	err := api.NewServer(context.Background())
	if err != nil {
		os.Exit(1)
	}
}
