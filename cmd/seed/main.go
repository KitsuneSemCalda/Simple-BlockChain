package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type PeerInfo struct {
	Addr     string    `json:"addr"`
	LastSeen time.Time `json:"last_seen"`
}

var (
	peers   = make(map[string]PeerInfo)
	peersMu sync.RWMutex
)

func main() {
	port := os.Getenv("SEED_PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/seeds", handleSeeds)
	http.HandleFunc("/announce", handleAnnounce)
	http.HandleFunc("/info", handleInfo)

	go cleanupPeers()

	fmt.Printf("[SEED] Server starting on :%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start seed server: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func handleSeeds(w http.ResponseWriter, r *http.Request) {
	peersMu.RLock()
	defer peersMu.RUnlock()

	var list []map[string]string
	for addr := range peers {
		list = append(list, map[string]string{"addr": addr})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"peers": list})
}

func handleAnnounce(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Addr string `json:"addr"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Addr != "" {
		peersMu.Lock()
		peers[req.Addr] = PeerInfo{
			Addr:     req.Addr,
			LastSeen: time.Now(),
		}
		peersMu.Unlock()
		fmt.Printf("[SEED] Peer announced: %s\n", req.Addr)
	}

	w.WriteHeader(http.StatusOK)
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	peersMu.RLock()
	count := len(peers)
	peersMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "online",
		"peers_count": count,
		"uptime": time.Since(time.Now()), // Just a placeholder
	})
}

func cleanupPeers() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		now := time.Now()
		peersMu.Lock()
		for addr, info := range peers {
			if now.Sub(info.LastSeen) > 30*time.Minute {
				delete(peers, addr)
				fmt.Printf("[SEED] Cleaned up inactive peer: %s\n", addr)
			}
		}
		peersMu.Unlock()
	}
}
