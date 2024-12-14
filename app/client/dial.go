package client

import (
	"log"

	"github.com/chenx-dust/paracat/config"
)

func (client *Client) dialTCPRelay(addr string) error {
	_, err := client.newTCPRelay(addr)
	if err != nil {
		log.Println("error dialing tcp relay:", err)
		return err
	}
	log.Println("connected to tcp relay", addr)
	return nil
}

func (client *Client) dialUDPRelay(addr string) error {
	_, err := client.newUDPRelay(addr)
	if err != nil {
		log.Println("error dialing udp relay:", err)
		return err
	}
	log.Println("connected to udp relay", addr)
	return nil
}

func (client *Client) dialRelays() {
	for _, relay := range client.cfg.RelayServers {
		for i := 0; i < relay.Weight; i++ {
			if relay.ConnType == config.NotDefinedConnectionType {
				log.Fatalln("not defined connection type")
			}
			enableTCP := relay.ConnType&config.TCPConnectionType != 0
			enableUDP := relay.ConnType&config.UDPConnectionType != 0
			if enableTCP {
				client.dialTCPRelay(relay.Address)
			}
			if enableUDP {
				client.dialUDPRelay(relay.Address)
			}
		}
	}
}
