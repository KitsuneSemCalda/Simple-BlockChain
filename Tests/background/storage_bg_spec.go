package background_tests

import (
	"KitsuneSemCalda/SBC/internal/blockchain"
	"KitsuneSemCalda/SBC/internal/storage"
	"github.com/caiolandgraf/gest/gest"
	"os"
	"sync"
)

func init() {
	s := gest.Describe("Storage Background")

	testDir := "bg_storage_data"

	s.It("should handle rapid store operations concurrently", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		store, _ := storage.NewStore(testDir)
		defer func() {
			store.Close()
			os.RemoveAll(testDir)
		}()
		
		var wg sync.WaitGroup
		numOperations := 10
		
		for i := 0; i < numOperations; i++ {
			wg.Add(1)
			go func(bpm int) {
				defer wg.Done()
				bc.AddBlock(bpm)
				store.Save(bc)
			}(100 + i)
		}
		
		wg.Wait()
		
		newBC := blockchain.NewBlockchain()
		store.Load(newBC)
		t.Expect(newBC.Length()).Not().ToBe(0)
	})

	gest.Register(s)
}
