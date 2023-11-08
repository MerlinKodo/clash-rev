package inbound

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/MerlinKodo/clash-rev/common/util"
	C "github.com/MerlinKodo/clash-rev/constant"
	"github.com/MerlinKodo/clash-rev/transport/socks5"
)

func parseSocksAddr(target socks5.Addr) *C.Metadata {
	metadata := &C.Metadata{}

	switch target[0] {
	case socks5.AtypDomainName:
		// trim for FQDN
		metadata.Host = strings.TrimRight(string(target[2:2+target[1]]), ".")
		metadata.DstPort = C.Port((int(target[2+target[1]]) << 8) | int(target[2+target[1]+1]))
	case socks5.AtypIPv4:
		ip := net.IP(target[1 : 1+net.IPv4len])
		metadata.DstIP = ip
		metadata.DstPort = C.Port((int(target[1+net.IPv4len]) << 8) | int(target[1+net.IPv4len+1]))
	case socks5.AtypIPv6:
		ip := net.IP(target[1 : 1+net.IPv6len])
		metadata.DstIP = ip
		metadata.DstPort = C.Port((int(target[1+net.IPv6len]) << 8) | int(target[1+net.IPv6len+1]))
	}

	return metadata
}

func parseHTTPAddr(request *http.Request) *C.Metadata {
	host := request.URL.Hostname()
	port, _ := strconv.ParseUint(util.EmptyOr(request.URL.Port(), "80"), 10, 16)

	// trim FQDN (#737)
	host = strings.TrimRight(host, ".")

	metadata := &C.Metadata{
		NetWork: C.TCP,
		Host:    host,
		DstIP:   nil,
		DstPort: C.Port(port),
	}

	if ip := net.ParseIP(host); ip != nil {
		metadata.DstIP = ip
	}

	return metadata
}

func parseAddr(addr net.Addr) (net.IP, int, error) {
	switch a := addr.(type) {
	case *net.TCPAddr:
		return a.IP, a.Port, nil
	case *net.UDPAddr:
		return a.IP, a.Port, nil
	default:
		return nil, 0, fmt.Errorf("unknown address type %s", addr.String())
	}
}
