package server

import (
	"log"
	"net"

	"github.com/chenx-dust/paracat/packet"
)

func (server *Server) handleForward(newPacket *packet.Packet) (int, error) {
	server.forwardMutex.RLock()
	conn, ok := server.forwardConns[newPacket.ConnID]
	server.forwardMutex.RUnlock()
	if !ok {
		server.forwardMutex.Lock()
		conn, ok = server.forwardConns[newPacket.ConnID]
		if !ok {
			remoteAddr, err := net.ResolveUDPAddr("udp", server.cfg.RemoteAddr)
			if err != nil {
				log.Fatalln("error resolving remote addr:", err)
			}
			conn, err = net.DialUDP("udp", nil, remoteAddr)
			if err != nil {
				log.Println("error dialing relay:", err)
				return 0, err
			}
			server.forwardConns[newPacket.ConnID] = conn
			go server.handleReverse(conn, newPacket.ConnID)
		}
		server.forwardMutex.Unlock()
	}

	n, err := conn.Write(newPacket.Buffer)
	if err != nil {
		log.Println("error writing to udp:", err)
	} else if n != len(newPacket.Buffer) {
		log.Println("error writing to udp: wrote", n, "bytes instead of", len(newPacket.Buffer))
	}
	return n, nil
}

func (server *Server) handleReverse(conn *net.UDPConn, connID uint16) {
	for {
		buf := make([]byte, server.cfg.BufferSize)
		n, err := conn.Read(buf)
		if err != nil {
			log.Println("error reading from udp:", err)
			log.Println("stop handling reverse conn from:", conn.RemoteAddr().String())
			return
		}

		packetID := packet.NewPacketID(&server.idIncrement)
		newPacket := &packet.Packet{
			Buffer:   buf[:n],
			ConnID:   connID,
			PacketID: packetID,
		}
		packed := newPacket.Pack()

		server.dispatcher.Dispatch(packed)
	}
}
