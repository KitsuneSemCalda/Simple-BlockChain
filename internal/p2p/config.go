package p2p

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/multiformats/go-multiaddr"
)

type Config struct {
	ListenAddr     string
	DataDir        string
	BootNode       []string
	DNSSeed        string
	EnableMDNS     bool
	EnableUDP      bool
	DiscoveryPort  int
	EnableDHT      bool
	Rendezvous     string
	BootstrapPeers []string
}

func DefaultDataDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "sbc")
	}

	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".local/share/sbc")
	}

	if home := os.Getenv("USERPROFILE"); home != "" {
		return filepath.Join(home, "AppData", "Roaming", "SBC")
	}

	return "sbc_data"
}

var DefaultBootNodes = []string{}

var WellKnownBootstrapNodes = []string{}

var DefaultDNSServers = []string{}

func GetBootstrapNodes() []string {
	if env := os.Getenv("SBC_BOOTNODES"); env != "" {
		return strings.Split(env, ",")
	}
	if len(WellKnownBootstrapNodes) > 0 {
		return WellKnownBootstrapNodes
	}
	return DefaultBootNodes
}

func GetDNSServers() string {
	if env := os.Getenv("SBC_DNS"); env != "" {
		return env
	}
	if len(DefaultDNSServers) > 0 {
		return strings.Join(DefaultDNSServers, ",")
	}
	return ""
}

func DefaultConfig() *Config {
	bootNodes := GetBootstrapNodes()

	return &Config{
		ListenAddr:    "/ip4/0.0.0.0/tcp/0", // 0 = random port
		DataDir:       DefaultDataDir(),
		BootNode:      bootNodes,
		DNSSeed:       GetDNSServers(),
		EnableMDNS:    true,
		EnableUDP:     true,
		DiscoveryPort: 9999,
		EnableDHT:     true,
		Rendezvous:    "sbc-peers",
	}
}

func (c *Config) ParseFlags() {
	var bootNodes string
	var bootstrapPeers string
	flag.StringVar(&c.ListenAddr, "listen", c.ListenAddr, "Address to listen on")
	flag.StringVar(&c.DataDir, "datadir", c.DataDir, "Directory to store blockchain data")
	flag.StringVar(&bootNodes, "bootnode", "", "Comma-separated list of boot nodes")
	flag.StringVar(&c.DNSSeed, "dns", c.DNSSeed, "DNS seed servers for peer discovery")
	flag.BoolVar(&c.EnableMDNS, "enable-mdns", c.EnableMDNS, "Enable mDNS discovery")
	flag.BoolVar(&c.EnableUDP, "enable-udp", c.EnableUDP, "Enable UDP broadcast discovery")
	flag.IntVar(&c.DiscoveryPort, "discovery-port", c.DiscoveryPort, "UDP port for discovery broadcast")
	flag.BoolVar(&c.EnableDHT, "enable-dht", c.EnableDHT, "Enable Kademlia DHT discovery")
	flag.StringVar(&c.Rendezvous, "rendezvous", c.Rendezvous, "Rendezvous string for DHT discovery")
	flag.StringVar(&bootstrapPeers, "bootstrap", "", "Comma-separated list of DHT bootstrap peers")
	flag.Parse()

	if bootNodes != "" {
		c.BootNode = strings.Split(bootNodes, ",")
	}
	if bootstrapPeers != "" {
		c.BootstrapPeers = strings.Split(bootstrapPeers, ",")
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
