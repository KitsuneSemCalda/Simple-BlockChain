package p2p

import (
	"flag"
	"fmt"
	"strings"

	"github.com/multiformats/go-multiaddr"
)

type Config struct {
	ListenAddr string
	DataDir    string
	BootNode   []string
}

func DefaultConfig() *Config {
	return &Config{
		ListenAddr: "/ip4/0.0.0.0/tcp/8333",
		DataDir:    ".",
		BootNode:   []string{},
	}
}

func (c *Config) ParseFlags() {
	var bootNodes string
	flag.StringVar(&c.ListenAddr, "listen", c.ListenAddr, "Address to listen on")
	flag.StringVar(&c.DataDir, "datadir", c.DataDir, "Directory to store blockchain data")
	flag.StringVar(&bootNodes, "bootnode", "", "Comma-separated list of boot nodes")
	flag.Parse()

	if bootNodes != "" {
		c.BootNode = strings.Split(bootNodes, ",")
	}
}

func (c *Config) GetBootNodesAddr() ([]multiaddr.Multiaddr, error) {
	var addrs []multiaddr.Multiaddr

	for _, addr := range c.BootNode {
		if addr == "" {
			continue
		}
		ma, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil, fmt.Errorf("invalid multiaddr %s: %w", addr, err)
		}

		addrs = append(addrs, ma)
	}

	return addrs, nil
}
