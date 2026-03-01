package storage_tests

import (
	"KitsuneSemCalda/SBC/internal/blockchain"
	"KitsuneSemCalda/SBC/internal/storage"
	"github.com/caiolandgraf/gest/gest"
	"os"
)

func init() {
	s := gest.Describe("Storage Store")

	testDir := "test_sbc_data"

	s.It("should create store with valid directory", func(t *gest.T) {
		store, err := storage.NewStore(testDir)
		t.Expect(err).ToBeNil()
		t.Expect(store).Not().ToBeNil()
		store.Close()
		os.RemoveAll(testDir)
	})

	s.It("should save and load blockchain blocks", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		bc.AddBlock(100)
		bc.AddBlock(200)
		
		store, _ := storage.NewStore(testDir)
		err := store.Save(bc)
		t.Expect(err).ToBeNil()
		
		newBC := blockchain.NewBlockchain()
		err = store.Load(newBC)
		t.Expect(err).ToBeNil()
		t.Expect(newBC.Length()).ToBe(3)
		t.Expect(newBC.GetLastBlock().BPM).ToBe(200)
		
		store.Close()
		os.RemoveAll(testDir)
	})

	gest.Register(s)
}
