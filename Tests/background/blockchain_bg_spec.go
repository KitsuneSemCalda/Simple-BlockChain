package background_tests

import (
	"KitsuneSemCalda/SBC/internal/blockchain"
	"github.com/caiolandgraf/gest/gest"
	"sync"
)

func init() {
	s := gest.Describe("Blockchain Background")

	s.It("should handle rapid block additions", func(t *gest.T) {
		bc := blockchain.NewBlockchain()
		
		var wg sync.WaitGroup
		numGoroutines := 10
		blocksPerGoroutine := 10
		
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < blocksPerGoroutine; j++ {
					bc.AddBlock(100 + id*10 + j)
				}
			}(i)
		}
		
		wg.Wait()
		
		expectedLength := 1 + (numGoroutines * blocksPerGoroutine)
		t.Expect(bc.Length()).ToBe(expectedLength)
		t.Expect(bc.IsValid()).ToBeTrue()
	})

	gest.Register(s)
}
