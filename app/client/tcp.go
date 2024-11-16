package client

import (
	"log"
	"net"

	"github.com/chenx-dust/paracat/packet"
)

func (client *Client) handleTCPReverse(conn *net.TCPConn) error {
	for {
		buf := make([]byte, client.cfg.BufferSize)
		n, connID, packetID, err := packet.ReadPacket(conn, buf)
		if err != nil {
			log.Println("error reading from reverse conn:", err)
			log.Println("close reverse conn from:", conn.RemoteAddr())
			return err
		}
		client.packetStat.ReverseRecv.CountPacket(uint32(n))

		isDuplicate := client.packetFilter.CheckDuplicatePacketID(packetID)
		if isDuplicate {
			continue
		}

		go client.sendReverse(buf[:n], n, connID)
	}
}
