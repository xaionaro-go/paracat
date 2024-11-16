/* FilterChannel is a channel with duplicate packet filter. For MISO usage. */
package channel

import "github.com/chenx-dust/paracat/packet"

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

func (mux *FilterChannel) InChan() chan<- *packet.Packet {
	return mux.inChan
}

func (mux *FilterChannel) OutChan() <-chan *packet.Packet {
	return mux.outChan
}

func (mux *FilterChannel) Start() {
	for newPacket := range mux.inChan {
		go func(newPacket *packet.Packet) {
			mux.Statistic.CountPacket(uint32(len(newPacket.Buffer)))
			if mux.filter.CheckDuplicatePacketID(newPacket.PacketID) {
				return
			}
			mux.outChan <- newPacket
		}(newPacket)
	}
}
