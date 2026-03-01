package p2p

import (
	"encoding/json"
	"time"
)

type MessageType string

const (
	MsgVersion    MessageType = "version"
	MsgVerAck     MessageType = "verack"
	MsgGetBlocks  MessageType = "getblocks"
	MsgBlock      MessageType = "block"
	MsgInv        MessageType = "inv"
	MsgGetData    MessageType = "getdata"
	MsgPing       MessageType = "ping"
	MsgPong       MessageType = "pong"
	MsgGetPeers   MessageType = "getpeers"
	MsgPeers      MessageType = "peers"
	MsgFindBlock  MessageType = "findblock"
	MsgBlockFound MessageType = "blockfound"
)

type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type VersionPayload struct {
	Version    int       `json:"version"`
	Timestamp  time.Time `json:"timestamp"`
	AddrRecv   string    `json:"addr_recv"`
	AddrFrom   string    `json:"addr_from"`
	UserAgent  string    `json:"user_agent"`
	BestHeight int       `json:"best_height"`
}

type VerAckPayload struct {
	Accept bool `json:"accept"`
}

type BlockPayload struct {
	Index     int       `json:"index"`
	Timestamp time.Time `json:"timestamp"`
	BPM       int       `json:"bpm"`
	Hash      string    `json:"hash"`
	PrevHash  string    `json:"prev_hash"`
}

type GetBlocksPayload struct {
	StartHash string `json:"start_hash"`
	StopHash  string `json:"stop_hash"`
}

type InvPayload struct {
	Count  int      `json:"count"`
	InvVec []InvVec `json:"inv_vec"`
}

type InvVec struct {
	Type string `json:"type"` // "block"
	Hash string `json:"hash"`
}

type GetPeersPayload struct{}

type PeersPayload struct {
	Peers []string `json:"peers"`
}

type FindBlockPayload struct {
	Hash string `json:"hash"`
}

type BlockFoundPayload struct {
	Found bool          `json:"found"`
	Block *BlockPayload `json:"block,omitempty"`
}

func NewMessage(msgType MessageType, payload any) (*Message, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &Message{
		Type:    msgType,
		Payload: payloadBytes,
	}, nil
}

func (m *Message) Encode() ([]byte, error) {
	return json.Marshal(m)
}

func DecodeMessage(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	return &msg, err
}
