package main

import (
	"log"

	"example.invalid/access-service/internal/app"
)

func main() {
	server, err := app.New()
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(server.Run(":8080"))
}
