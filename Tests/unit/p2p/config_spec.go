package p2p_tests

import (
	"KitsuneSemCalda/SBC/internal/p2p"
	"github.com/caiolandgraf/gest/gest"
	"os"
)

func init() {
	s := gest.Describe("P2P Config")

	s.It("should return a default data directory", func(t *gest.T) {
		dir := p2p.DefaultDataDir()
		t.Expect(dir).Not().ToBe("")
	})

	s.It("should return default bootstrap nodes", func(t *gest.T) {
		nodes := p2p.GetBootstrapNodes()
		t.Expect(len(nodes)).ToBe(0) // Default is empty as per current code
	})

	s.It("should return a default config object", func(t *gest.T) {
		cfg := p2p.DefaultConfig()
		t.Expect(cfg).Not().ToBeNil()
		t.Expect(cfg.ListenAddr).ToBe("/ip4/0.0.0.0/tcp/0")
		t.Expect(cfg.DiscoveryPort).ToBe(9999)
	})


	s.It("should handle bootstrap nodes from environment variable", func(t *gest.T) {
		os.Setenv("SBC_BOOTNODES", "node1,node2")
		defer os.Unsetenv("SBC_BOOTNODES")
		
		nodes := p2p.GetBootstrapNodes()
		t.Expect(len(nodes)).ToBe(2)
		t.Expect(nodes[0]).ToBe("node1")
		t.Expect(nodes[1]).ToBe("node2")
	})

	gest.Register(s)
}
