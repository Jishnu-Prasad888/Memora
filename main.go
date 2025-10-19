package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"Memora/client"
	"Memora/server"
	"Memora/store"
)

func main() {
	mode := flag.String("mode", "server", "Mode: server or client")
	host := flag.String("host", "localhost", "Server host")
	port := flag.String("port", "6379", "Server port")
	flag.Parse()

	switch *mode {
	case "server":
		startServer(*host, *port)
	case "client":
		startClient(*host, *port)
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		os.Exit(1)
	}
}

func startServer(host, port string) {
	srv := server.NewServer(host, port)

	// Initialize persistence
	persistence := store.NewPersistence(srv.Store, "memora-dump.rdb")

	// Load existing data
	err := persistence.Load()
	if err != nil {
		log.Printf("Warning: Could not load snapshot: %v", err)
	}

	// Start auto-save in background
	go persistence.StartAutoSave(5 * time.Minute)

	log.Fatal(srv.Start())
}

//goland:noinspection Annotator
func startClient(host, port string) {
	cli, err := client.NewClient(host, port)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer cli.Close()

	// Test connection
	result, err := cli.Ping()
	if err != nil {
		log.Fatalf("Failed to ping server: %v", err)
	}
	fmt.Printf("Connected: %s\n", result)

	cli.StartCLI()
}
