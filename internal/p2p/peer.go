package p2p

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"KitsuneSemCalda/SBC/internal/structures"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type Server struct {
	host       *Host
	blockchain *structures.Blockchain
	peers      map[peer.ID]*Peer
}

func NewServer(cfg *Config, bc *structures.Blockchain) (*Server, error) {
	host, err := NewHost(cfg)
	if err != nil {
		return nil, err
	}

	s := &Server{
		host:       host,
		blockchain: bc,
		peers:      make(map[peer.ID]*Peer),
	}

	s.host.SetStreamHandler(s.handleStream)

	return s, nil
}

func (s *Server) handleStream(stream network.Stream) {
	peer := NewPeer(stream, stream.Conn().RemotePeer())
	s.peers[stream.Conn().RemotePeer()] = peer
	go s.readMessages(peer)
	s.sendVersion(peer)
}

func (s *Server) readMessages(p *Peer) {
	for {
		msg, err := p.ReadMessage()
		if err != nil {
			log.Printf("Erro ao ler mensagem de %s: %v", p.ID, err)
			delete(s.peers, p.ID)
			return
		}
		s.handleMessage(p, msg)
	}
}

func (s *Server) handleMessage(p *Peer, msg *Message) {
	switch msg.Type {
	case MsgVersion:
		var payload VersionPayload
		json.Unmarshal(msg.Payload, &payload)
		p.BestHeight = payload.BestHeight
		verAck, _ := NewMessage(MsgVerAck, VerAckPayload{Accept: true})
		p.SendMessage(verAck)
	case MsgVerAck:
		log.Printf("Handshake completo com peer: %s", p.ID)
	case MsgBlock:
		var block BlockPayload
		json.Unmarshal(msg.Payload, &block)
		s.blockchain.AddBlock(block.BPM)
		log.Printf("Bloco #%d recebido e adicionado", block.Index)
	case MsgGetBlocks:
		var payload GetBlocksPayload
		json.Unmarshal(msg.Payload, &payload)
		s.sendInv(p)
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

func (s *Server) Close() error {
	return s.host.Close()
}
