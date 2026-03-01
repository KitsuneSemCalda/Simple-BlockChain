package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"time"
)

const (
	MaxTimeDrift = 10 * time.Minute
	MaxBlockAge  = 24 * time.Hour
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

func (b *Block) Validate(prevBlock *Block) error {
	// Verificar hash
	if b.Hash != b.CalculateHash() {
		return errors.New("invalid hash")
	}

	// Verificar encadeamento
	if prevBlock != nil {
		if b.Index != prevBlock.Index+1 {
			return errors.New("invalid index")
		}
		if b.PrevHash != prevBlock.Hash {
			return errors.New("invalid prev hash")
		}
	} else if b.Index != 0 {
		return errors.New("missing previous block for non-genesis block")
	}

	// Verificar timestamp
	now := time.Now()
	if b.Timestamp.After(now.Add(MaxTimeDrift)) {
		return errors.New("timestamp too far in future")
	}

	// Verificar BPM
	if b.BPM < 0 {
		return errors.New("invalid BPM")
	}

	return nil
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
