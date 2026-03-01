package callbacks

import "KitsuneSemCalda/SBC/internal/blockchain"

type BlockCallbacks interface {
	OnBlockReceived(block *blockchain.Block)
}
