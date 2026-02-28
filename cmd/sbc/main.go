package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"KitsuneSemCalda/SBC/internal/p2p"
	"KitsuneSemCalda/SBC/internal/structures"
)

func main() {
	blockchain := structures.NewBlockchain()
	cfg := p2p.DefaultConfig()
	cfg.ParseFlags()

	server, err := p2p.NewServer(cfg, blockchain)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := server.Start(ctx); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	for _, addr := range cfg.BootNode {
		if addr == "" {
			continue
		}
		if err := server.ConnectToPeer(addr); err != nil {
			log.Printf("Can't connect to peer %s: %v", addr, err)
		}
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Simple Blockchain CLI (P2P Enabled)")
	fmt.Printf("Node ID: %s\n", server.GetHostID())
	fmt.Printf("Listening on: %s\n", server.GetAddrs())
	fmt.Println("Commands: add <bpm>, print, validate, length, quit")

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		parts := strings.Fields(input)
		command := parts[0]
		switch command {
		case "add":
			if len(parts) < 2 {
				fmt.Println("Usage: add <bpm>")
				continue
			}
			bpm, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("Error: BPM must be a number")
				continue
			}
			blockchain.AddBlock(bpm)
			fmt.Printf("Block added and broadcasting...\n")
		case "print":
			blockchain.Print()
		case "validate":
			if blockchain.IsValid() {
				fmt.Println("Blockchain is valid!")
			} else {
				fmt.Println("Blockchain is INVALID!")
			}
		case "length":
			fmt.Printf("Blockchain length: %d\n", blockchain.Length())
		case "quit", "exit":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Printf("Unknown command: %s\n", command)
			fmt.Println("Available commands: add, print, validate, length, quit")
		}
	}
}
