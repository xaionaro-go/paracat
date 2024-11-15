package relay

import (
	"io"
	"sync"
)

func (relay *Relay) handleUDPForward() {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		io.Copy(relay.udpListener, relay.udpDialer)
	}()
	go func() {
		defer wg.Done()
		io.Copy(relay.udpDialer, relay.udpListener)
	}()
	wg.Wait()
}
