package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"KitsuneSemCalda/SBC/internal/blockchain"
	"KitsuneSemCalda/SBC/internal/p2p"
	"KitsuneSemCalda/SBC/internal/storage"

	"github.com/libp2p/go-libp2p/core/peer"
)

var logger *slog.Logger

type DaemonCallbacks struct {
	blockchain *blockchain.Blockchain
	server     *p2p.Server
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (c *DaemonCallbacks) OnNewPeer(id peer.ID) {
	p2p.Info("SBCD", "new peer connected: %s", id.String()[:min(16, len(id.String()))])
}

func (c *DaemonCallbacks) OnDisconnect(id peer.ID) {
	p2p.Warn("SBCD", "peer disconnected: %s", id.String()[:min(16, len(id.String()))])
}

func (c *DaemonCallbacks) OnPeerFound(info peer.AddrInfo) {
	p2p.Debug("SBCD", "peer found via discovery: %s", info.ID.String()[:min(16, len(info.ID.String()))])
}

func (c *DaemonCallbacks) OnBlockReceived(block *blockchain.Block) {
	hash := block.Hash
	if len(hash) > 8 {
		hash = hash[:8]
	}
	logger.Info("block received", "index", block.Index, "hash", hash)
	logger.Debug("blockchain status", "length", c.blockchain.Length())
}

func initLogger() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	// Use colored output for terminal
	handler := &ColoredHandler{
		Handler: slog.NewTextHandler(os.Stdout, opts),
	}
	logger = slog.New(handler)
}

type ColoredHandler struct {
	slog.Handler
}

func (h *ColoredHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String()
	var colorCode string
	switch r.Level {
	case slog.LevelDebug:
		colorCode = "\033[36m" // Cyan
	case slog.LevelInfo:
		colorCode = "\033[32m" // Green
	case slog.LevelWarn:
		colorCode = "\033[33m" // Yellow
	case slog.LevelError:
		colorCode = "\033[31m" // Red
	}

	fmt.Printf("%s[%s]\033[0m %s ", colorCode, level, r.Message)
	r.Attrs(func(a slog.Attr) bool {
		fmt.Printf("\033[90m%s=%v\033[0m ", a.Key, a.Value)
		return true
	})
	fmt.Println()
	return nil
}

func main() {
	// Daemon defaults to Info level
	p2p.SetLogLevel(p2p.LevelInfo)
	p2p.Info("SBCD", "starting sbc daemon")

	bc := blockchain.NewBlockchain()
	cfg := p2p.DefaultConfig()
	// Daemon should default to 8333
	cfg.ListenAddr = "/ip4/0.0.0.0/tcp/8333"
	cfg.ParseFlags()

	store, err := storage.NewStore(cfg.DataDir)
	if err != nil {
		p2p.Error("SBCD", "failed to open database: %v", err)
		os.Exit(1)
	}
	defer store.Close()

	err = store.Load(bc)
	if err != nil {
		p2p.Error("SBCD", "failed to load blockchain from store: %v", err)
		os.Exit(1)
	}

	server, err := p2p.NewServer(cfg, bc)
	if err != nil {
		p2p.Error("SBCD", "failed to create server: %v", err)
		os.Exit(1)
	}

	cbs := &DaemonCallbacks{blockchain: bc, server: server}
	server.SetPeerCallback(cbs)
	server.SetBlockCallback(cbs)

	for _, addr := range cfg.BootNode {
		if addr == "" {
			continue
		}
		err = server.ConnectToPeer(addr)
		if err != nil {
			p2p.Warn("SBCD", "can't connect to bootnode %s: %v", addr, err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server.StartMaintenance(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop
		err := store.Save(bc)
		if err != nil {
			p2p.Error("SBCD", "failed to save blockchain: %v", err)
		} else {
			p2p.Info("SBCD", "blockchain saved successfully")
		}
		cancel()
	}()

	p2p.Info("SBCD", "daemon initialized: peer_id=%s listening=%v height=%d",
		server.GetHostID(), server.GetAddrs(), bc.Length())

	announceData := map[string]string{
		"peer_id": server.GetHostID(),
		"addr":    "/ip4/0.0.0.0/tcp/8333",
	}
	announceBytes, _ := json.Marshal(announceData)
	os.WriteFile("/tmp/sbc-daemon.json", announceBytes, 0o644)
	logger.Info("announce file written", "path", "/tmp/sbc-daemon.json")

	err = server.Start(ctx)
	if err != nil {
		p2p.Error("SBCD", "server error: %v", err)
	}
}
