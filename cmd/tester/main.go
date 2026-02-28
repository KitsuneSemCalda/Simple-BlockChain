package main

import (
	"os"

	_ "KitsuneSemCalda/SBC/internal/structures"
	_ "KitsuneSemCalda/SBC/internal/p2p"
	"github.com/caiolandgraf/gest/gest"
)

func main() {
	if !gest.RunRegistered() {
		os.Exit(1)
	}
}
