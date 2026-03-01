package sbcd_tests

import (
	"KitsuneSemCalda/SBC/internal/blockchain"
	"github.com/caiolandgraf/gest/gest"
)

func init() {
	s := gest.Describe("SBCD Daemon Logic")

	s.It("should maintain blockchain integrity in maintenance tasks", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		bc.AddBlock(100)
		t.Expect(bc.IsValid()).ToBeTrue()
	})

	s.It("should have correct initial state", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		t.Expect(bc.Length()).ToBe(1)
		t.Expect(bc.GetLastBlock().Index).ToBe(0)
	})

	gest.Register(s)
}
