package p2p

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"KitsuneSemCalda/SBC/internal/p2p/callbacks"
	"KitsuneSemCalda/SBC/internal/structures"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type Server struct {
	host       *Host
	blockchain *structures.Blockchain
	peers      map[peer.ID]*Peer

	peer_callback  callbacks.PeerCallbacks
	block_callback callbacks.BlockCallbacks

	processedBlocks map[string]bool
}

func NewServer(cfg *Config, bc *structures.Blockchain) (*Server, error) {
	host, err := NewHost(cfg)
	if err != nil {
		return nil, err
	}

	s := &Server{
		host:            host,
		blockchain:      bc,
		peers:           make(map[peer.ID]*Peer),
		processedBlocks: make(map[string]bool),
	}

	s.host.SetStreamHandler(s.handleStream)

	s.blockchain.Subscribe(func(block *structures.Block) {
		if s.processedBlocks[block.Hash] {
			return
		}
		s.BroadcastBlock(block)
	})

	return s, nil
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
	case MsgBlock:
		var payload BlockPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Printf("Error unmarshaling block: %v", err)
			return
		}

		// Convert payload to Block for callback
		block := &structures.Block{
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
		s.processedBlocks[block.Hash] = true
		s.blockchain.ProcessBlock(block)
		log.Printf("Bloco #%d recebido e adicionado", payload.Index)
	case MsgGetBlocks:
		var payload GetBlocksPayload
		json.Unmarshal(msg.Payload, &payload)
		s.sendInv(p)
	}
}

func (s *Server) BroadcastBlock(block *structures.Block) {
	s.processedBlocks[block.Hash] = true
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
	return s.host.Connect(ctx, ma)
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
