package main

import (
	"flag"
	"log"
	"os"

	"github.com/anujdecoder/ashta-board/server"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "8080", "Server port")
	flag.Parse()

	// Allow port to be set via environment variable
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	srv := server.NewServer(port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
