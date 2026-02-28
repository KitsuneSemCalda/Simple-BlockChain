package main

import (
	"os"

	_ "KitsuneSemCalda/SBC/internal/blockchain"
	_ "KitsuneSemCalda/SBC/internal/p2p"
	_ "KitsuneSemCalda/SBC/internal/storage"
	"github.com/caiolandgraf/gest/gest"
)

func main() {
	if !gest.RunRegistered() {
		os.Exit(1)
	}
}
