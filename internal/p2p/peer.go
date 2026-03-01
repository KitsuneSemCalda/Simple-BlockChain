package p2p

import (
	"bufio"
	"encoding/json"
	"fmt"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Peer struct {
	ID         peer.ID
	Stream     network.Stream
	BestHeight int
}

func NewPeer(stream network.Stream, id peer.ID) *Peer {
	return &Peer{
		ID:     id,
		Stream: stream,
	}
}

func (p *Peer) SendMessage(msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = p.Stream.Write(append(data, '\n'))
	return err
}

func (p *Peer) ReadMessage() (*Message, error) {
	reader := bufio.NewReader(p.Stream)
	data, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	var msg Message
	err = json.Unmarshal(data, &msg)
	return &msg, err
}

func (p *Peer) String() string {
	return fmt.Sprintf("Peer{%s}", p.ID)
}
