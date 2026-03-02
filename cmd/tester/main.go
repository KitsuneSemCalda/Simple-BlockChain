package main

import (
	"flag"
	"os"
	"strings"

	_ "KitsuneSemCalda/SBC/Tests/background"
	_ "KitsuneSemCalda/SBC/Tests/e2e"
	_ "KitsuneSemCalda/SBC/Tests/unit/blockchain"
	_ "KitsuneSemCalda/SBC/Tests/unit/p2p"
	_ "KitsuneSemCalda/SBC/Tests/unit/sbc"
	_ "KitsuneSemCalda/SBC/Tests/unit/sbcd"
	_ "KitsuneSemCalda/SBC/Tests/unit/seed"
	_ "KitsuneSemCalda/SBC/Tests/unit/storage"
	"github.com/caiolandgraf/gest/gest"
)

func main() {
	all := flag.Bool("all", true, "Run all tests")
	unit := flag.Bool("unit", false, "Run unit tests")
	e2e := flag.Bool("e2e", false, "Run e2e tests")
	background := flag.Bool("background", false, "Run background tests")
	
	flag.Parse()

	// If any specific flag is set, don't run all by default unless --all is explicit
	anySpecific := *unit || *e2e || *background
	if anySpecific && !isFlagPassed("all") {
		*all = false
	}

	filter := ""
	if !*all {
		var filters []string
		if *unit {
			filters = append(filters, "Block", "Blockchain", "BlockchainNode", "P2P Config", "P2P Message", "P2P Peer", "P2P Logger", "P2P Server Core", "P2P Host Core", "Storage Store", "SBC CLI Logic", "SBCD Daemon Logic", "Seed Server Logic")
		}
		if *e2e {
			filters = append(filters, "Blockchain E2E", "P2P E2E", "Storage E2E")
		}
		if *background {
			filters = append(filters, "Blockchain Background", "P2P Background", "Storage Background")
		}
		filter = strings.Join(filters, "|")
	}

	if filter != "" {
		// gest.RunRegistered handles filtering via regex if supported
	}

	if !gest.RunRegistered() {
		os.Exit(1)
	}
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
