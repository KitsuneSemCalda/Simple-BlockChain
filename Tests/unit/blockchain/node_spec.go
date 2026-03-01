package blockchain_tests

import (
	"KitsuneSemCalda/SBC/internal/blockchain"
	"github.com/caiolandgraf/gest/gest"
	"fmt"
)

func init() {
	s := gest.Describe("BlockchainNode")

	s.It("should create node correctly", func(t *gest.T) {
		block := blockchain.NewBlock(1, 100, "prev")
		node := blockchain.NewBlockchainNode(block)
		t.Expect(node).Not().ToBeNil()
	})

	s.It("should format node string", func(t *gest.T) {
		block := blockchain.NewBlock(1, 100, "prev")
		node := blockchain.NewBlockchainNode(block)
		expected := fmt.Sprintf("Block #1 [Hash: %s...]", block.Hash[:8])
		t.Expect(node.String()).ToBe(expected)
	})

	s.It("should link nodes bidirectionally", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		bc.AddBlock(80)
		
		// Get all blocks to verify sequence
		blocks := bc.GetAllBlocks()
		t.Expect(len(blocks)).ToBe(2)
		t.Expect(blocks[0].Hash).Not().ToBe(blocks[1].Hash)
		t.Expect(blocks[1].PrevHash).ToBe(blocks[0].Hash)
	})

	gest.Register(s)
}
