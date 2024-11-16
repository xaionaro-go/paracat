package server

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/chenx-dust/paracat/packet"
)

type udpConnContext struct {
	ctx      context.Context
	cancel   context.CancelFunc
	addr     *net.UDPAddr
	sendChan <-chan []byte
	timer    *time.Timer
}

func (server *Server) newUDPConnContext(addr *net.UDPAddr) *udpConnContext {
	ctx, cancel := context.WithCancel(context.Background())
	sendChan := server.dispatcher.NewOutChan()
	newCtx := &udpConnContext{ctx, cancel, addr, sendChan, time.NewTimer(server.cfg.UDPTimeout)}
	go server.handleUDPConnContextCancel(newCtx)
	return newCtx
}

func (server *Server) handleUDPConnContextCancel(ctx *udpConnContext) {
	<-ctx.ctx.Done()
	ctx.timer.Stop()
	server.dispatcher.CloseOutChan(ctx.sendChan)
	server.sourceMutex.Lock()
	delete(server.sourceUDPAddrs, ctx.addr.String())
	server.sourceMutex.Unlock()
}

func (server *Server) handleUDP() {
	for {
		buf := make([]byte, server.cfg.BufferSize)
		n, udpAddr, err := server.udpListener.ReadFromUDP(buf)
		if err != nil {
			log.Fatalln("error reading packet:", err)
		}

		newPacket, err := packet.Unpack(buf[:n])
		if err != nil {
			log.Println("error unpacking packet:", err)
			continue
		}

		server.handleUDPAddr(udpAddr)

		select {
		case server.filterChan.InChan() <- newPacket:
		default:
			log.Println("filter channel is full, drop packet")
		}
	}
}

func (server *Server) handleUDPAddr(addr *net.UDPAddr) {
	server.sourceMutex.RLock()
	ctx, ok := server.sourceUDPAddrs[addr.String()]
	server.sourceMutex.RUnlock()
	if ok {
		ctx.timer.Reset(server.cfg.UDPTimeout)
	} else {
		newCtx := server.newUDPConnContext(addr)
		server.sourceMutex.Lock()
		server.sourceUDPAddrs[addr.String()] = newCtx
		server.sourceMutex.Unlock()
		go server.handleUDPConnSend(newCtx)
		go server.handleUDPConnTimeout(newCtx)
	}
}

func (server *Server) handleUDPConnSend(ctx *udpConnContext) {
	defer ctx.cancel()
	for {
		select {
		case packet := <-ctx.sendChan:
			n, err := server.udpListener.WriteToUDP(packet, ctx.addr)
			if err != nil {
				log.Println("error writing packet:", err)
				log.Println("stop sending to:", ctx.addr.String())
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

func (server *Server) handleUDPConnTimeout(ctx *udpConnContext) {
	select {
	case <-ctx.timer.C:
		ctx.cancel()
	case <-ctx.ctx.Done():
		return
	}
}
