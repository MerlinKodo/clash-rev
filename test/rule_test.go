package main

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClash_RuleInbound(t *testing.T) {
	basic := `
socks-port: 7890
inbounds:
  - socks://127.0.0.1:7891
  - type: socks
    bind-address: 127.0.0.1:7892
rules:
  - INBOUND-PORT,7891,REJECT
log-level: silent
`

	err := parseAndApply(basic)
	require.NoError(t, err)
	defer cleanup()

	require.True(t, TCPing(net.JoinHostPort("127.0.0.1", "7890")))
	require.True(t, TCPing(net.JoinHostPort("127.0.0.1", "7891")))
	require.True(t, TCPing(net.JoinHostPort("127.0.0.1", "7892")))

	require.Error(t, testPingPongWithSocksPort(t, 7891))
	require.NoError(t, testPingPongWithSocksPort(t, 7890))
	require.NoError(t, testPingPongWithSocksPort(t, 7892))
}
