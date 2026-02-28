package p2p

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"KitsuneSemCalda/SBC/internal/blockchain"
	"KitsuneSemCalda/SBC/internal/p2p/callbacks"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type Server struct {
	host       *Host
	blockchain *blockchain.Blockchain
	peers      map[peer.ID]*Peer

	peer_callback  callbacks.PeerCallbacks
	block_callback callbacks.BlockCallbacks

	processedMutex  sync.RWMutex
	processedBlocks map[string]time.Time
}

func NewServer(cfg *Config, bc *blockchain.Blockchain) (*Server, error) {
	host, err := NewHost(cfg)
	if err != nil {
		return nil, err
	}

	s := &Server{
		host:            host,
		blockchain:      bc,
		peers:           make(map[peer.ID]*Peer),
		processedBlocks: make(map[string]time.Time),
	}

	s.host.SetStreamHandler(s.handleStream)

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

	return s, nil
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
	peer := NewPeer(stream, pID)
	s.peers[pID] = peer

	if s.peer_callback != nil {
		s.peer_callback.OnNewPeer(pID)
	}

	go s.readMessages(peer)
	s.sendVersion(peer)
}

func (s *Server) readMessages(p *Peer) {
	defer func() {
		delete(s.peers, p.ID)
		if s.peer_callback != nil {
			s.peer_callback.OnDisconnect(p.ID)
		}
	}()

	for {
		msg, err := p.ReadMessage()
		if err != nil {
			log.Printf("Erro ao ler mensagem de %s: %v", p.ID, err)
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
			log.Printf("Error unmarshaling version: %v", err)
			return
		}
		p.BestHeight = payload.BestHeight
		verAck, _ := NewMessage(MsgVerAck, VerAckPayload{Accept: true})
		p.SendMessage(verAck)
	case MsgVerAck:
		log.Printf("Handshake completo com peer: %s", p.ID)
		if p.BestHeight > s.blockchain.Length() {
			s.sendGetBlocks(p)
		}
	case MsgBlock:
		var payload BlockPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Printf("Error unmarshaling block: %v", err)
			return
		}

		// Convert payload to Block for callback
		block := &blockchain.Block{
			Index:     payload.Index,
			Timestamp: payload.Timestamp,
			BPM:       payload.BPM,
			Hash:      payload.Hash,
			PrevHash:  payload.PrevHash,
		}

		if s.block_callback != nil {
			s.block_callback.OnBlockReceived(block)
		}

		// Avoid adding block if it would trigger another broadcast (loop)
		s.processedMutex.Lock()
		s.processedBlocks[block.Hash] = time.Now()
		s.processedMutex.Unlock()
		s.blockchain.ProcessBlock(block)
		log.Printf("Bloco #%d recebido e adicionado", payload.Index)
	case MsgGetBlocks:
		var payload GetBlocksPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Printf("Error unmarshaling getblocks: %v", err)
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
			log.Printf("Error unmarshaling inv: %v", err)
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
			log.Printf("Error unmarshaling getdata: %v", err)
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

	for _, p := range s.peers {
		if err := p.SendMessage(msg); err != nil {
			log.Printf("Error broadcasting to %s: %v", p.ID, err)
		}
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
}

func (s *Server) ConnectToPeer(addr string) error {
	ctx := context.Background()
	ma, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return err
	}

	err = s.host.Connect(ctx, ma)
	if err != nil {
		return err
	}

	pi, err := peer.AddrInfoFromP2pAddr(ma)
	if err != nil {
		return err
	}

	stream, err := s.host.NewStream(ctx, pi.ID)
	if err != nil {
		return err
	}

	s.handleStream(stream)
	return nil
}

func (s *Server) Start(ctx context.Context) error {
	log.Printf("P2P Server started on: %s", s.host.Addrs())
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
