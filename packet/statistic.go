package packet

import (
	"sync/atomic"
)

type PacketStatistic struct {
	packetCount atomic.Uint32
	bandwidth   atomic.Uint64
}

func NewPacketStatistic() *PacketStatistic {
	return &PacketStatistic{}
}

func (ps *PacketStatistic) CountPacket(size uint32) {
	ps.packetCount.Add(1)
	ps.bandwidth.Add(uint64(size))
}

func (ps *PacketStatistic) GetAndReset() (count uint32, bandwidth uint64) {
	count = ps.packetCount.Swap(0)
	bandwidth = ps.bandwidth.Swap(0)
	return
}
