package server

import (
	"context"
	"log"
	"net"

	"github.com/chenx-dust/paracat/packet"
)

type tcpConnContext struct {
	ctx      context.Context
	cancel   context.CancelFunc
	conn     *net.TCPConn
	sendChan <-chan []byte
}

func (server *Server) newTCPConnContext(conn *net.TCPConn) *tcpConnContext {
	ctx, cancel := context.WithCancel(context.Background())
	sendChan := server.dispatcher.NewOutChan()
	newCtx := &tcpConnContext{ctx, cancel, conn, sendChan}
	go server.handleTCPConnContextCancel(newCtx)
	return newCtx
}

func (server *Server) handleTCPConnContextCancel(ctx *tcpConnContext) {
	<-ctx.ctx.Done()
	ctx.conn.Close()
	server.dispatcher.CloseOutChan(ctx.sendChan)
}

func (server *Server) handleTCP() {
	for {
		conn, err := server.tcpListener.AcceptTCP()
		if err != nil {
			log.Fatalln("error accepting tcp connection:", err)
		}
		server.handleTCPConn(conn)
	}
}

func (server *Server) handleTCPConn(conn *net.TCPConn) {
	log.Println("new tcp connection from", conn.RemoteAddr().String())
	ctx := server.newTCPConnContext(conn)
	go server.handleTCPConnRecv(ctx)
	go server.handleTCPConnSend(ctx)
}

func (server *Server) handleTCPConnRecv(ctx *tcpConnContext) {
	defer ctx.cancel()
	for {
		select {
		case <-ctx.ctx.Done():
			return
		default:
		}
		newPacket, err := packet.ReadPacket(ctx.conn)
		if err != nil {
			log.Println("error reading packet:", err)
			log.Println("stop handling connection from:", ctx.conn.RemoteAddr().String())
			return
		}

		select {
		case server.filterChan.InChan() <- newPacket:
		default:
			log.Println("filter channel is full, drop packet")
		}
	}
}

func (server *Server) handleTCPConnSend(ctx *tcpConnContext) {
	defer ctx.cancel()
	for {
		select {
		case packet := <-ctx.sendChan:
			n, err := ctx.conn.Write(packet)
			if err != nil {
				log.Println("error writing packet:", err)
				log.Println("stop handling connection to:", ctx.conn.RemoteAddr().String())
				return
			}
			if n != len(packet) {
				log.Println("error writing packet: wrote", n, "bytes instead of", len(packet))
				continue
			}
		case <-ctx.ctx.Done():
			return
		}
	}
}
