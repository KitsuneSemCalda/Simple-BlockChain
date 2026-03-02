package p2p_tests

import (
	"context"
	"KitsuneSemCalda/SBC/internal/blockchain"
	"KitsuneSemCalda/SBC/internal/p2p"
	"github.com/caiolandgraf/gest/gest"
	"time"
)

func init() {
	s := gest.Describe("P2P Server Core")

	s.It("should manage failed peers cache", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		cfg := p2p.DefaultConfig()
		server, _ := p2p.NewServer(cfg, bc)
		
		// Valid multihash for a peer ID
		addr := "/ip4/1.1.1.1/tcp/8333/p2p/12D3KooWLGpBKHhshKgGJiiKxKUPKJVBsE3w1td6jSPTquU1xCjk"
		
		// Attempt connection to non-existent but valid addr (will fail and mark)
		err := server.ConnectToPeer(addr)
		t.Expect(err).Not().ToBeNil()
		
		// Second attempt should return nil quickly (cached failure)
		start := time.Now()
		err = server.ConnectToPeer(addr)
		duration := time.Since(start)
		
		t.Expect(err).ToBeNil()
		t.Expect(duration < 100*time.Millisecond).ToBeTrue()
		
		server.Close()
	})

	s.It("should start maintenance tasks without crashing", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		cfg := p2p.DefaultConfig()
		server, _ := p2p.NewServer(cfg, bc)
		
		ctx, cancel := context.WithCancel(context.Background())
		server.StartMaintenance(ctx)
		
		time.Sleep(50 * time.Millisecond)
		cancel()
		
		t.Expect(true).ToBeTrue()
		server.Close()
	})

	gest.Register(s)
}
