package p2p

import (
	"KitsuneSemCalda/SBC/internal/blockchain"
	"github.com/caiolandgraf/gest/gest"
)

func init() {
	s := gest.Describe("P2P Server")

	s.It("should create a P2P server with correct default config", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		cfg := DefaultConfig()
		server, err := NewServer(cfg, bc)

		t.Expect(err).ToBeNil()
		t.Expect(server).Not().ToBeNil()
		t.Expect(server.GetHostID()).Not().ToBe("")
	})

	s.It("should have a valid protocol ID", func(t *gest.T) {
		t.Expect(ProtocolID).Not().ToBe("")
		t.Expect(ProtocolID).ToContain("/sbc/")
	})

	gest.Register(s)
}
