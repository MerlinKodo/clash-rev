package process

import (
	"net"
	"net/netip"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testConn(t *testing.T, network, address string) {
	l, err := net.Listen(network, address)
	if err != nil {
		assert.FailNow(t, "Listen failed", err)
	}
	defer l.Close()

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		assert.FailNow(t, "Dial failed", err)
	}
	defer conn.Close()

	rConn, err := l.Accept()
	if err != nil {
		assert.FailNow(t, "Accept conn failed", err)
	}
	defer rConn.Close()

	path, err := FindProcessPath(TCP, conn.LocalAddr().(*net.TCPAddr).AddrPort(), conn.RemoteAddr().(*net.TCPAddr).AddrPort())
	if err != nil {
		assert.FailNow(t, "Find process path failed", err)
	}

	exePath, err := os.Executable()
	if err != nil {
		assert.FailNow(t, "Get executable failed", err)
	}

	assert.Equal(t, exePath, path)
}

func TestFindProcessPathTCP(t *testing.T) {
	t.Run("v4", func(t *testing.T) {
		testConn(t, "tcp4", "127.0.0.1:0")
	})
	t.Run("v6", func(t *testing.T) {
		testConn(t, "tcp6", "[::1]:0")
	})
}

func testPacketConn(t *testing.T, network, lAddress, rAddress string) {
	lConn, err := net.ListenPacket(network, lAddress)
	if err != nil {
		assert.FailNow(t, "ListenPacket failed", err)
	}
	defer lConn.Close()

	rConn, err := net.ListenPacket(network, rAddress)
	if err != nil {
		assert.FailNow(t, "ListenPacket failed", err)
	}
	defer rConn.Close()

	_, err = lConn.WriteTo([]byte{0}, rConn.LocalAddr())
	if err != nil {
		assert.FailNow(t, "Send message failed", err)
	}

	_, lAddr, err := rConn.ReadFrom([]byte{0})
	if err != nil {
		assert.FailNow(t, "Receive message failed", err)
	}

	path, err := FindProcessPath(UDP, lAddr.(*net.UDPAddr).AddrPort(), rConn.LocalAddr().(*net.UDPAddr).AddrPort())
	if err != nil {
		assert.FailNow(t, "Find process path", err)
	}

	exePath, err := os.Executable()
	if err != nil {
		assert.FailNow(t, "Find executable", err)
	}

	assert.Equal(t, exePath, path)
}

func TestFindProcessPathUDP(t *testing.T) {
	t.Run("v4", func(t *testing.T) {
		testPacketConn(t, "udp4", "127.0.0.1:0", "127.0.0.1:0")
	})
	t.Run("v6", func(t *testing.T) {
		testPacketConn(t, "udp6", "[::1]:0", "[::1]:0")
	})
	t.Run("v4AnyLocal", func(t *testing.T) {
		testPacketConn(t, "udp4", "0.0.0.0:0", "127.0.0.1:0")
	})
	t.Run("v6AnyLocal", func(t *testing.T) {
		testPacketConn(t, "udp6", "[::]:0", "[::1]:0")
	})
}

func BenchmarkFindProcessName(b *testing.B) {
	from := netip.MustParseAddrPort("127.0.0.1:11447")
	to := netip.MustParseAddrPort("127.0.0.1:33669")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FindProcessPath(TCP, from, to)
	}
}
