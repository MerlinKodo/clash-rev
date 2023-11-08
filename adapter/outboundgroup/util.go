package outboundgroup

import (
	"fmt"
	"net"
	"strconv"
	"time"

	C "github.com/MerlinKodo/clash-rev/constant"
)

func addrToMetadata(rawAddress string) (addr *C.Metadata, err error) {
	host, port, err := net.SplitHostPort(rawAddress)
	if err != nil {
		err = fmt.Errorf("addrToMetadata failed: %w", err)
		return
	}

	ip := net.ParseIP(host)
	p, _ := strconv.ParseUint(port, 10, 16)
	if ip == nil {
		addr = &C.Metadata{
			Host:    host,
			DstIP:   nil,
			DstPort: C.Port(p),
		}
		return
	} else if ip4 := ip.To4(); ip4 != nil {
		addr = &C.Metadata{
			Host:    "",
			DstIP:   ip4,
			DstPort: C.Port(p),
		}
		return
	}

	addr = &C.Metadata{
		Host:    "",
		DstIP:   ip,
		DstPort: C.Port(p),
	}
	return
}

func tcpKeepAlive(c net.Conn) {
	if tcp, ok := c.(*net.TCPConn); ok {
		tcp.SetKeepAlive(true)
		tcp.SetKeepAlivePeriod(30 * time.Second)
	}
}
