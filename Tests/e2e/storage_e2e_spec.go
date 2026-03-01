package e2e_tests

import (
	"KitsuneSemCalda/SBC/internal/blockchain"
	"KitsuneSemCalda/SBC/internal/storage"
	"github.com/caiolandgraf/gest/gest"
	"os"
)

func init() {
	s := gest.Describe("Storage E2E")

	testDir := "e2e_storage_data"

	s.It("should store and retrieve a complete chain after restart", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		bc.AddBlock(100)
		bc.AddBlock(200)
		bc.AddBlock(300)
		
		store, _ := storage.NewStore(testDir)
		store.Save(bc)
		store.Close()
		
		// Simulate restart
		newBC := blockchain.NewBlockchain()
		newStore, _ := storage.NewStore(testDir)
		newStore.Load(newBC)
		
		t.Expect(newBC.Length()).ToBe(4)
		t.Expect(newBC.GetLastBlock().BPM).ToBe(300)
		
		newStore.Close()
		os.RemoveAll(testDir)
	})

	gest.Register(s)
}
