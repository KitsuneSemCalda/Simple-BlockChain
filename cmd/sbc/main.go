package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"KitsuneSemCalda/SBC/internal/blockchain"
	"KitsuneSemCalda/SBC/internal/p2p"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[90m"
)

func main() {
	bc := blockchain.NewBlockchain()
	cfg := p2p.DefaultConfig()
	cfg.ParseFlags()

	// Set log level to Warn by default to keep the UI clean
	p2p.SetLogLevel(p2p.LevelWarn)

	server, err := p2p.NewServer(cfg, bc)
	if err != nil {
		fmt.Printf("%sError: failed to create server: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := server.Start(ctx); err != nil {
			p2p.Error("Server", "Server error: %v", err)
		}
	}()

	for _, addr := range cfg.BootNode {
		if addr == "" {
			continue
		}
		if err := server.ConnectToPeer(addr); err != nil {
			p2p.Warn("P2P", "Can't connect to peer %s: %v", addr, err)
		}
	}

	reader := bufio.NewReader(os.Stdin)

	printHeader(server)

	for {
		fmt.Printf("%sSBC %s>%s ", ColorBlue, ColorCyan, ColorReset)
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
			if len(args) < 1 {
				fmt.Println("Usage: add <bpm>")
				continue
			}
			bpm, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Println("Error: BPM must be a number")
				continue
			}
			bc.AddBlock(bpm)
			fmt.Printf("%s[+] Block added and broadcasting...%s\n", ColorGreen, ColorReset)
		case "print":
			limit := 10
			if len(args) > 0 {
				if l, err := strconv.Atoi(args[0]); err == nil {
					limit = l
				}
			}
			blocks := bc.GetAllBlocks()
			start := 0
			if len(blocks) > limit {
				start = len(blocks) - limit
			}
			fmt.Printf("%s--- Last %d blocks (total: %d) ---%s\n", ColorYellow, len(blocks)-start, len(blocks), ColorReset)
			for i := start; i < len(blocks); i++ {
				b := blocks[i]
				fmt.Printf("%s[%d]%s BPM: %s%d%s Hash: %s%s%s\n", 
					ColorGray, b.Index, ColorReset, 
					ColorCyan, b.BPM, ColorReset,
					ColorYellow, b.Hash[:12], ColorReset)
			}
		case "validate":
			if bc.IsValid() {
				fmt.Printf("%sBlockchain is valid!%s\n", ColorGreen, ColorReset)
			} else {
				fmt.Printf("%sBlockchain is INVALID!%s\n", ColorRed, ColorReset)
			}
		case "length":
			fmt.Printf("Blockchain length: %s%d%s\n", ColorCyan, bc.Length(), ColorReset)
		case "peers":
			peers := server.GetPeers()
			if len(peers) == 0 {
				fmt.Println("No peers connected")
				continue
			}
			fmt.Printf("%sConnected peers (%d):%s\n", ColorYellow, len(peers), ColorReset)
			for id, p := range peers {
				fmt.Printf("  - %s%s%s (height: %s%d%s)\n", ColorCyan, id.String()[:12], ColorReset, ColorYellow, p.BestHeight, ColorReset)
			}
		case "addr":
			addrs := server.GetAddrs()
			nodeID := server.GetHostID()
			fmt.Println("Your full addresses (share these to connect):")
			for _, addr := range addrs {
				fmt.Printf("  %s%s/p2p/%s%s\n", ColorCyan, addr, nodeID, ColorReset)
			}
		case "connect":
			if len(args) < 1 {
				fmt.Println("Usage: connect <multiaddr>")
				continue
			}
			addr := args[0]
			if err := server.ConnectToPeer(addr); err != nil {
				fmt.Printf("%sError connecting to peer: %v%s\n", ColorRed, err, ColorReset)
			} else {
				fmt.Printf("%sConnected to %s%s\n", ColorGreen, addr, ColorReset)
			}
		case "sync":
			fmt.Println("Requesting blocks from peers...")
			server.RequestSync()
		case "discover":
			fmt.Println("Discovering more peers...")
			server.DiscoverPeers()
		case "find":
			if len(args) < 1 {
				fmt.Println("Usage: find <block_hash>")
				continue
			}
			hash := args[0]
			fmt.Printf("Searching for block: %s\n", hash)
			server.FindBlock(hash)
		case "log":
			if len(args) < 1 {
				fmt.Println("Usage: log <debug|info|warn|error|none>")
				continue
			}
			level := strings.ToLower(args[0])
			switch level {
			case "debug": p2p.SetLogLevel(p2p.LevelDebug)
			case "info":  p2p.SetLogLevel(p2p.LevelInfo)
			case "warn":  p2p.SetLogLevel(p2p.LevelWarn)
			case "error": p2p.SetLogLevel(p2p.LevelError)
			case "none":  p2p.SetLogLevel(p2p.LevelNone)
			default: fmt.Println("Unknown level. Use: debug, info, warn, error, none")
			}
			fmt.Printf("Log level set to: %s\n", level)
		case "help":
			printHelp()
		case "quit", "exit":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Printf("Unknown command: %s. Type 'help' for commands.\n", command)
		}
	}
}

func printHeader(server *p2p.Server) {
	fmt.Printf("%s----------------------------------------%s\n", ColorBlue, ColorReset)
	fmt.Printf("%s  Simple Blockchain CLI (P2P Enabled)%s\n", ColorCyan, ColorReset)
	fmt.Printf("%s----------------------------------------%s\n", ColorBlue, ColorReset)
	fmt.Printf("Node ID: %s%s%s\n", ColorYellow, server.GetHostID(), ColorReset)
	fmt.Printf("Listening on: %s%v%s\n", ColorYellow, server.GetAddrs(), ColorReset)
	fmt.Printf("Type %s'help'%s for a list of commands.\n", ColorGreen, ColorReset)
	fmt.Println()
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Printf("  %sadd <bpm>%s       - Add a new block with given BPM\n", ColorCyan, ColorReset)
	fmt.Printf("  %sprint [n]%s       - Print last n blocks (default 10)\n", ColorCyan, ColorReset)
	fmt.Printf("  %svalidate%s        - Validate blockchain integrity\n", ColorCyan, ColorReset)
	fmt.Printf("  %slength%s          - Show blockchain length\n", ColorCyan, ColorReset)
	fmt.Printf("  %speers%s           - Show connected peers\n", ColorCyan, ColorReset)
	fmt.Printf("  %saddr%s             - Show your full addresses to share\n", ColorCyan, ColorReset)
	fmt.Printf("  %sconnect <addr>%s  - Connect to a peer via multiaddr\n", ColorCyan, ColorReset)
	fmt.Printf("  %ssync%s             - Request blocks from connected peers\n", ColorCyan, ColorReset)
	fmt.Printf("  %sdiscover%s         - Discover more peers\n", ColorCyan, ColorReset)
	fmt.Printf("  %sfind <hash>%s      - Find a block by hash\n", ColorCyan, ColorReset)
	fmt.Printf("  %slog <level>%s      - Set log level (debug, info, warn, error, none)\n", ColorCyan, ColorReset)
	fmt.Printf("  %shelp%s             - Show this help message\n", ColorCyan, ColorReset)
	fmt.Printf("  %squit%s             - Exit the program\n", ColorCyan, ColorReset)
}
