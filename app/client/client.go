package client

import (
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chenx-dust/paracat/config"
	"github.com/chenx-dust/paracat/packet"
)

type Client struct {
	cfg         *config.Config
	udpListener *net.UDPConn
	tcpRelays   []*net.TCPConn
	udpRelays   []*net.UDPConn

	connMutex     sync.RWMutex
	connIncrement atomic.Uint32
	connIDAddrMap map[uint16]*net.UDPAddr
	connAddrIDMap map[string]uint16

	packetFilter *packet.PacketFilter
	packetStat   *packet.BiPacketStatistic
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		cfg:           cfg,
		connIDAddrMap: make(map[uint16]*net.UDPAddr),
		connAddrIDMap: make(map[string]uint16),
		packetFilter:  packet.NewPacketManager(),
		packetStat:    packet.NewBiPacketStatistic(),
	}
}

func (client *Client) Run() error {
	log.Println("running client")

	udpAddr, err := net.ResolveUDPAddr("udp", client.cfg.ListenAddr)
	if err != nil {
		return err
	}
	client.udpListener, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	log.Println("listening on", client.cfg.ListenAddr)

	client.initRelays()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		client.handleForward()
	}()
	for _, relay := range client.tcpRelays {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client.handleTCPReverse(relay)
		}()
	}
	for _, relay := range client.udpRelays {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client.handleUDPReverse(relay)
		}()
	}
	if client.cfg.ReportInterval > 0 {
		go func() {
			ticker := time.NewTicker(client.cfg.ReportInterval)
			defer ticker.Stop()
			for range ticker.C {
				client.packetStat.Print(client.cfg.ReportInterval)
			}
		}()
	}
	wg.Wait()

	return nil
}

func (client *Client) initRelays() {
	for _, relay := range client.cfg.RelayServers {
		for i := 0; i < relay.Weight; i++ {
			if relay.ConnType == config.NotDefinedConnectionType {
				log.Fatalln("not defined connection type")
			}
			enableTCP := relay.ConnType&config.TCPConnectionType != 0
			enableUDP := relay.ConnType&config.UDPConnectionType != 0
			if enableTCP {
				tcpAddr, err := net.ResolveTCPAddr("tcp", relay.Address)
				if err != nil {
					log.Println("error resolving tcp addr:", err)
					continue
				}
				conn, err := net.DialTCP("tcp", nil, tcpAddr)
				if err != nil {
					log.Println("error dialing tcp:", err)
					continue
				}
				client.tcpRelays = append(client.tcpRelays, conn)
				log.Println("connected to tcp relay", relay.Address)
			}
			if enableUDP {
				udpAddr, err := net.ResolveUDPAddr("udp", relay.Address)
				if err != nil {
					log.Println("error resolving udp addr:", err)
					continue
				}
				conn, err := net.DialUDP("udp", nil, udpAddr)
				if err != nil {
					log.Println("error dialing udp:", err)
					continue
				}
				client.udpRelays = append(client.udpRelays, conn)
				log.Println("connected to udp relay", relay.Address)
			}
		}
	}
}
