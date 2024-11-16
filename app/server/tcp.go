package server

import (
	"context"
	"log"
	"net"

	"github.com/chenx-dust/paracat/packet"
)

type tcpConnContext struct {
	ctx    context.Context
	cancel context.CancelFunc
	conn   *net.TCPConn
}

func (server *Server) newTCPConnContext(conn *net.TCPConn) *tcpConnContext {
	ctx, cancel := context.WithCancel(context.Background())
	server.dispatcher.NewOutput(conn)
	newCtx := &tcpConnContext{ctx, cancel, conn}
	go server.handleTCPConnContextCancel(newCtx)
	return newCtx
}

func (server *Server) handleTCPConnContextCancel(ctx *tcpConnContext) {
	<-ctx.ctx.Done()
	ctx.conn.Close()
	server.dispatcher.RemoveOutput(ctx.conn)
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

		server.filterChan.Forward(newPacket)
	}
}

func (ctx *tcpConnContext) Write(packet []byte) (n int, err error) {
	n, err = ctx.conn.Write(packet)
	if err != nil {
		log.Println("error writing packet:", err)
		log.Println("stop handling connection to:", ctx.conn.RemoteAddr().String())
	} else if n != len(packet) {
		log.Println("error writing packet: wrote", n, "bytes instead of", len(packet))
	}
	return
}
