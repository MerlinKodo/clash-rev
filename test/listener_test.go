package main

import (
	"net"
	"strconv"
	"testing"
	"time"

	C "github.com/MerlinKodo/clash-rev/constant"
	"github.com/MerlinKodo/clash-rev/listener"
	"github.com/MerlinKodo/clash-rev/tunnel"

	"github.com/stretchr/testify/require"
)

func TestClash_Listener(t *testing.T) {
	basic := `
log-level: silent
port: 7890
socks-port: 7891
redir-port: 7892
tproxy-port: 7893
mixed-port: 7894
`

	err := parseAndApply(basic)
	require.NoError(t, err)
	defer cleanup()

	time.Sleep(waitTime)

	for i := 7890; i <= 7894; i++ {
		require.True(t, TCPing(net.JoinHostPort("127.0.0.1", strconv.Itoa(i))), "tcp port %d", i)
	}
}

func TestClash_ListenerCreate(t *testing.T) {
	basic := `
log-level: silent
`
	err := parseAndApply(basic)
	require.NoError(t, err)
	defer cleanup()

	time.Sleep(waitTime)
	tcpIn := tunnel.TCPIn()
	udpIn := tunnel.UDPIn()

	ports := listener.Ports{
		Port: 7890,
	}
	listener.ReCreatePortsListeners(ports, tcpIn, udpIn)
	require.True(t, TCPing("127.0.0.1:7890"))
	require.Equal(t, ports, *listener.GetPorts())

	inbounds := []C.Inbound{
		{
			Type:        C.InboundTypeHTTP,
			BindAddress: "127.0.0.1:7891",
		},
	}
	listener.ReCreateListeners(inbounds, tcpIn, udpIn)
	require.True(t, TCPing("127.0.0.1:7890"))
	require.Equal(t, ports, *listener.GetPorts())

	require.True(t, TCPing("127.0.0.1:7891"))
	require.Equal(t, len(inbounds), len(listener.GetInbounds()))

	ports.Port = 0
	ports.SocksPort = 7892
	listener.ReCreatePortsListeners(ports, tcpIn, udpIn)
	require.False(t, TCPing("127.0.0.1:7890"))
	require.True(t, TCPing("127.0.0.1:7892"))
	require.Equal(t, ports, *listener.GetPorts())

	require.True(t, TCPing("127.0.0.1:7891"))
	require.Equal(t, len(inbounds), len(listener.GetInbounds()))
}
