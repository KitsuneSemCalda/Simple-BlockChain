package e2e_tests

import (
	"KitsuneSemCalda/SBC/internal/blockchain"
	"github.com/caiolandgraf/gest/gest"
)

func init() {
	s := gest.Describe("Blockchain E2E")

	s.It("should replace chain with longer valid chain", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		
		// Create a separate longer chain
		newBC := blockchain.NewBlockchain()
		newBC.AddBlock(100)
		newBC.AddBlock(110)
		newBC.AddBlock(120)
		
		longerBlocks := newBC.GetAllBlocks()
		
		success, msg := bc.TryAcceptChain(longerBlocks)
		t.Expect(success).ToBeTrue()
		t.Expect(msg).ToBe("chain replaced with longer version")
		t.Expect(bc.Length()).ToBe(4)
	})

	s.It("should validate chain with multiple blocks", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		bc.AddBlock(100)
		bc.AddBlock(110)
		
		blocks := bc.GetAllBlocks()
		valid, msg := bc.ValidateChain(blocks)
		t.Expect(valid).ToBeTrue()
		t.Expect(msg).ToBe("")
	})

	gest.Register(s)
}
