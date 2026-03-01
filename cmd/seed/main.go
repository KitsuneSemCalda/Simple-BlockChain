package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"KitsuneSemCalda/SBC/internal/blockchain"
	"KitsuneSemCalda/SBC/internal/p2p"
	"KitsuneSemCalda/SBC/internal/storage"
)

type SeedServer struct {
	peers map[string]time.Time
	mu    sync.Mutex
}

type SeedPeerInfo struct {
	Addr   string `json:"addr"`
	SeenAt int64  `json:"seen_at"`
}

type SeedResponse struct {
	Peers []SeedPeerInfo `json:"peers"`
}

func main() {
	bc := blockchain.NewBlockchain()
	cfg := p2p.DefaultConfig()
	cfg.ParseFlags()

	store, err := storage.NewStore(cfg.DataDir)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer store.Close()

	err = store.Load(bc)
	if err != nil {
		log.Fatalf("failed to load blockchain from store: %v", err)
	}

	server, err := p2p.NewServer(cfg, bc)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	seed := &SeedServer{
		peers: make(map[string]time.Time),
	}

	go func() {
		for {
			time.Sleep(30 * time.Second)
			seed.cleanupPeers()
		}
	}()

	go func() {
		for {
			time.Sleep(10 * time.Second)
			seed.updatePeers(server)
		}
	}()

	http.HandleFunc("/seeds", seed.handleSeeds)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	addr := ":8080"
	if env := os.Getenv("SEED_PORT"); env != "" {
		addr = ":" + env
	}

	log.Printf("Seed server starting on http://localhost%s", addr)
	log.Printf("Seed endpoint: http://localhost%s/seeds", addr)

	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil && err != http.ErrServerClosed {
			log.Fatalf("seed server error: %v", err)
		}
	}()

	for _, bootAddr := range cfg.BootNode {
		if bootAddr == "" {
			continue
		}
		err = server.ConnectToPeer(bootAddr)
		if err != nil {
			log.Printf("Can't connect to bootnode %s: %v", bootAddr, err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop
		err := store.Save(bc)
		if err != nil {
			log.Printf("failed to save blockchain: %v", err)
		} else {
			log.Println("blockchain saved successfully")
		}
		cancel()
	}()

	log.Printf("Daemon Node ID: %s", server.GetHostID())
	log.Printf("Listening on: %s", server.GetAddrs())

	server.Start(ctx)
}

func (s *SeedServer) updatePeers(server *p2p.Server) {
	peers := server.GetPeers()
	s.mu.Lock()
	defer s.mu.Unlock()

	for id := range peers {
		for _, addr := range server.GetAddrs() {
			peerAddr := addr.String() + "/p2p/" + id.String()
			s.peers[peerAddr] = time.Now()
		}
	}
}

func (s *SeedServer) cleanupPeers() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for addr, seen := range s.peers {
		if now.Sub(seen) > 5*time.Minute {
			delete(s.peers, addr)
		}
	}
}

func (s *SeedServer) handleSeeds(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")

	var peers []SeedPeerInfo
	for addr, seen := range s.peers {
		peers = append(peers, SeedPeerInfo{
			Addr:   addr,
			SeenAt: seen.Unix(),
		})
	}

	if len(peers) == 0 {
		log.Printf("[SEED] No peers available")
	} else {
		log.Printf("[SEED] Serving %d peers", len(peers))
	}

	json.NewEncoder(w).Encode(SeedResponse{Peers: peers})
}
