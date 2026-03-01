package p2p_tests

import (
	"KitsuneSemCalda/SBC/internal/p2p"
	"github.com/caiolandgraf/gest/gest"
	"encoding/json"
)

func init() {
	s := gest.Describe("P2P Message")

	s.It("should create a new message with payload", func(t *gest.T) {
		payload := p2p.VersionPayload{Version: 1, BestHeight: 10}
		msg, err := p2p.NewMessage(p2p.MsgVersion, payload)
		t.Expect(err).ToBeNil()
		t.Expect(msg.Type).ToBe(p2p.MsgVersion)
		
		var decodedPayload p2p.VersionPayload
		err = json.Unmarshal(msg.Payload, &decodedPayload)
		t.Expect(err).ToBeNil()
		t.Expect(decodedPayload.BestHeight).ToBe(10)
	})

	s.It("should encode message to bytes", func(t *gest.T) {
		msg := &p2p.Message{Type: p2p.MsgVerAck, Payload: []byte("{}")}
		bytes, err := msg.Encode()
		t.Expect(err).ToBeNil()
		t.Expect(len(bytes)).Not().ToBe(0)
	})

	s.It("should decode message from bytes", func(t *gest.T) {
		originalMsg := &p2p.Message{Type: p2p.MsgGetPeers, Payload: []byte("{}")}
		bytes, _ := originalMsg.Encode()
		
		decodedMsg, err := p2p.DecodeMessage(bytes)
		t.Expect(err).ToBeNil()
		t.Expect(decodedMsg.Type).ToBe(p2p.MsgGetPeers)
	})

	s.It("should handle empty payloads", func(t *gest.T) {
		msg, err := p2p.NewMessage(p2p.MsgVerAck, nil)
		t.Expect(err).ToBeNil()
		t.Expect(msg.Type).ToBe(p2p.MsgVerAck)
	})

	gest.Register(s)
}
