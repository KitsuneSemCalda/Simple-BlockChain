package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"KitsuneSemCalda/SBC/internal/blockchain"
	"KitsuneSemCalda/SBC/internal/p2p/callbacks"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

const FailedPeerCooldown = 10 * time.Minute

type Server struct {
	host       *Host
	blockchain *blockchain.Blockchain
	peers      map[peer.ID]*Peer
	peerMutex  sync.RWMutex
	config     *Config

	peer_callback  callbacks.PeerCallbacks
	block_callback callbacks.BlockCallbacks

	processedMutex  sync.RWMutex
	processedBlocks map[string]time.Time

	failedPeers   map[string]time.Time
	failedPeersMu sync.RWMutex
}

func NewServer(cfg *Config, bc *blockchain.Blockchain) (*Server, error) {
	s := &Server{
		blockchain:      bc,
		peers:           make(map[peer.ID]*Peer),
		config:          cfg,
		processedBlocks: make(map[string]time.Time),
		failedPeers:     make(map[string]time.Time),
	}

	host, err := NewHost(cfg, s)
	if err != nil {
		return nil, err
	}
	s.host = host

	s.host.SetStreamHandler(s.handleStream)

	go func() {
		time.Sleep(2 * time.Second)
		s.tryAutoConnect()
	}()

	s.blockchain.Subscribe(func(block *blockchain.Block) {
		s.processedMutex.Lock()
		if !s.processedBlocks[block.Hash].IsZero() {
			s.processedMutex.Unlock()
			return
		}
		s.processedBlocks[block.Hash] = time.Now()
		s.processedMutex.Unlock()
		s.BroadcastBlock(block)
	})

	go s.cleanupProcessedBlocks()
	go s.cleanupFailedPeers()

	return s, nil
}

func (s *Server) cleanupFailedPeers() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		s.failedPeersMu.Lock()
		for addr, timestamp := range s.failedPeers {
			if now.Sub(timestamp) > FailedPeerCooldown {
				delete(s.failedPeers, addr)
			}
		}
		s.failedPeersMu.Unlock()
	}
}

func (s *Server) StartMaintenance(ctx context.Context) {
	go s.syncTask(ctx)
	go s.periodicPeerDiscovery()
}

func (s *Server) syncTask(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			peers := s.GetPeers()
			if len(peers) == 0 {
				continue
			}

			var maxHeight int
			var bestPeer peer.ID
			for pID, p := range peers {
				if p.BestHeight > maxHeight {
					maxHeight = p.BestHeight
					bestPeer = pID
				}
			}

			if maxHeight > s.blockchain.Length() {
				Debug("Sync", "Peer has longer chain, requesting blocks")
				msg, _ := NewMessage(MsgGetBlocks, GetBlocksPayload{StartHash: s.blockchain.GetLastBlock().Hash})
				if p, ok := peers[bestPeer]; ok {
					p.SendMessage(msg)
				}
			}
		}
	}
}

func (s *Server) cleanupProcessedBlocks() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		s.processedMutex.Lock()
		for hash, timestamp := range s.processedBlocks {
			if now.Sub(timestamp) > 10*time.Minute {
				delete(s.processedBlocks, hash)
			}
		}
		s.processedMutex.Unlock()
	}
}

func (s *Server) handleStream(stream network.Stream) {
	pID := stream.Conn().RemotePeer()

	s.peerMutex.Lock()
	if _, exists := s.peers[pID]; exists {
		s.peerMutex.Unlock()
		stream.Reset()
		return
	}

	peer := NewPeer(stream, pID)
	s.peers[pID] = peer
	s.peerMutex.Unlock()

	if s.peer_callback != nil {
		s.peer_callback.OnNewPeer(pID)
	}

	go s.readMessages(peer)
	s.sendVersion(peer)
}

func (s *Server) readMessages(p *Peer) {
	defer func() {
		s.peerMutex.Lock()
		delete(s.peers, p.ID)
		s.peerMutex.Unlock()

		if s.peer_callback != nil {
			s.peer_callback.OnDisconnect(p.ID)
		}
	}()

	for {
		msg, err := p.ReadMessage()
		if err != nil {
			return
		}
		s.handleMessage(p, msg)
	}
}

