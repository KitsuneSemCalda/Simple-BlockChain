package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"KitsuneSemCalda/SBC/internal/p2p"
	"KitsuneSemCalda/SBC/internal/structures"
)

func main() {
	// Initialize the blockchain
	_ = structures.NewBlockchain()
	fmt.Println("Simple Blockchain Daemon starting...")

	// Initialize the P2P host
	cfg := p2p.DefaultConfig()
	host, err := p2p.NewHost(cfg)
	if err != nil {
		log.Fatalf("failed to create host: %v", err)
	}

	fmt.Printf("Node ID: %s\n", host.ID().String())
	fmt.Println("Listening on:")
	for _, addr := range host.Addrs() {
		fmt.Printf("- %s/p2p/%s\n", addr, host.ID().String())
	}

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Daemon is running. Press Ctrl+C to stop.")
	<-stop

	fmt.Println("Shutting down...")
	if err := host.Close(); err != nil {
		log.Printf("error closing host: %v", err)
	}
	fmt.Println("Goodbye!")
}
