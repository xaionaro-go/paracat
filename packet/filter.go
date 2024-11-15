package packet

import (
	"sync"
	"sync/atomic"
)

type PacketFilter struct {
	packetIncrement atomic.Uint32
	packetMutex     sync.Mutex
	packetLowMap    map[uint16]struct{}
	packetLowClear  bool
	packetHighMap   map[uint16]struct{}
	packetHighClear bool
}

func NewPacketManager() *PacketFilter {
	return &PacketFilter{
		packetLowMap:    make(map[uint16]struct{}),
		packetHighMap:   make(map[uint16]struct{}),
		packetLowClear:  true,
		packetHighClear: true,
	}
}

func (pm *PacketFilter) NewPacketID() uint16 {
	return uint16(pm.packetIncrement.Add(1) - 1)
}

func (pm *PacketFilter) CheckDuplicatePacketID(id uint16) bool {
	/*
		divide packet id into four partitions:
		0x0000 ~ 0x3FFF: low map, allowing high map packet input
		0x4000 ~ 0x7FFF: low map, clearing high map
		0x8000 ~ 0xBFFF: high map, allowing low map packet input
		0xC000 ~ 0xFFFF: high map, clearing low map
	*/
	var ok bool
	pm.packetMutex.Lock()
	defer pm.packetMutex.Unlock()
	if id < 0x8000 {
		_, ok = pm.packetLowMap[id]
		if !ok {
			pm.packetLowMap[id] = struct{}{}
		}
		pm.packetLowClear = false
		if id > 0x3FFF && !pm.packetHighClear {
			pm.packetHighMap = make(map[uint16]struct{})
			pm.packetHighClear = true
		}
	} else {
		_, ok = pm.packetHighMap[id]
		if !ok {
			pm.packetHighMap[id] = struct{}{}
		}
		pm.packetHighClear = false
		if id < 0xC000 && !pm.packetLowClear {
			pm.packetLowMap = make(map[uint16]struct{})
			pm.packetLowClear = true
		}
	}
	return ok
}
