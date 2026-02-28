package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	for _, addr := range cfg.BootNode {
		if addr == "" {
			continue
		}
		err = server.ConnectToPeer(addr)
		if err != nil {
			log.Printf("Can't connect to server %s because %v", addr, err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop
		cancel()
	}()

	log.Printf("Daemon Node ID: %s", server.GetHostID())
	log.Printf("Listening on: %s", server.GetAddrs())

	err = server.Start(ctx)
	if err != nil {
		log.Printf("Can't start server because %v", err)
	}
}