func (s *Server) handleMessage(p *Peer, msg *Message) {
	switch msg.Type {
	case MsgVersion:
		var payload VersionPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			Debug("P2P", "Error unmarshaling version: %v", err)
			return
		}
		p.BestHeight = payload.BestHeight
		verAck, _ := NewMessage(MsgVerAck, VerAckPayload{Accept: true})
		p.SendMessage(verAck)
	case MsgVerAck:
		Debug("P2P", "Handshake complete with %s", p.ID)
		if p.BestHeight > s.blockchain.Length() {
			s.sendGetBlocks(p)
		}
		msg, _ := NewMessage(MsgGetPeers, GetPeersPayload{})
		p.SendMessage(msg)
	case MsgBlock:
		var payload BlockPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}

		block := &blockchain.Block{
			Index:     payload.Index,
			Timestamp: payload.Timestamp,
			BPM:       payload.BPM,
			Hash:      payload.Hash,
			PrevHash:  payload.PrevHash,
		}

		s.processedMutex.RLock()
		if !s.processedBlocks[block.Hash].IsZero() {
			s.processedMutex.RUnlock()
			return
		}
		s.processedMutex.RUnlock()

		if s.block_callback != nil {
			s.block_callback.OnBlockReceived(block)
		}

		s.blockchain.ProcessBlock(block)
		s.processedMutex.Lock()
		s.processedBlocks[block.Hash] = time.Now()
		s.processedMutex.Unlock()
	case MsgGetBlocks:
		var payload GetBlocksPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}

		blocks := s.blockchain.GetBlocksAfter(payload.StartHash, 500)
		if len(blocks) > 0 {
			var invVecs []InvVec
			for _, b := range blocks {
				invVecs = append(invVecs, InvVec{Type: "block", Hash: b.Hash})
			}
			inv := InvPayload{Count: len(invVecs), InvVec: invVecs}
			msg, _ := NewMessage(MsgInv, inv)
			p.SendMessage(msg)
		}
	case MsgInv:
		var payload InvPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}

		var missingVecs []InvVec
		s.processedMutex.RLock()
		for _, vec := range payload.InvVec {
			if vec.Type == "block" && s.processedBlocks[vec.Hash].IsZero() {
				missingVecs = append(missingVecs, vec)
			}
		}
		s.processedMutex.RUnlock()

		if len(missingVecs) > 0 {
			msg, _ := NewMessage(MsgGetData, InvPayload{
				Count:  len(missingVecs),
				InvVec: missingVecs,
			})
			p.SendMessage(msg)
		}
	case MsgGetData:
		var payload InvPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}

		for _, vec := range payload.InvVec {
			if vec.Type == "block" {
				block := s.blockchain.GetBlockByHash(vec.Hash)
				if block != nil {
					bPayload := BlockPayload{
						Index:     block.Index,
						Timestamp: block.Timestamp,
						BPM:       block.BPM,
						Hash:      block.Hash,
						PrevHash:  block.PrevHash,
					}
					msg, _ := NewMessage(MsgBlock, bPayload)
					p.SendMessage(msg)
				}
			}
		}
	case MsgGetPeers:
		var peers []string
		for _, peer := range s.peers {
			for _, addr := range s.host.Addrs() {
				peers = append(peers, addr.String()+"/p2p/"+peer.ID.String())
			}
		}
		peersMsg, _ := NewMessage(MsgPeers, PeersPayload{Peers: peers})
		p.SendMessage(peersMsg)
	case MsgPeers:
		var payload PeersPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		for _, addr := range payload.Peers {
			if addr == "" {
				continue
			}
			s.ConnectToPeer(addr)
		}
	case MsgFindBlock:
		var payload FindBlockPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		block := s.blockchain.GetBlockByHash(payload.Hash)
		if block != nil {
			bPayload := BlockPayload{
				Index:     block.Index,
				Timestamp: block.Timestamp,
				BPM:       block.BPM,
				Hash:      block.Hash,
				PrevHash:  block.PrevHash,
			}
			resp, _ := NewMessage(MsgBlockFound, BlockFoundPayload{Found: true, Block: &bPayload})
			p.SendMessage(resp)
		} else {
			resp, _ := NewMessage(MsgBlockFound, BlockFoundPayload{Found: false})
			p.SendMessage(resp)
		}
	case MsgBlockFound:
		var payload BlockFoundPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		if payload.Found && payload.Block != nil {
			block := &blockchain.Block{
				Index:     payload.Block.Index,
				Timestamp: payload.Block.Timestamp,
				BPM:       payload.Block.BPM,
				Hash:      payload.Block.Hash,
				PrevHash:  payload.Block.PrevHash,
			}
			s.blockchain.ProcessBlock(block)
			Debug("Discovery", "Block #%d found via network", block.Index)
		}
	}
}

func (s *Server) BroadcastBlock(block *blockchain.Block) {
	s.processedMutex.Lock()
	s.processedBlocks[block.Hash] = time.Now()
	s.processedMutex.Unlock()
	payload := BlockPayload{
		Index:     block.Index,
		Timestamp: block.Timestamp,
		BPM:       block.BPM,
		Hash:      block.Hash,
		PrevHash:  block.PrevHash,
	}
	msg, _ := NewMessage(MsgBlock, payload)

	s.peerMutex.RLock()
	defer s.peerMutex.RUnlock()
	for _, p := range s.peers {
		p.SendMessage(msg)
	}
}

