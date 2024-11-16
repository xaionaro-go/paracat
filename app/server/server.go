package server

import (
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chenx-dust/paracat/channel"
	"github.com/chenx-dust/paracat/config"
)

type Server struct {
	cfg         *config.Config
	tcpListener *net.TCPListener
	udpListener *net.UDPConn

	filterChan  *channel.FilterChannel
	dispatcher  *channel.Dispatcher
	idIncrement atomic.Uint32

	sourceMutex    sync.RWMutex
	sourceUDPAddrs map[string]*udpConnContext

	forwardMutex sync.RWMutex
	forwardConns map[uint16]*net.UDPConn
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		cfg:            cfg,
		filterChan:     channel.NewFilterChannel(cfg.ChannelSize),
		dispatcher:     channel.NewDispatcher(cfg.ChannelSize),
		forwardConns:   make(map[uint16]*net.UDPConn),
		sourceUDPAddrs: make(map[string]*udpConnContext),
	}
}

func (server *Server) Run() error {
	log.Println("running server")

	tcpAddr, err := net.ResolveTCPAddr("tcp", server.cfg.ListenAddr)
	if err != nil {
		return err
	}
	server.tcpListener, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	udpAddr, err := net.ResolveUDPAddr("udp", server.cfg.ListenAddr)
	if err != nil {
		return err
	}
	server.udpListener, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	log.Println("listening on", server.cfg.ListenAddr)
	log.Println("dialing to", server.cfg.RemoteAddr)

	go server.filterChan.Start()
	switch server.cfg.DispatchType {
	case config.RoundRobinDispatchType:
		go server.dispatcher.StartRoundRobin()
	case config.ConcurrentDispatchType:
		go server.dispatcher.StartConcurrent()
	default:
		log.Println("unknown dispatch type, using concurrent")
		go server.dispatcher.StartConcurrent()
	}

	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		server.handleTCP()
	}()
	go func() {
		defer wg.Done()
		server.handleUDP()
	}()
	go func() {
		defer wg.Done()
		server.handleForward()
	}()
	if server.cfg.ReportInterval > 0 {
		go func() {
			ticker := time.NewTicker(server.cfg.ReportInterval)
			defer ticker.Stop()
			for range ticker.C {
				pkg, band := server.dispatcher.Statistic.GetAndReset()
				log.Printf("dispatcher recv: %d packets, %d bytes in %s, %.2f bytes/s", pkg, band, server.cfg.ReportInterval, float64(band)/server.cfg.ReportInterval.Seconds())
				pkg, band = server.filterChan.Statistic.GetAndReset()
				log.Printf("filter recv: %d packets, %d bytes in %s, %.2f bytes/s", pkg, band, server.cfg.ReportInterval, float64(band)/server.cfg.ReportInterval.Seconds())
			}
		}()
	}
	wg.Wait()

	return nil
}
