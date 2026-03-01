package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"KitsuneSemCalda/SBC/internal/blockchain"
	"KitsuneSemCalda/SBC/internal/p2p"
)

func main() {
	bc := blockchain.NewBlockchain()
	cfg := p2p.DefaultConfig()
	cfg.ParseFlags()

	server, err := p2p.NewServer(cfg, bc)
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
	fmt.Println("Commands: add <bpm>, print, validate, length, peers, addr, connect <addr>, sync, discover, find <hash>, help, quit")

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		parts := strings.Fields(input)
		command := parts[0]
		args := parts[1:]
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
			bc.AddBlock(bpm)
			fmt.Printf("Block added and broadcasting...\n")
		case "print":
			bc.Print()
		case "validate":
			if bc.IsValid() {
				fmt.Println("Blockchain is valid!")
			} else {
				fmt.Println("Blockchain is INVALID!")
			}
		case "length":
			fmt.Printf("Blockchain length: %d\n", bc.Length())
		case "peers":
			peers := server.GetPeers()
			if len(peers) == 0 {
				fmt.Println("No peers connected")
				continue
			}
			fmt.Printf("Connected peers (%d):\n", len(peers))
			for id, p := range peers {
				fmt.Printf("  - %s (height: %d)\n", id, p.BestHeight)
			}
		case "addr":
			addrs := server.GetAddrs()
			nodeID := server.GetHostID()
			fmt.Println("Your full addresses (share these to connect):")
			for _, addr := range addrs {
				fmt.Printf("  %s/p2p/%s\n", addr, nodeID)
			}
		case "connect":
			if len(args) < 1 {
				fmt.Println("Usage: connect <multiaddr>")
				fmt.Println("Example: connect /ip4/127.0.0.1/tcp/8333/p2p/Qm...")
				continue
			}
			addr := args[0]
			if err := server.ConnectToPeer(addr); err != nil {
				fmt.Printf("Error connecting to peer: %v\n", err)
			} else {
				fmt.Printf("Connected to %s\n", addr)
			}
		case "sync":
			peers := server.GetPeers()
			if len(peers) == 0 {
				fmt.Println("No peers connected to sync")
				continue
			}
			fmt.Println("Requesting blocks from peers...")
			server.RequestSync()
		case "discover":
			peers := server.GetPeers()
			if len(peers) == 0 {
				fmt.Println("No peers connected to discover more")
				continue
			}
			fmt.Println("Discovering more peers...")
			server.DiscoverPeers()
		case "find":
			if len(args) < 1 {
				fmt.Println("Usage: find <block_hash>")
				continue
			}
			hash := args[0]
			peers := server.GetPeers()
			if len(peers) == 0 {
				fmt.Println("No peers connected to search")
				continue
			}
			fmt.Printf("Searching for block: %s\n", hash)
			server.FindBlock(hash)
		case "help":
			fmt.Println("Available commands:")
			fmt.Println("  add <bpm>       - Add a new block with given BPM")
			fmt.Println("  print           - Print all blocks in the blockchain")
			fmt.Println("  validate        - Validate blockchain integrity")
			fmt.Println("  length          - Show blockchain length")
			fmt.Println("  peers           - Show connected peers")
			fmt.Println("  addr            - Show your full addresses to share")
			fmt.Println("  connect <addr>  - Connect to a peer via multiaddr (use 'addr' on other node)")
			fmt.Println("  sync            - Request blocks from connected peers")
			fmt.Println("  discover        - Discover more peers from connected peers")
			fmt.Println("  find <hash>     - Find a block by hash in the network")
			fmt.Println("  help            - Show this help message")
			fmt.Println("  quit            - Exit the program")
		case "quit", "exit":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Printf("Unknown command: %s\n", command)
			fmt.Println("Available commands: add, print, validate, length, peers, addr, connect, sync, discover, find, help, quit")
		}
	}
}
