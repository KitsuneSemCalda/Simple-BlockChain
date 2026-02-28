package structures

import (
	"github.com/caiolandgraf/gest/gest"
)

func init() {
	s := gest.Describe("Blockchain")

	s.It("should create a blockchain with a genesis block", func(t *gest.T) {
		bc := NewBlockchain()
		t.Expect(bc.Length()).ToBe(1)
		t.Expect(bc.GetLastBlock()).Not().ToBeNil()
		t.Expect(bc.GetLastBlock().Index).ToBe(0)
	})

	s.It("should add blocks correctly", func(t *gest.T) {
		bc := NewBlockchain()
		bc.AddBlock(100)
		t.Expect(bc.Length()).ToBe(2)
		t.Expect(bc.GetLastBlock().BPM).ToBe(100)
		t.Expect(bc.GetLastBlock().Index).ToBe(1)
	})

	s.It("should validate the blockchain correctly", func(t *gest.T) {
		bc := NewBlockchain()
		bc.AddBlock(100)
		bc.AddBlock(200)
		t.Expect(bc.IsValid()).ToBeTrue()
	})

	gest.Register(s)
}
