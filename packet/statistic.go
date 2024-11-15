package packet

import (
	"log"
	"sync/atomic"
	"time"
)

type PacketStatistic struct {
	packetCount atomic.Uint32
	bandwidth   atomic.Uint64
}

type BiPacketStatistic struct {
	ForwardRecv PacketStatistic
	ForwardSend PacketStatistic
	ReverseRecv PacketStatistic
	ReverseSend PacketStatistic
}

func NewPacketStatistic() *PacketStatistic {
	return &PacketStatistic{}
}

func NewBiPacketStatistic() *BiPacketStatistic {
	return &BiPacketStatistic{}
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

func (bps *BiPacketStatistic) Print(period time.Duration) {
	fwdRecvCount, fwdRecvBandwidth := bps.ForwardRecv.GetAndReset()
	fwdSendCount, fwdSendBandwidth := bps.ForwardSend.GetAndReset()
	log.Printf("forward recv: %d packets, %d bytes in %s, %.2f bytes/s", fwdRecvCount, fwdRecvBandwidth, period, float64(fwdRecvBandwidth)/period.Seconds())
	log.Printf("forward send: %d packets, %d bytes in %s, %.2f bytes/s", fwdSendCount, fwdSendBandwidth, period, float64(fwdSendBandwidth)/period.Seconds())

	rvsRecvCount, rvsRecvBandwidth := bps.ReverseRecv.GetAndReset()
	rvsSendCount, rvsSendBandwidth := bps.ReverseSend.GetAndReset()
	log.Printf("reverse recv: %d packets, %d bytes in %s, %.2f bytes/s", rvsRecvCount, rvsRecvBandwidth, period, float64(rvsRecvBandwidth)/period.Seconds())
	log.Printf("reverse send: %d packets, %d bytes in %s, %.2f bytes/s", rvsSendCount, rvsSendBandwidth, period, float64(rvsSendBandwidth)/period.Seconds())
}
