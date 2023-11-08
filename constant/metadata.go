package constant

import (
	"encoding/json"
	"net"
	"net/netip"
	"strconv"

	"github.com/MerlinKodo/clash-rev/transport/socks5"
)

// Socks addr type
const (
	TCP NetWork = iota
	UDP

	HTTP Type = iota
	HTTPCONNECT
	SOCKS4
	SOCKS5
	REDIR
	TPROXY
	TUNNEL
)

type NetWork int

func (n NetWork) String() string {
	if n == TCP {
		return "tcp"
	}
	return "udp"
}

func (n NetWork) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.String())
}

type Type int

func (t Type) String() string {
	switch t {
	case HTTP:
		return "HTTP"
	case HTTPCONNECT:
		return "HTTP Connect"
	case SOCKS4:
		return "Socks4"
	case SOCKS5:
		return "Socks5"
	case REDIR:
		return "Redir"
	case TPROXY:
		return "TProxy"
	case TUNNEL:
		return "Tunnel"
	default:
		return "Unknown"
	}
}

func (t Type) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// Metadata is used to store connection address
type Metadata struct {
	NetWork      NetWork `json:"network"`
	Type         Type    `json:"type"`
	SrcIP        net.IP  `json:"sourceIP"`
	DstIP        net.IP  `json:"destinationIP"`
	SrcPort      Port    `json:"sourcePort"`
	DstPort      Port    `json:"destinationPort"`
	Host         string  `json:"host"`
	DNSMode      DNSMode `json:"dnsMode"`
	ProcessPath  string  `json:"processPath"`
	SpecialProxy string  `json:"specialProxy"`

	OriginDst netip.AddrPort `json:"-"`
}

func (m *Metadata) RemoteAddress() string {
	return net.JoinHostPort(m.String(), m.DstPort.String())
}

func (m *Metadata) SourceAddress() string {
	return net.JoinHostPort(m.SrcIP.String(), m.SrcPort.String())
}

func (m *Metadata) AddrType() int {
	switch true {
	case m.Host != "" || m.DstIP == nil:
		return socks5.AtypDomainName
	case m.DstIP.To4() != nil:
		return socks5.AtypIPv4
	default:
		return socks5.AtypIPv6
	}
}

func (m *Metadata) Resolved() bool {
	return m.DstIP != nil
}

// Pure is used to solve unexpected behavior
// when dialing proxy connection in DNSMapping mode.
func (m *Metadata) Pure() *Metadata {
	if m.DNSMode == DNSMapping && m.DstIP != nil {
		copy := *m
		copy.Host = ""
		return &copy
	}

	return m
}

func (m *Metadata) UDPAddr() *net.UDPAddr {
	if m.NetWork != UDP || m.DstIP == nil {
		return nil
	}
	return &net.UDPAddr{
		IP:   m.DstIP,
		Port: int(m.DstPort),
	}
}

func (m *Metadata) String() string {
	if m.Host != "" {
		return m.Host
	} else if m.DstIP != nil {
		return m.DstIP.String()
	} else {
		return "<nil>"
	}
}

func (m *Metadata) Valid() bool {
	return m.Host != "" || m.DstIP != nil
}

// Port is used to compatible with old version
type Port uint16

func (n Port) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.String())
}

func (n Port) String() string {
	return strconv.FormatUint(uint64(n), 10)
}
