package main

import (
	"context"
	"log"
	"os"
	"winx-match/api"
)

func main() {
	log.Println("starting match service")

	if err := api.NewServer(context.Background()); err != nil {
		os.Exit(1)
	}
}
