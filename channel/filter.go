/* FilterChannel is a channel with duplicate packet filter. For MISO usage. */
package channel

import (
	"github.com/chenx-dust/paracat/packet"
)

type FilterChannel struct {
	inChan  chan *packet.Packet
	outChan chan *packet.Packet
	filter  *packet.PacketFilter

	Statistic *packet.PacketStatistic
}

func NewFilterChannel(channelSize int) *FilterChannel {
	return &FilterChannel{
		inChan:    make(chan *packet.Packet, channelSize),
		outChan:   make(chan *packet.Packet, channelSize),
		filter:    packet.NewPacketFilter(),
		Statistic: packet.NewPacketStatistic(),
	}
}

func (ch *FilterChannel) InChan() chan<- *packet.Packet {
	return ch.inChan
}

func (ch *FilterChannel) OutChan() <-chan *packet.Packet {
	return ch.outChan
}

func (ch *FilterChannel) Start() {
	for newPacket := range ch.inChan {
		if ch.filter.CheckDuplicatePacketID(newPacket.PacketID) {
			continue
		}
		ch.Statistic.CountPacket(uint32(len(newPacket.Buffer)))
		ch.outChan <- newPacket
	}
}
