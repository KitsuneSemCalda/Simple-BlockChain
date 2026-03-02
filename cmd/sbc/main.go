package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"KitsuneSemCalda/SBC/internal/blockchain"
	"KitsuneSemCalda/SBC/internal/p2p"
)

func main() {
	bc := blockchain.NewBlockchain()
	cfg := p2p.DefaultConfig()
	
	debug := flag.Bool("debug", false, "Enable debug logging")
	cfg.ParseFlags() // Note: ParseFlags calls flag.Parse(), so our debug flag is handled

	if *debug {
		p2p.SetLogLevel(p2p.LevelDebug)
	} else {
		p2p.SetLogLevel(p2p.LevelWarn)
	}

	server, err := p2p.NewServer(cfg, bc)
	if err != nil {
		fmt.Printf("failed to create server: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server.StartMaintenance(ctx)

	go func() {
		if err := server.Start(ctx); err != nil {
			// server.Start currently only waits for context, but it's good practice
		}
	}()

	for _, addr := range cfg.BootNode {
		if addr == "" {
			continue
		}
		server.ConnectToPeer(addr)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Simple Blockchain CLI (DAG Viewer Mode)")
	fmt.Printf("Node ID: %s\n", server.GetHostID())
	fmt.Printf("Listening on: %s\n", server.GetAddrs())
	fmt.Println("Commands: print, validate, length, peers, addr, connect <addr>, discover, help, quit")

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
		case "print":
			blocks := bc.GetAllBlocks()
			fmt.Println("\nBlockchain DAG View:")
			for i, b := range blocks {
				prefix := "  "
				if i > 0 {
					prefix = "──→ "
				}
				fmt.Printf("%s[%d|%s...]", prefix, b.Index, b.Hash[:8])
			}
			fmt.Println("\n")
		case "validate":
			if bc.IsValid() {
				fmt.Println("✓ Blockchain is locally valid!")
			} else {
				fmt.Println("✗ Blockchain is CORRUPTED!")
			}
		case "length":
			fmt.Printf("Blockchain length: %d blocks\n", bc.Length())
		case "peers":
			peers := server.GetPeers()
			if len(peers) == 0 {
				fmt.Println("No peers connected")
				continue
			}
			fmt.Printf("Connected peers (%d):\n", len(peers))
			for id, p := range peers {
				fmt.Printf("  - %s (height: %d)\n", id.String()[:16], p.BestHeight)
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
				continue
			}
			addr := args[0]
			if err := server.ConnectToPeer(addr); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Connected to %s\n", addr)
			}
		case "discover":
			fmt.Println("Initiating peer discovery via network...")
			server.DiscoverPeers()
		case "help":
			fmt.Println("Available commands:")
			fmt.Println("  print           - Visual DAG view of the blockchain")
			fmt.Println("  validate        - Validate blockchain integrity")
			fmt.Println("  length          - Show blockchain length")
			fmt.Println("  peers           - Show connected peers")
			fmt.Println("  addr            - Show your full addresses")
			fmt.Println("  connect <addr>  - Connect to a peer via multiaddr")
			fmt.Println("  discover        - Discover more peers via gossip")
			fmt.Println("  help            - Show this help message")
			fmt.Println("  quit            - Exit")
		case "quit", "exit":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Printf("Unknown command: %s. Type 'help' for commands.\n", command)
		}
	}
}
