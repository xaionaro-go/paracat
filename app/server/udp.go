package server

import (
	"log"

	"github.com/chenx-dust/paracat/packet"
)

func (server *Server) handleUDP() {
	for {
		buf := make([]byte, server.cfg.BufferSize)
		n, udpAddr, err := server.udpListener.ReadFromUDP(buf)
		if err != nil {
			log.Fatalln("error reading packet:", err)
		}

		connID, packetID, data, err := packet.Unpack(buf[:n])
		if err != nil {
			log.Println("error unpacking packet:", err)
			continue
		}

		server.sourceMutex.RLock()
		_, ok := server.sourceUDPAddrs[udpAddr.String()]
		server.sourceMutex.RUnlock()
		if !ok {
			server.sourceMutex.Lock()
			server.sourceUDPAddrs[udpAddr.String()] = struct{}{}
			server.sourceMutex.Unlock()
		}

		isDuplicate := server.packetFilter.CheckDuplicatePacketID(packetID)
		if isDuplicate {
			continue
		}

		go server.sendForward(data, connID)
	}
}
