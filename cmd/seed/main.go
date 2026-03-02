package main

import (
	"context"
	"encoding/json"
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

	// Seed server usually wants to see what's happening
	p2p.SetLogLevel(p2p.LevelInfo)

	store, err := storage.NewStore(cfg.DataDir)
	if err != nil {
		p2p.Error("SEED", "failed to open database: %v", err)
		os.Exit(1)
	}
	defer store.Close()

	err = store.Load(bc)
	if err != nil {
		p2p.Error("SEED", "failed to load blockchain from store: %v", err)
		os.Exit(1)
	}

	server, err := p2p.NewServer(cfg, bc)
	if err != nil {
		p2p.Error("SEED", "failed to create server: %v", err)
		os.Exit(1)
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

	p2p.Info("SEED", "Seed server starting on http://localhost%s", addr)
	p2p.Info("SEED", "Seed endpoint: http://localhost%s/seeds", addr)

	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil && err != http.ErrServerClosed {
			p2p.Error("SEED", "seed server error: %v", err)
		}
	}()

	for _, bootAddr := range cfg.BootNode {
		if bootAddr == "" {
			continue
		}
		err = server.ConnectToPeer(bootAddr)
		if err != nil {
			p2p.Warn("SEED", "Can't connect to bootnode %s: %v", bootAddr, err)
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
			p2p.Error("SEED", "failed to save blockchain: %v", err)
		} else {
			p2p.Info("SEED", "blockchain saved successfully")
		}
		cancel()
	}()

	p2p.Info("SEED", "Daemon Node ID: %s", server.GetHostID())
	p2p.Info("SEED", "Listening on: %s", server.GetAddrs())

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
		p2p.Debug("SEED", "No peers available")
	} else {
		p2p.Debug("SEED", "Serving %d peers", len(peers))
	}

	json.NewEncoder(w).Encode(SeedResponse{Peers: peers})
}
