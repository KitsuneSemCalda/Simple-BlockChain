package p2p

import (
	"encoding/json"
	"os"

	"KitsuneSemCalda/SBC/internal/structures"
)

type Store struct {
	filePath string
}

func NewStore(path string) *Store {
	return &Store{filePath: path}
}

func (s *Store) Save(bc *structures.Blockchain) error {
	blocks := bc.GetAllBlocks() // preciso implementar isso
	data, err := json.Marshal(blocks)
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0o644)
}

func (s *Store) Load() ([]*structures.Block, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, err
	}
	var blocks []*structures.Block
	err = json.Unmarshal(data, &blocks)
	return blocks, err
}