func (s *Server) sendVersion(p *Peer) {
	payload := VersionPayload{
		Version:    1,
		Timestamp:  time.Now(),
		BestHeight: s.blockchain.Length(),
	}
	msg, _ := NewMessage(MsgVersion, payload)
	p.SendMessage(msg)
}

func (s *Server) sendGetBlocks(p *Peer) {
	lastBlock := s.blockchain.GetLastBlock()
	payload := GetBlocksPayload{
		StartHash: lastBlock.Hash,
	}
	msg, _ := NewMessage(MsgGetBlocks, payload)
	p.SendMessage(msg)
}

func (s *Server) sendInv(p *Peer) {
	lastBlock := s.blockchain.GetLastBlock()
	inv := InvPayload{
		Count: 1,
		InvVec: []InvVec{
			{Type: "block", Hash: lastBlock.Hash},
		},
	}
	msg, _ := NewMessage(MsgInv, inv)
	p.SendMessage(msg)
}

func (s *Server) SetBlockCallback(cb callbacks.BlockCallbacks) {
	s.block_callback = cb
}

func (s *Server) SetPeerCallback(cb callbacks.PeerCallbacks) {
	s.peer_callback = cb
	if s.host != nil {
		s.host.SetPeerCallback(s)
	}
}

func (s *Server) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == s.host.ID() {
		return
	}

	s.peerMutex.RLock()
	if _, connected := s.peers[pi.ID]; connected {
		s.peerMutex.RUnlock()
		return
	}
	s.peerMutex.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.host.ConnectPeer(ctx, pi); err != nil {
		return
	}
	stream, err := s.host.NewStream(ctx, pi.ID)
	if err != nil {
		return
	}
	s.handleStream(stream)
}

func (s *Server) OnPeerFound(pi peer.AddrInfo) {
	s.HandlePeerFound(pi)
}

func (s *Server) isFailed(addr string) bool {
	s.failedPeersMu.RLock()
	defer s.failedPeersMu.RUnlock()
	lastFailed, exists := s.failedPeers[addr]
	if !exists {
		return false
	}
	return time.Since(lastFailed) < FailedPeerCooldown
}

func (s *Server) markFailed(addr string) {
	s.failedPeersMu.Lock()
	defer s.failedPeersMu.Unlock()
	s.failedPeers[addr] = time.Now()
}

func (s *Server) ConnectToPeer(addr string) error {
	if s.isFailed(addr) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ma, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return err
	}

	if err := s.host.Connect(ctx, ma); err != nil {
		s.markFailed(addr)
		return err
	}

	pi, err := peer.AddrInfoFromP2pAddr(ma)
	if err != nil {
		return err
	}

	stream, err := s.host.NewStream(ctx, pi.ID)
	if err != nil {
		s.markFailed(addr)
		return err
	}

	s.handleStream(stream)
	return nil
}

func (s *Server) Start(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (s *Server) GetHostID() string {
	return s.host.ID().String()
}

func (s *Server) GetAddrs() []multiaddr.Multiaddr {
	return s.host.Addrs()
}

func (s *Server) Close() error {
	return s.host.Close()
}

func (s *Server) GetPeers() map[peer.ID]*Peer {
	s.peerMutex.RLock()
	defer s.peerMutex.RUnlock()

	peersCopy := make(map[peer.ID]*Peer)
	for id, p := range s.peers {
		peersCopy[id] = p
	}
	return peersCopy
}

func (s *Server) RequestSync() {
	s.peerMutex.RLock()
	defer s.peerMutex.RUnlock()
	for _, p := range s.peers {
		msg, _ := NewMessage(MsgGetBlocks, GetBlocksPayload{StartHash: ""})
		p.SendMessage(msg)
	}
}

func (s *Server) DiscoverPeers() {
	s.peerMutex.RLock()
	defer s.peerMutex.RUnlock()
	for _, p := range s.peers {
		msg, _ := NewMessage(MsgGetPeers, GetPeersPayload{})
		p.SendMessage(msg)
	}
}

func (s *Server) FindBlock(hash string) {
	s.peerMutex.RLock()
	defer s.peerMutex.RUnlock()
	for _, p := range s.peers {
		msg, _ := NewMessage(MsgFindBlock, FindBlockPayload{Hash: hash})
		p.SendMessage(msg)
	}
}

func (s *Server) tryAutoConnect() {
	s.tryLocalDiscovery()

	if len(s.config.BootNode) > 0 {
		go s.connectToBootNodes()
	}

	if s.config.DNSSeed != "" {
		go s.resolveDNSSeeds(s.config.DNSSeed)
		seeds := strings.Split(s.config.DNSSeed, ",")
		for _, seed := range seeds {
			seed = strings.TrimSpace(seed)
			if seed == "" {
				continue
			}
			go s.fetchSeedsFromHTTP(fmt.Sprintf("http://%s:8080/seeds", seed))
		}
	}

	go s.periodicPeerDiscovery()
}

func (s *Server) connectToBootNodes() {
	for _, addr := range s.config.BootNode {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}

		if s.isFailed(addr) {
			continue
		}

		ma, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := s.host.Connect(ctx, ma); err != nil {
			cancel()
			s.markFailed(addr)
			continue
		}
		cancel()

		pi, _ := peer.AddrInfoFromP2pAddr(ma)
		stream, err := s.host.NewStream(context.Background(), pi.ID)
		if err != nil {
			s.markFailed(addr)
			continue
		}

		s.handleStream(stream)
		Info("BOOT", "Connected to bootnode: %s", addr)
	}
}

