package callbacks

import "github.com/libp2p/go-libp2p/core/peer"

type PeerCallbacks interface {
	OnNewPeer(id peer.ID)
	OnDisconnect(id peer.ID)
	OnPeerFound(info peer.AddrInfo)
}
