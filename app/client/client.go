package client

import (
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chenx-dust/paracat/channel"
	"github.com/chenx-dust/paracat/config"
)

type Client struct {
	cfg *config.Config

	filterChan  *channel.FilterChannel
	dispatcher  *channel.Dispatcher
	idIncrement atomic.Uint32

	udpListener *net.UDPConn

	connMutex     sync.RWMutex
	connIncrement atomic.Uint32
	connIDAddrMap map[uint16]*net.UDPAddr
	connAddrIDMap map[string]uint16
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		cfg:           cfg,
		filterChan:    channel.NewFilterChannel(cfg.ChannelSize),
		dispatcher:    channel.NewDispatcher(cfg.ChannelSize),
		connIDAddrMap: make(map[uint16]*net.UDPAddr),
		connAddrIDMap: make(map[string]uint16),
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

	client.dialRelays()

	go client.filterChan.Start()
	switch client.cfg.DispatchType {
	case config.RoundRobinDispatchType:
		go client.dispatcher.StartRoundRobin()
	case config.ConcurrentDispatchType:
		go client.dispatcher.StartConcurrent()
	default:
		log.Println("unknown dispatch type, using concurrent")
		go client.dispatcher.StartConcurrent()
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		client.handleForward()
	}()
	go func() {
		defer wg.Done()
		client.handleReverse()
	}()

	if client.cfg.ReportInterval > 0 {
		go func() {
			ticker := time.NewTicker(client.cfg.ReportInterval)
			defer ticker.Stop()
			for range ticker.C {
				pkg, band := client.dispatcher.Statistic.GetAndReset()
				log.Printf("dispatcher recv: %d packets, %d bytes in %s, %.2f bytes/s", pkg, band, client.cfg.ReportInterval, float64(band)/client.cfg.ReportInterval.Seconds())
				pkg, band = client.filterChan.Statistic.GetAndReset()
				log.Printf("filter recv: %d packets, %d bytes in %s, %.2f bytes/s", pkg, band, client.cfg.ReportInterval, float64(band)/client.cfg.ReportInterval.Seconds())
			}
		}()
	}

	wg.Wait()

	return nil
}
