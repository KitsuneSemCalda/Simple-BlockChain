package callbacks

import "KitsuneSemCalda/SBC/internal/structures"

type BlockCallbacks interface {
	OnBlockReceived(block *structures.Block)
}
