/* Dispatcher is a SIMO channel. */
package channel

import (
	"errors"
	"sync"

	"github.com/chenx-dust/paracat/packet"
)

type Dispatcher struct {
	chanMutex     sync.RWMutex
	inChan        chan []byte
	outChan       []chan []byte
	roundRobinIdx int

	Statistic *packet.PacketStatistic
}

func NewDispatcher(channelSize int) *Dispatcher {
	return &Dispatcher{
		inChan:        make(chan []byte, channelSize),
		outChan:       make([]chan []byte, 0, channelSize),
		roundRobinIdx: 0,
		Statistic:     packet.NewPacketStatistic(),
	}
}

func (d *Dispatcher) InChan() chan<- []byte {
	return d.inChan
}

func (d *Dispatcher) NewOutChan() <-chan []byte {
	d.chanMutex.Lock()
	defer d.chanMutex.Unlock()
	newChan := make(chan []byte)
	d.outChan = append(d.outChan, newChan)
	return newChan
}

func (d *Dispatcher) CloseOutChan(c <-chan []byte) error {
	d.chanMutex.Lock()
	defer d.chanMutex.Unlock()
	for i := 0; i < len(d.outChan); i++ {
		if d.outChan[i] == c {
			close(d.outChan[i])
			d.outChan[i] = d.outChan[len(d.outChan)-1]
			d.outChan = d.outChan[:len(d.outChan)-1]
			return nil
		}
	}
	return errors.New("channel not found")
}

func (d *Dispatcher) StartRoundRobin() {
	for newData := range d.inChan {
		go func(newData []byte) {
			d.Statistic.CountPacket(uint32(len(newData)))
			d.chanMutex.RLock()
			nowIdx := d.roundRobinIdx % len(d.outChan)
			for i := 0; i < len(d.outChan); i++ {
				select {
				case d.outChan[nowIdx] <- newData:
					return
				default:
					nowIdx = (nowIdx + 1) % len(d.outChan)
				}
			}
			d.roundRobinIdx = (nowIdx + 1) % len(d.outChan)
			d.chanMutex.RUnlock()
		}(newData)
	}
}

func (d *Dispatcher) StartConcurrent() {
	for newData := range d.inChan {
		go func(newData []byte) {
			d.Statistic.CountPacket(uint32(len(newData)))
			d.chanMutex.RLock()
			for _, outChan := range d.outChan {
				select {
				case outChan <- newData:
				default:
				}
			}
			d.chanMutex.RUnlock()
		}(newData)
	}
}
