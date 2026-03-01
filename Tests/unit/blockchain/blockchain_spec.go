package blockchain_tests

import (
	"KitsuneSemCalda/SBC/internal/blockchain"
	"github.com/caiolandgraf/gest/gest"
)

func init() {
	s := gest.Describe("Blockchain")

	s.It("should create blockchain with genesis", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		t.Expect(bc.Length()).ToBe(1)
		t.Expect(bc.GetLastBlock().Index).ToBe(0)
	})

	s.It("should add blocks correctly", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		bc.AddBlock(80)
		t.Expect(bc.Length()).ToBe(2)
		t.Expect(bc.GetLastBlock().BPM).ToBe(80)
		t.Expect(bc.GetLastBlock().Index).ToBe(1)
	})

	s.It("should process valid blocks", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		last := bc.GetLastBlock()
		newBlock := blockchain.NewBlock(1, 100, last.Hash)
		bc.ProcessBlock(newBlock)
		t.Expect(bc.Length()).ToBe(2)
		t.Expect(bc.GetLastBlock().Hash).ToBe(newBlock.Hash)
	})

	s.It("should reject invalid blocks in process", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		invalidBlock := blockchain.NewBlock(1, 100, "invalid-prev-hash")
		bc.ProcessBlock(invalidBlock)
		t.Expect(bc.Length()).ToBe(1) // Should remain 1
	})

	s.It("should subscribe and notify callbacks", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		called := false
		bc.Subscribe(func(b *blockchain.Block) {
			called = true
			t.Expect(b.BPM).ToBe(120)
		})
		bc.AddBlock(120)
		t.Expect(called).ToBe(true)
	})

	s.It("should return last block", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		bc.AddBlock(70)
		bc.AddBlock(90)
		t.Expect(bc.GetLastBlock().BPM).ToBe(90)
	})

	s.It("should validate valid chain", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		bc.AddBlock(80)
		bc.AddBlock(90)
		t.Expect(bc.IsValid()).ToBeTrue()
	})

	s.It("should return all blocks", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		bc.AddBlock(80)
		bc.AddBlock(90)
		blocks := bc.GetAllBlocks()
		t.Expect(len(blocks)).ToBe(3)
		t.Expect(blocks[0].Index).ToBe(0)
		t.Expect(blocks[1].Index).ToBe(1)
		t.Expect(blocks[2].Index).ToBe(2)
	})

	s.It("should find block by hash", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		bc.AddBlock(80)
		last := bc.GetLastBlock()
		found := bc.GetBlockByHash(last.Hash)
		t.Expect(found).Not().ToBeNil()
		t.Expect(found.Hash).ToBe(last.Hash)
	})

	s.It("should get blocks after hash", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		bc.AddBlock(80) // index 1
		bc.AddBlock(90) // index 2
		bc.AddBlock(100) // index 3
		
		genesis := bc.GetAllBlocks()[0]
		afterGenesis := bc.GetBlocksAfter(genesis.Hash, 2)
		t.Expect(len(afterGenesis)).ToBe(2)
		t.Expect(afterGenesis[0].Index).ToBe(1)
		t.Expect(afterGenesis[1].Index).ToBe(2)
	})

	gest.Register(s)
}
