package client

import (
	"log"
	"net"

	"github.com/chenx-dust/paracat/packet"
)

func (client *Client) handleUDPReverse(conn *net.UDPConn) error {
	for {
		buf := make([]byte, client.cfg.BufferSize)
		n, err := conn.Read(buf)
		if err != nil {
			log.Println("error reading from reverse conn:", err)
			log.Println("close reverse conn from:", conn.RemoteAddr())
			return err
		}
		client.packetStat.ReverseRecv.CountPacket(uint32(n))

		connID, packetID, data, err := packet.Unpack(buf[:n])
		if err != nil {
			log.Println("error unpacking packet:", err)
			continue
		}

		isDuplicate := client.packetFilter.CheckDuplicatePacketID(packetID)
		if isDuplicate {
			continue
		}

		go client.sendReverse(data, len(data), connID)
	}
}
