package e2e_tests

import (
	"KitsuneSemCalda/SBC/internal/blockchain"
	"KitsuneSemCalda/SBC/internal/p2p"
	"github.com/caiolandgraf/gest/gest"
	"time"
)

func init() {
	s := gest.Describe("P2P E2E")

	s.It("should synchronize blocks between two nodes", func(t *gest.T) {
		bc1 := blockchain.NewBlockchain()
		bc2 := blockchain.NewBlockchain()
		
		cfg1 := p2p.DefaultConfig()
		cfg1.ListenAddr = "/ip4/127.0.0.1/tcp/0"
		
		server1, _ := p2p.NewServer(cfg1, bc1)
		
		cfg2 := p2p.DefaultConfig()
		cfg2.ListenAddr = "/ip4/127.0.0.1/tcp/0"
		server2, _ := p2p.NewServer(cfg2, bc2)
		
		// Add block to node 1
		bc1.AddBlock(100)
		t.Expect(bc1.Length()).ToBe(2)
		
		// Manually connect server2 to server1
		addr1 := server1.GetAddrs()[0].String() + "/p2p/" + server1.GetHostID()
		err := server2.ConnectToPeer(addr1)
		t.Expect(err).ToBeNil()
		
		// Give some time for the handshake to complete asynchronously
		time.Sleep(100 * time.Millisecond)
		
		// Verify they are connected
		t.Expect(len(server1.GetPeers())).Not().ToBe(0)
		t.Expect(len(server2.GetPeers())).Not().ToBe(0)
		
		server1.Close()
		server2.Close()
	})

	gest.Register(s)
}
