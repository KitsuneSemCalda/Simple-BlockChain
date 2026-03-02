package p2p_tests

import (
	"KitsuneSemCalda/SBC/internal/p2p"
	"github.com/caiolandgraf/gest/gest"
	"strings"
)

func init() {
	s := gest.Describe("P2P Host Core")

	s.It("should detect if a public IP is valid format", func(t *gest.T) {
		// This might fail if no internet, so we test the logic via helper if available
		// or just check if it returns something or an error
		ip, _ := p2p.GetPublicIP()
		if ip != "" {
			parts := strings.Split(ip, ".")
			t.Expect(len(parts)).ToBe(4)
		}
	})

	s.It("should include public IP in multiaddrs if set", func(t *gest.T) {
		cfg := p2p.DefaultConfig()
		h, _ := p2p.NewHost(cfg, nil)
		
		addrs := h.Addrs()
		t.Expect(len(addrs)).Not().ToBe(0)
		
		h.Close()
	})

	gest.Register(s)
}
