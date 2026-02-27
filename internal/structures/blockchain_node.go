package structures

import "fmt"

type BlockchainNode struct {
	m_block    *Block
	next_block *BlockchainNode
	prev_block *BlockchainNode
}

func NewBlockchainNode(block *Block) *BlockchainNode {
	return &BlockchainNode{
		block,
		nil,
		nil,
	}
}

func (bn *BlockchainNode) String() string {
	return fmt.Sprintf("Block #%d [Hash: %s...]",
		bn.m_block.Index,
		bn.m_block.Hash[:min(8, len(bn.m_block.Hash))])
}
