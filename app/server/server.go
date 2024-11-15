package server

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/chenx-dust/paracat/config"
	"github.com/chenx-dust/paracat/packet"
)

type Server struct {
	cfg         *config.Config
	tcpListener *net.TCPListener
	udpListener *net.UDPConn

	sourceMutex    sync.RWMutex
	sourceTCPConns []*net.TCPConn
	sourceUDPAddrs map[string]struct{}

	forwardMutex sync.RWMutex
	forwardConns map[uint16]*net.UDPConn

	packetFilter *packet.PacketFilter
	packetStat   *packet.BiPacketStatistic
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		cfg:            cfg,
		forwardConns:   make(map[uint16]*net.UDPConn),
		sourceUDPAddrs: make(map[string]struct{}),
		packetFilter:   packet.NewPacketManager(),
		packetStat:     packet.NewBiPacketStatistic(),
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

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		server.handleTCP()
	}()
	go func() {
		defer wg.Done()
		server.handleUDP()
	}()
	if server.cfg.ReportInterval > 0 {
		go func() {
			ticker := time.NewTicker(server.cfg.ReportInterval)
			defer ticker.Stop()
			for range ticker.C {
				server.packetStat.Print(server.cfg.ReportInterval)
			}
		}()
	}
	wg.Wait()

	return nil
}
