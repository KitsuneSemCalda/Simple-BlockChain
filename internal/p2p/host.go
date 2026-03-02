package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"
)

const (
	ProtocolID = "/sbc/1.0.0"
)

type HostCallbacks interface {
	OnPeerFound(peer.AddrInfo)
}

type Host struct {
	host   host.Host
	ping   *ping.PingService
	config *Config

	callbacks HostCallbacks
	cbMutex   sync.RWMutex
	mdns      mdns.Service
	dht       *dht.IpfsDHT
}

// HandlePeerFound implements the mdns.Notifee interface
func (h *Host) HandlePeerFound(pi peer.AddrInfo) {
	h.cbMutex.RLock()
	cb := h.callbacks
	h.cbMutex.RUnlock()

	if cb != nil {
		cb.OnPeerFound(pi)
	}
}

type DiscoveryMsg struct {
	Type      string   `json:"type"`
	PeerID    string   `json:"peer_id"`
	Addresses []string `json:"addresses"`
}

func startUDPBroadcast(cfg *Config, h *Host, getAddrs func() []multiaddr.Multiaddr, getID func() peer.ID) {
	if !cfg.EnableUDP {
		return
	}

	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", cfg.DiscoveryPort))
	if err != nil {
		fmt.Printf("[UDP] Error: failed to resolve broadcast address: %v\n", err)
		return
	}

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: cfg.DiscoveryPort})
	if err != nil {
		fmt.Printf("[UDP] Warning: cannot listen on UDP %d: %v. Discovery only mode.\n", cfg.DiscoveryPort, err)
		return
	}

	fmt.Printf("[UDP] Discovery broadcast listening on :%d\n", cfg.DiscoveryPort)

	go func() {
		defer conn.Close()
		buf := make([]byte, 1024)
		for {
			n, remoteAddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				continue
			}

			var msg DiscoveryMsg
			if err := json.Unmarshal(buf[:n], &msg); err != nil {
				continue
			}

			if msg.Type == "announce" && msg.PeerID != "" {
				peerID, err := peer.Decode(msg.PeerID)
				if err != nil {
					continue
				}

				if peerID == getID() {
					continue
				}

				var addrs []multiaddr.Multiaddr
				for _, a := range msg.Addresses {
					ma, err := multiaddr.NewMultiaddr(a)
					if err == nil {
						addrs = append(addrs, ma)
					}
				}

				if len(addrs) > 0 {
					pi := peer.AddrInfo{
						ID:    peerID,
						Addrs: addrs,
					}
					fmt.Printf("[UDP] Discovered peer: %s from %s\n", peerID, remoteAddr.IP)
					h.HandlePeerFound(pi)
				}
			}
		}
	}()

	go func() {
		for {
			peerID := getID()
			addrs := getAddrs()
			if peerID != "" && len(addrs) > 0 {
				var addrStrings []string
				for _, a := range addrs {
					addrStrings = append(addrStrings, a.String())
				}
				msg := DiscoveryMsg{
					Type:      "announce",
					PeerID:    peerID.String(),
					Addresses: addrStrings,
				}
				data, _ := json.Marshal(msg)
				conn.WriteToUDP(data, addr)
			}
			time.Sleep(10 * time.Second)
		}
	}()
}

func (h *Host) startDHT(ctx context.Context) error {
	opts := []dht.Option{
		dht.Mode(dht.ModeAuto),
	}

	kdht, err := dht.New(ctx, h.host, opts...)
	if err != nil {
		return err
	}

	if err = kdht.Bootstrap(ctx); err != nil {
		return err
	}

	h.dht = kdht

	// Connect to bootstrap nodes
	var wg sync.WaitGroup
	bootstrapPeers := dht.DefaultBootstrapPeers
	if len(h.config.BootstrapPeers) > 0 {
		bootstrapPeers = []multiaddr.Multiaddr{}
		for _, s := range h.config.BootstrapPeers {
			ma, err := multiaddr.NewMultiaddr(s)
			if err == nil {
				bootstrapPeers = append(bootstrapPeers, ma)
			}
		}
	}

	for _, peerAddr := range bootstrapPeers {
		pi, err := peer.AddrInfoFromP2pAddr(peerAddr)
		if err != nil {
			continue
		}
		wg.Add(1)
		go func(pi peer.AddrInfo) {
			defer wg.Done()
			if err := h.host.Connect(ctx, pi); err != nil {
				// Silently fail for default bootstrap nodes to avoid log noise
			} else {
				fmt.Printf("[DHT] Connected to bootstrap node %s\n", pi.ID)
			}
		}(*pi)
	}

	routingDiscovery := routing.NewRoutingDiscovery(kdht)
	util.Advertise(ctx, routingDiscovery, h.config.Rendezvous)

	// Look for peers
	go func() {
		ticker := time.NewTicker(time.Second * 30)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				peers, err := routingDiscovery.FindPeers(ctx, h.config.Rendezvous)
				if err != nil {
					fmt.Printf("[DHT] Error finding peers: %v\n", err)
					continue
				}

				for p := range peers {
					if p.ID == h.host.ID() {
						continue
					}
					h.HandlePeerFound(p)
				}
			}
		}
	}()

	return nil
}

func NewHost(cfg *Config, cb HostCallbacks) (*Host, error) {
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(cfg.ListenAddr),
		libp2p.DisableRelay(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}
	pingService := ping.NewPingService(h)

	host := &Host{
		host:      h,
		ping:      pingService,
		config:    cfg,
		callbacks: cb,
	}

	if cfg.EnableMDNS {
		// Use 'host' itself as the notifee
		mdnsService := mdns.NewMdnsService(h, "sbc-p2p", host)
		if mdnsService == nil {
			fmt.Println("[mDNS] Warning: mDNS discovery not available")
		} else {
			if err := mdnsService.Start(); err != nil {
				fmt.Printf("[mDNS] Error starting discovery service: %v\n", err)
			} else {
				host.mdns = mdnsService
				fmt.Println("[mDNS] Discovery service started")
			}
		}
	}

	if cfg.EnableUDP {
		startUDPBroadcast(cfg, host, h.Addrs, h.ID)
	}

	if cfg.EnableDHT {
		if err := host.startDHT(context.Background()); err != nil {
			fmt.Printf("[DHT] Error starting DHT: %v\n", err)
		} else {
			fmt.Println("[DHT] Discovery service started")
		}
	}

	return host, nil
}

func (h *Host) SetPeerCallback(cb HostCallbacks) {
	h.cbMutex.Lock()
	defer h.cbMutex.Unlock()
	h.callbacks = cb
}

func (h *Host) ID() peer.ID {
	return h.host.ID()
}

func (h *Host) Addrs() []multiaddr.Multiaddr {
	return h.host.Addrs()
}

func (h *Host) Connect(ctx context.Context, addr multiaddr.Multiaddr) error {
	pi, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return err
	}
	return h.host.Connect(ctx, *pi)
}

func (h *Host) ConnectPeer(ctx context.Context, pi peer.AddrInfo) error {
	return h.host.Connect(ctx, pi)
}

func (h *Host) NewStream(ctx context.Context, p peer.ID) (network.Stream, error) {
	return h.host.NewStream(ctx, p, ProtocolID)
}

func (h *Host) SetStreamHandler(handler network.StreamHandler) {
	h.host.SetStreamHandler(ProtocolID, handler)
}

func (h *Host) Close() error {
	if h.dht != nil {
		h.dht.Close()
	}
	return h.host.Close()
}

func GetPublicIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
