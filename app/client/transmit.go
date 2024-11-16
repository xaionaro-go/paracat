package client

import (
	"errors"
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
			log.Println("new connection from:", addr.String())
		}
		packetID := packet.NewPacketID(&client.idIncrement)

		newPacket := &packet.Packet{
			Buffer:   buf[:n],
			ConnID:   connID,
			PacketID: packetID,
		}
		packed := newPacket.Pack()
		client.dispatcher.Dispatch(packed)
	}
}

func (client *Client) handleReverse(newPacket *packet.Packet) (n int, err error) {
	client.connMutex.RLock()
	udpAddr, ok := client.connIDAddrMap[newPacket.ConnID]
	client.connMutex.RUnlock()
	if !ok {
		log.Println("conn not found")
		return 0, errors.New("conn not found")
	}
	n, err = client.udpListener.WriteToUDP(newPacket.Buffer, udpAddr)
	if err != nil {
		log.Println("error writing to udp:", err)
	} else if n != len(newPacket.Buffer) {
		log.Println("error writing to udp: wrote", n, "bytes instead of", len(newPacket.Buffer))
	}
	return
}
