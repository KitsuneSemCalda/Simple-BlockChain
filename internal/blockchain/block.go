package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"
)

type Block struct {
	Index     int
	Timestamp time.Time
	BPM       int
	Hash      string
	PrevHash  string
}

func (b *Block) CalculateHash() string {
	data := strconv.Itoa(b.Index) + b.Timestamp.Format(time.RFC3339) + strconv.Itoa(b.BPM) + b.PrevHash
	hash := sha256.Sum256([]byte(data))

	return hex.EncodeToString(hash[:])
}

func NewBlock(index int, BPM int, prevHash string) *Block {
	block := &Block{
		Index:     index,
		Timestamp: time.Now(),
		BPM:       BPM,
		PrevHash:  prevHash,
	}

	block.Hash = block.CalculateHash()
	return block
}

func GenerateGenesisBlock() *Block {
	block := &Block{
		Index:     0,
		Timestamp: time.Unix(0, 0).UTC(),
		BPM:       0,
		PrevHash:  "0",
	}
	block.Hash = block.CalculateHash()
	return block
}

func (b *Block) IsValid() bool {
	return b.Hash == b.CalculateHash()
}
