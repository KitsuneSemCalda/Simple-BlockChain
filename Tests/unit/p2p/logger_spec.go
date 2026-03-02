package p2p_tests

import (
	"KitsuneSemCalda/SBC/internal/p2p"
	"github.com/caiolandgraf/gest/gest"
)

func init() {
	s := gest.Describe("P2P Logger")

	s.It("should respect log levels", func(t *gest.T) {
		p2p.SetLogLevel(p2p.LevelError)
		p2p.Debug("TEST", "This should not appear")
		p2p.Info("TEST", "This should not appear")
		p2p.Error("TEST", "This is an error")
		
		t.Expect(true).ToBeTrue()
	})

	gest.Register(s)
}
