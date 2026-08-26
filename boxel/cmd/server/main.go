package main

import (
	"log"
	"os"

	"boxel/internal/hub"
)

func main() {
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	log.Fatal(hub.ListenAndServe(host, port))
}