func (s *Server) periodicPeerDiscovery() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		<-ticker.C
		if len(s.peers) > 0 {
			Debug("Discovery", "Periodic peer discovery (connected peers: %d)", len(s.peers))
			s.DiscoverPeers()
		}
	}
}

func (s *Server) tryLocalDiscovery() {
	announceFile := "/tmp/sbc-daemon.json"
	data, err := os.ReadFile(announceFile)
	if err != nil {
		return
	}

	var daemon struct {
		Addr   string `json:"addr"`
		PeerID string `json:"peer_id"`
	}
	if err := json.Unmarshal(data, &daemon); err != nil {
		return
	}

	if daemon.Addr != "" && daemon.PeerID != "" {
		daemonAddr := daemon.Addr + "/p2p/" + daemon.PeerID

		if s.isFailed(daemonAddr) {
			return
		}

		localIP := s.getLocalIP()
		if localIP != "" {
			daemonAddr = strings.Replace(daemonAddr, "/ip4/0.0.0.0/", "/ip4/"+localIP+"/", 1)
			daemonAddr = strings.Replace(daemonAddr, "/ip4/127.0.0.1/", "/ip4/"+localIP+"/", 1)
		}

		ma, err := multiaddr.NewMultiaddr(daemonAddr)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := s.host.Connect(ctx, ma); err == nil {
				cancel()
				pi, _ := peer.AddrInfoFromP2pAddr(ma)
				if stream, err := s.host.NewStream(context.Background(), pi.ID); err == nil {
					s.handleStream(stream)
					Info("AUTO", "Connected to local daemon: %s", daemonAddr)
					return
				}
			} else {
				cancel()
				s.markFailed(daemonAddr)
			}
		}
	}
}

func (s *Server) getLocalIP() string {
	addrs := s.host.Addrs()
	for _, addr := range addrs {
		parts := strings.Split(addr.String(), "/")
		for _, part := range parts {
			if ip := net.ParseIP(part); ip != nil && ip.To4() != nil {
				return part
			}
		}
	}
	return ""
}

func (s *Server) resolveDNSSeeds(dnsSeed string) {
	seeds := strings.Split(dnsSeed, ",")
	for _, seed := range seeds {
		seed = strings.TrimSpace(seed)
		if seed == "" {
			continue
		}

		ips, err := net.LookupIP(seed)
		if err != nil {
			continue
		}

		for _, ip := range ips {
			addr := fmt.Sprintf("/ip4/%s/tcp/8333", ip.String())
			if s.isFailed(addr) {
				continue
			}

			ma, err := multiaddr.NewMultiaddr(addr)
			if err != nil {
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := s.host.Connect(ctx, ma); err != nil {
				cancel()
				s.markFailed(addr)
				continue
			}
			cancel()

			pi, _ := peer.AddrInfoFromP2pAddr(ma)
			stream, err := s.host.NewStream(context.Background(), pi.ID)
			if err != nil {
				s.markFailed(addr)
				continue
			}

			s.handleStream(stream)
			Info("DNS", "Connected to bootstrap node: %s", addr)
			return
		}
	}
}

func (s *Server) fetchSeedsFromHTTP(seedURL string) {
	resp, err := http.Get(seedURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var result struct {
		Peers []struct {
			Addr string `json:"addr"`
		} `json:"peers"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return
	}

	for _, p := range result.Peers {
		if p.Addr == "" || s.isFailed(p.Addr) {
			continue
		}

		ma, err := multiaddr.NewMultiaddr(p.Addr)
		if err != nil {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := s.host.Connect(ctx, ma); err != nil {
			cancel()
			s.markFailed(p.Addr)
			continue
		}
		cancel()

		pi, _ := peer.AddrInfoFromP2pAddr(ma)
		stream, err := s.host.NewStream(context.Background(), pi.ID)
		if err != nil {
			s.markFailed(p.Addr)
			continue
		}

		s.handleStream(stream)
		Info("HTTP", "Connected to peer: %s", p.Addr)
		return
	}
}
