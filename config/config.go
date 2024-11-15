package config

import "time"

type AppMode int
type ConnectionType int

const (
	NotDefined AppMode = iota
	ClientMode
	RelayMode // optional for udp-to-tcp or tcp-to-udp
	ServerMode
)

const (
	NotDefinedConnectionType ConnectionType = iota
	TCPConnectionType
	UDPConnectionType
	BothConnectionType
)

type Config struct {
	Mode           AppMode
	ListenAddr     string
	RemoteAddr     string        // not necessary in ClientMode
	RelayServers   []RelayServer // only used in ClientMode
	RelayType      RelayType     // only used in RelayMode
	BufferSize     int
	ReportInterval time.Duration
}

type RelayServer struct {
	Address  string
	ConnType ConnectionType
	Weight   int
}

type RelayType struct {
	ListenType  ConnectionType
	ForwardType ConnectionType
}

func ConnTypeToString(connType ConnectionType) string {
	switch connType {
	case TCPConnectionType:
		return "tcp"
	case UDPConnectionType:
		return "udp"
	case BothConnectionType:
		return "both"
	default:
		return "unknown"
	}
}
