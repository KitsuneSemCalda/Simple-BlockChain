package background_tests

import (
	"KitsuneSemCalda/SBC/internal/blockchain"
	"KitsuneSemCalda/SBC/internal/p2p"
	"github.com/caiolandgraf/gest/gest"
	"sync"
)

func init() {
	s := gest.Describe("P2P Background")

	s.It("should handle rapid block broadcasts concurrently", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		cfg := p2p.DefaultConfig()
		cfg.ListenAddr = "/ip4/127.0.0.1/tcp/0"
		
		server, _ := p2p.NewServer(cfg, bc)
		defer server.Close()
		
		var wg sync.WaitGroup
		numBroadcasters := 5
		
		for i := 0; i < numBroadcasters; i++ {
			wg.Add(1)
			go func(bpm int) {
				defer wg.Done()
				block := blockchain.NewBlock(1, bpm, bc.GetLastBlock().Hash)
				server.BroadcastBlock(block)
			}(100 + i)
		}
		
		wg.Wait()
		t.Expect(len(server.GetPeers())).ToBe(0) // No peers to broadcast to, but should not crash
	})

	gest.Register(s)
}
