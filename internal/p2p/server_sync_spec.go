package p2p

import (
	"context"
	"time"

	"KitsuneSemCalda/SBC/internal/blockchain"

	"github.com/caiolandgraf/gest/gest"
)

func init() {
	s := gest.Describe("P2P Synchronization")

	s.It("should synchronize blocks between two nodes", func(t *gest.T) {
		bcA := blockchain.NewBlockchain()
		for i := 1; i <= 5; i++ {
			bcA.AddBlock(i * 10)
		}

		cfgA := &Config{
			ListenAddr: "/ip4/127.0.0.1/tcp/0",
		}
		serverA, err := NewServer(cfgA, bcA)
		t.Expect(err).ToBeNil()

		ctxA, cancelA := context.WithCancel(context.Background())
		defer cancelA()
		go serverA.Start(ctxA)

		bcB := blockchain.NewBlockchain()
		cfgB := &Config{
			ListenAddr: "/ip4/127.0.0.1/tcp/0",
		}
		serverB, err := NewServer(cfgB, bcB)
		t.Expect(err).ToBeNil()

		ctxB, cancelB := context.WithCancel(context.Background())
		defer cancelB()
		go serverB.Start(ctxB)

		time.Sleep(100 * time.Millisecond) // Give servers time to start

		addrA := serverA.GetAddrs()[0].String() + "/p2p/" + serverA.GetHostID()
		err = serverB.ConnectToPeer(addrA)
		t.Expect(err).ToBeNil()

		// Wait for synchronization
		time.Sleep(2 * time.Second)

		t.Expect(bcB.Length()).ToBe(6) // Genesis + 5 blocks
	})

	gest.Register(s)
}
