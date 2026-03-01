package blockchain_tests

import (
	"KitsuneSemCalda/SBC/internal/blockchain"
	"time"

	"github.com/caiolandgraf/gest/gest"
)

func init() {
	s := gest.Describe("Block")

	s.It("should calculate hash correctly", func(t *gest.T) {
		b := blockchain.NewBlock(1, 100, "prev-hash")
		expectedHash := b.CalculateHash()
		t.Expect(b.Hash).ToBe(expectedHash)
	})

	s.It("should validate valid block", func(t *gest.T) {
		prev := blockchain.GenerateGenesisBlock()
		b := blockchain.NewBlock(1, 100, prev.Hash)
		err := b.Validate(prev)
		t.Expect(err).ToBeNil()
	})

	s.It("should reject block with invalid hash", func(t *gest.T) {
		prev := blockchain.GenerateGenesisBlock()
		b := blockchain.NewBlock(1, 100, prev.Hash)
		b.Hash = "invalid-hash"
		err := b.Validate(prev)
		t.Expect(err).Not().ToBeNil()
		t.Expect(err.Error()).ToBe("invalid hash")
	})

	s.It("should reject block with wrong index", func(t *gest.T) {
		prev := blockchain.GenerateGenesisBlock()
		b := blockchain.NewBlock(2, 100, prev.Hash) // Index should be 1
		err := b.Validate(prev)
		t.Expect(err).Not().ToBeNil()
		t.Expect(err.Error()).ToBe("invalid index")
	})

	s.It("should reject block with wrong prevHash", func(t *gest.T) {
		prev := blockchain.GenerateGenesisBlock()
		b := blockchain.NewBlock(1, 100, "wrong-prev-hash")
		err := b.Validate(prev)
		t.Expect(err).Not().ToBeNil()
		t.Expect(err.Error()).ToBe("invalid prev hash")
	})

	s.It("should reject block with future timestamp", func(t *gest.T) {
		prev := blockchain.GenerateGenesisBlock()
		b := blockchain.NewBlock(1, 100, prev.Hash)
		b.Timestamp = time.Now().Add(1 * time.Hour)
		b.Hash = b.CalculateHash() // Re-calculate hash to pass hash check
		err := b.Validate(prev)
		t.Expect(err).Not().ToBeNil()
		t.Expect(err.Error()).ToBe("timestamp too far in future")
	})

	s.It("should reject block with negative BPM", func(t *gest.T) {
		prev := blockchain.GenerateGenesisBlock()
		b := blockchain.NewBlock(1, -50, prev.Hash)
		err := b.Validate(prev)
		t.Expect(err).Not().ToBeNil()
		t.Expect(err.Error()).ToBe("invalid BPM")
	})

	s.It("should create genesis block correctly", func(t *gest.T) {
		genesis := blockchain.GenerateGenesisBlock()
		t.Expect(genesis.Index).ToBe(0)
		t.Expect(genesis.BPM).ToBe(0)
		t.Expect(genesis.PrevHash).ToBe("0")
		t.Expect(genesis.Hash).ToBe(genesis.CalculateHash())
	})

	gest.Register(s)
}
