package relay

import (
	"io"
	"log"
	"net"
)

func (relay *Relay) handleTCPForward() {
	for {
		conn, err := relay.tcpListener.AcceptTCP()
		if err != nil {
			log.Println("accept tcp error:", err)
			continue
		}
		relay.handleTCPConnection(conn)
	}
}

func (relay *Relay) handleTCPConnection(conn *net.TCPConn) {
	if relay.forwardTCP {
		go io.Copy(conn, relay.tcpDialer)
		go io.Copy(relay.tcpDialer, conn)
	}
	if relay.forwardUDP {
		go io.Copy(conn, relay.udpDialer)
		go io.Copy(relay.udpDialer, conn)
	}
}
