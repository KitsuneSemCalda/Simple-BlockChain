package sbc_tests

import (
	"KitsuneSemCalda/SBC/internal/blockchain"
	"github.com/caiolandgraf/gest/gest"
)

func init() {
	s := gest.Describe("SBC CLI Logic")

	s.It("should validate blockchain via cli logic", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		bc.AddBlock(100)
		t.Expect(bc.IsValid()).ToBeTrue()
	})

	s.It("should report correct blockchain length", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		bc.AddBlock(100)
		bc.AddBlock(200)
		t.Expect(bc.Length()).ToBe(3)
	})

	gest.Register(s)
}
