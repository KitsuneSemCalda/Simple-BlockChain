package blockchain

import (
	"fmt"
	"sync"
)

type BlockCallback func(*Block)

type Blockchain struct {
	head      *BlockchainNode
	tail      *BlockchainNode
	length    int
	callbacks []BlockCallback
	mu        sync.RWMutex
}

func NewBlockchain() *Blockchain {
	genesisBlock := GenerateGenesisBlock()
	genesisNode := NewBlockchainNode(genesisBlock)

	return &Blockchain{
		head:      genesisNode,
		tail:      genesisNode,
		length:    1,
		callbacks: []BlockCallback{},
	}
}

func (bc *Blockchain) AddBlock(bpm int) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	newBlock := NewBlock(bc.tail.m_block.Index+1, bpm, bc.tail.m_block.Hash)
	bc.processBlockInternal(newBlock)
}

func (bc *Blockchain) ProcessBlock(block *Block) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.processBlockInternal(block)
}

func (bc *Blockchain) processBlockInternal(block *Block) {
	if err := block.Validate(bc.tail.m_block); err != nil {
		return
	}

	newNode := NewBlockchainNode(block)
	newNode.prev_block = bc.tail
	bc.tail.next_block = newNode
	bc.tail = newNode
	bc.length++

	for _, cb := range bc.callbacks {
		cb(block)
	}
}

func (bc *Blockchain) Subscribe(cb BlockCallback) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.callbacks = append(bc.callbacks, cb)
}

func (bc *Blockchain) GetLastBlock() *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.tail.m_block
}

func (bc *Blockchain) IsValid() bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	current := bc.head

	for current != nil {
		var prev *Block
		if current.prev_block != nil {
			prev = current.prev_block.m_block
		}
		if err := current.m_block.Validate(prev); err != nil {
			return false
		}

		current = current.next_block
	}

	return true
}

func (bc *Blockchain) Print() {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	current := bc.head
	for current != nil {
		fmt.Printf("Index: %d, BPM: %d, Hash: %s\n",
			current.m_block.Index,
			current.m_block.BPM,
			current.m_block.Hash)
		current = current.next_block
	}
}

func (bc *Blockchain) Length() int {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.length
}

func (bc *Blockchain) GetAllBlocks() []*Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	var blocks []*Block
	current := bc.head
	for current != nil {
		blocks = append(blocks, current.m_block)
		current = current.next_block
	}
	return blocks
}

func (bc *Blockchain) GetBlockByHash(hash string) *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	current := bc.head
	for current != nil {
		if current.m_block.Hash == hash {
			return current.m_block
		}
		current = current.next_block
	}
	return nil
}

func (bc *Blockchain) GetBlocksAfter(hash string, limit int) []*Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	var blocks []*Block
	current := bc.head

	if hash != "" {
		for current != nil && current.m_block.Hash != hash {
			current = current.next_block
		}

		if current != nil {
			current = current.next_block
		} else {
			current = bc.head
		}
	}

	count := 0
	for current != nil && count < limit {
		blocks = append(blocks, current.m_block)
		current = current.next_block
		count++
	}

	return blocks
}

func (bc *Blockchain) ValidateChain(blocks []*Block) (bool, string) {
	// This doesn't need a lock because it only depends on input
	if len(blocks) == 0 {
		return false, "empty chain"
	}

	genesis := GenerateGenesisBlock()
	if blocks[0].Index == 0 {
		if blocks[0].Hash != genesis.Hash {
			return false, "genesis block mismatch"
		}
	}

	if err := blocks[0].Validate(nil); err != nil {
		return false, "first block invalid: " + err.Error()
	}

	for i := 1; i < len(blocks); i++ {
		if err := blocks[i].Validate(blocks[i-1]); err != nil {
			return false, fmt.Sprintf("block %d invalid: %v", i, err)
		}
	}

	return true, ""
}

func (bc *Blockchain) ReplaceChain(blocks []*Block) bool {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	
	if len(blocks) <= bc.length {
		return false
	}

	valid, _ := bc.ValidateChain(blocks)
	if !valid {
		return false
	}

	bc.head = nil
	bc.tail = nil
	bc.length = 0

	for _, block := range blocks {
		newNode := NewBlockchainNode(block)
		if bc.head == nil {
			bc.head = newNode
			bc.tail = newNode
		} else {
			newNode.prev_block = bc.tail
			bc.tail.next_block = newNode
			bc.tail = newNode
		}
		bc.length++
	}

	return true
}

func (bc *Blockchain) TryAcceptChain(newBlocks []*Block) (bool, string) {
	// We check length before locking for performance, then check again inside ReplaceChain
	bc.mu.RLock()
	localLen := bc.length
	bc.mu.RUnlock()

	if len(newBlocks) <= localLen {
		return false, "chain not longer"
	}

	valid, reason := bc.ValidateChain(newBlocks)
	if !valid {
		return false, reason
	}

	if bc.ReplaceChain(newBlocks) {
		return true, "chain replaced with longer version"
	}

	return false, "failed to replace chain"
}
