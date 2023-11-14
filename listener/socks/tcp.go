package socks

import (
	"io"
	"net"

	"github.com/MerlinKodo/clash-rev/adapter/inbound"
	N "github.com/MerlinKodo/clash-rev/common/net"
	C "github.com/MerlinKodo/clash-rev/constant"
	authStore "github.com/MerlinKodo/clash-rev/listener/auth"
	"github.com/MerlinKodo/clash-rev/transport/socks4"
	"github.com/MerlinKodo/clash-rev/transport/socks5"
)

type Listener struct {
	listener net.Listener
	addr     string
	closed   bool
}

// RawAddress implements C.Listener
func (l *Listener) RawAddress() string {
	return l.addr
}

// Address implements C.Listener
func (l *Listener) Address() string {
	return l.listener.Addr().String()
}

// Close implements C.Listener
func (l *Listener) Close() error {
	l.closed = true
	return l.listener.Close()
}

func New(addr string, tunnel C.Tunnel, additions ...inbound.Addition) (*Listener, error) {
	if len(additions) == 0 {
		additions = []inbound.Addition{
			inbound.WithInName("DEFAULT-SOCKS"),
			inbound.WithSpecialRules(""),
		}
	}
	l, err := inbound.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	sl := &Listener{
		listener: l,
		addr:     addr,
	}
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				if sl.closed {
					break
				}
				continue
			}
			go handleSocks(c, tunnel, additions...)
		}
	}()

	return sl, nil
}

func handleSocks(conn net.Conn, tunnel C.Tunnel, additions ...inbound.Addition) {
	N.TCPKeepAlive(conn)
	bufConn := N.NewBufferedConn(conn)
	head, err := bufConn.Peek(1)
	if err != nil {
		conn.Close()
		return
	}

	switch head[0] {
	case socks4.Version:
		HandleSocks4(bufConn, tunnel, additions...)
	case socks5.Version:
		HandleSocks5(bufConn, tunnel, additions...)
	default:
		conn.Close()
	}
}

func HandleSocks4(conn net.Conn, tunnel C.Tunnel, additions ...inbound.Addition) {
	authenticator := authStore.Authenticator()
	if inbound.IsRemoteAddrAllowed(conn.RemoteAddr()) {
		conn.Close()
		return
	}
	if inbound.SkipAuthRemoteAddr(conn.RemoteAddr()) {
		authenticator = nil
	}
	addr, _, err := socks4.ServerHandshake(conn, authenticator)
	if err != nil {
		conn.Close()
		return
	}
	tunnel.HandleTCPConn(inbound.NewSocket(socks5.ParseAddr(addr), conn, C.SOCKS4, additions...))
}

func HandleSocks5(conn net.Conn, tunnel C.Tunnel, additions ...inbound.Addition) {
	authenticator := authStore.Authenticator()
	if !inbound.IsRemoteAddrAllowed(conn.RemoteAddr()) {
		conn.Close()
		return
	}
	if inbound.SkipAuthRemoteAddr(conn.RemoteAddr()) {
		authenticator = nil
	}
	target, command, err := socks5.ServerHandshake(conn, authenticator)
	if err != nil {
		conn.Close()
		return
	}
	if command == socks5.CmdUDPAssociate {
		defer conn.Close()
		io.Copy(io.Discard, conn)
		return
	}
	tunnel.HandleTCPConn(inbound.NewSocket(target, conn, C.SOCKS5, additions...))
}
