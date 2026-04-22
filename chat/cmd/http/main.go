package main

import (
	"context"
	"log"
	"os"
	"winx-chat/api"
)

func main() {
	log.Println("starting chat service")

	if err := api.NewServer(context.Background()); err != nil {
		os.Exit(1)
	}
}
