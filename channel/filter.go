/* FilterChannel is a channel with duplicate packet filter. For MISO usage. */
package channel

import (
	"github.com/chenx-dust/paracat/packet"
)

type FilterChannel struct {
	outCallback func(packet *packet.Packet) (int, error)
	filter      *packet.PacketFilter

	StatisticIn  *packet.PacketStatistic
	StatisticOut *packet.PacketStatistic
}

func NewFilterChannel() *FilterChannel {
	return &FilterChannel{
		filter:       packet.NewPacketFilter(),
		StatisticIn:  packet.NewPacketStatistic(),
		StatisticOut: packet.NewPacketStatistic(),
	}
}

func (ch *FilterChannel) SetOutCallback(outCallback func(packet *packet.Packet) (int, error)) {
	ch.outCallback = outCallback
}

func (ch *FilterChannel) Forward(newPacket *packet.Packet) {
	ch.StatisticIn.CountPacket(uint32(len(newPacket.Buffer)))
	if ch.filter.CheckDuplicatePacketID(newPacket.PacketID) {
		return
	}
	n, err := ch.outCallback(newPacket)
	if err != nil {
		ch.StatisticOut.CountPacket(uint32(n))
	}
}
