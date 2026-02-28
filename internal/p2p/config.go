package p2p

import "github.com/multiformats/go-multiaddr"

type Config struct {
	ListenAddr string
	BootNode   []string
}

func DefaultConfig() *Config {
	return &Config{
		ListenAddr: "/ip4/0.0.0.0/tcp/8333",
		BootNode:   []string{},
	}
}

func (c *Config) GetBootNodesAddr() ([]multiaddr.Multiaddr, error) {
	var addrs []multiaddr.Multiaddr

	for _, addr := range c.BootNode {
		ma, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}

		addrs = append(addrs, ma)
	}

	return addrs, nil
}
