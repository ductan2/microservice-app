package main

import (
	"content-services/internal/config"
	"content-services/internal/server"
	"log"
)

func main() {
	// Determine port (defaults to 8001)
	port := config.GetPort()

	r := server.NewRouter()

	addr := ":" + port
	log.Printf("Starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
