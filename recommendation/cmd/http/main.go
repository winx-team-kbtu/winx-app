package main

import (
	"context"
	"log"
	"os"
	"winx-recommendation/api"
)

func main() {
	log.Println("starting recommendation service")

	if err := api.NewServer(context.Background()); err != nil {
		os.Exit(1)
	}
}
