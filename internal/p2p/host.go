package p2p

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"
)

type HostCallbacks interface {
}

type Host struct {
	host   host.Host
	ping   *ping.PingService
	config *Config

	callbacks HostCallbacks
}

const ProtocolID = "/sbc/1.0.0"

func NewHost(cfg *Config) (*Host, error) {
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(cfg.ListenAddr),
		libp2p.DisableRelay(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}
	pingService := ping.NewPingService(h)
	return &Host{
		host:   h,
		ping:   pingService,
		config: cfg,
	}, nil
}

func (h *Host) SetCallbacks(cb HostCallbacks) {
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

func (h *Host) NewStream(ctx context.Context, p peer.ID) (network.Stream, error) {
	return h.host.NewStream(ctx, p, ProtocolID)
}

func (h *Host) SetStreamHandler(handler network.StreamHandler) {
	h.host.SetStreamHandler(ProtocolID, handler)
}

func (h *Host) Close() error {
	return h.host.Close()
}
