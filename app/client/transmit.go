package client

import (
	"log"

	"github.com/chenx-dust/paracat/packet"
)

func (client *Client) handleForward() {
	for {
		buf := make([]byte, client.cfg.BufferSize)
		n, addr, err := client.udpListener.ReadFromUDP(buf)
		if err != nil {
			log.Fatalln("error reading from udp conn:", err)
		}

		connID, ok := client.connAddrIDMap[addr.String()]
		if !ok {
			connID = uint16(client.connIncrement.Add(1) - 1)
			client.connMutex.Lock()
			client.connIDAddrMap[connID] = addr
			client.connAddrIDMap[addr.String()] = connID
			client.connMutex.Unlock()
		}
		packetID := packet.NewPacketID(&client.idIncrement)

		newPacket := &packet.Packet{
			Buffer:   buf[:n],
			ConnID:   connID,
			PacketID: packetID,
		}
		packed := newPacket.Pack()

		select {
		case client.dispatcher.InChan() <- packed:
		default:
			log.Println("dispatcher channel is full, drop packet")
		}
	}
}

func (client *Client) handleReverse() {
	for newPacket := range client.filterChan.OutChan() {
		client.connMutex.RLock()
		udpAddr, ok := client.connIDAddrMap[newPacket.ConnID]
		client.connMutex.RUnlock()
		if !ok {
			log.Println("conn not found")
			return
		}
		n, err := client.udpListener.WriteToUDP(newPacket.Buffer, udpAddr)
		if err != nil {
			log.Println("error writing to udp:", err)
		}
		if n != len(newPacket.Buffer) {
			log.Println("error writing to udp: wrote", n, "bytes instead of", len(newPacket.Buffer))
		}
	}
}
