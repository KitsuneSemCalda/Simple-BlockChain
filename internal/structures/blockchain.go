package structures

import "fmt"

type BlockCallback func(*Block)

type Blockchain struct {
	head      *BlockchainNode
	tail      *BlockchainNode
	length    int
	callbacks []BlockCallback
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
	newBlock := NewBlock(bc.tail.m_block.Index+1, bpm, bc.tail.m_block.Hash)
	bc.ProcessBlock(newBlock)
}

func (bc *Blockchain) ProcessBlock(block *Block) {
	// Simple validation for now: only add if it's the next index
	if block.Index != bc.tail.m_block.Index+1 {
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
	bc.callbacks = append(bc.callbacks, cb)
}

func (bc *Blockchain) GetLastBlock() *Block {
	return bc.tail.m_block
}

func (bc *Blockchain) IsValid() bool {
	current := bc.head

	for current != nil {
		if !current.m_block.IsValid() {
			return false
		}

		if current.next_block != nil && current.next_block.m_block.PrevHash != current.m_block.Hash {
			return false
		}

		current = current.next_block
	}

	return true
}

func (bc *Blockchain) Print() {
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
	return bc.length
}

func (bc *Blockchain) GetAllBlocks() []*Block {
	var blocks []*Block
	current := bc.head
	for current != nil {
		blocks = append(blocks, current.m_block)
		current = current.next_block
	}
	return blocks
}
