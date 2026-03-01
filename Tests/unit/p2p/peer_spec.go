package p2p_tests

import (
	"KitsuneSemCalda/SBC/internal/p2p"
	"github.com/caiolandgraf/gest/gest"
	"github.com/libp2p/go-libp2p/core/peer"
	"fmt"
)

func init() {
	s := gest.Describe("P2P Peer")

	s.It("should represent peer correctly as string", func(t *gest.T) {
		pID, _ := peer.Decode("12D3KooWLGpBKHhshKgGJiiKxKUPKJVBsE3w1td6jSPTquU1xCjk")
		p := p2p.NewPeer(nil, pID)
		expected := fmt.Sprintf("Peer{%s}", pID)
		t.Expect(p.String()).ToBe(expected)
	})

	s.It("should store the correct peer ID", func(t *gest.T) {
		pID := peer.ID("test-peer")
		p := p2p.NewPeer(nil, pID)
		t.Expect(p.ID).ToBe(pID)
	})

	gest.Register(s)
}
