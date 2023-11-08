package obfs

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"net/netip"

	"github.com/MerlinKodo/protobytes"
)

type SessionStatus = byte

const (
	SessionStatusNew       SessionStatus = 0x01
	SessionStatusKeep      SessionStatus = 0x02
	SessionStatusEnd       SessionStatus = 0x03
	SessionStatusKeepAlive SessionStatus = 0x04
)

const (
	OptionNone  = byte(0x00)
	OptionData  = byte(0x01)
	OptionError = byte(0x02)
)

type MuxOption struct {
	ID   [2]byte
	Port uint16
	Host string
	Type string
}

// Mux is an mux-compatible client for v2ray-plugin, not a complete implementation
type Mux struct {
	net.Conn
	buf    protobytes.BytesWriter
	id     [2]byte
	length [2]byte
	status [2]byte
	otb    []byte
	remain int
}

func (m *Mux) Read(b []byte) (int, error) {
	if m.remain != 0 {
		length := m.remain
		if len(b) < m.remain {
			length = len(b)
		}

		n, err := m.Conn.Read(b[:length])
		if err != nil {
			return 0, err
		}
		m.remain -= n
		return n, nil
	}

	for {
		_, err := io.ReadFull(m.Conn, m.length[:])
		if err != nil {
			return 0, err
		}
		length := binary.BigEndian.Uint16(m.length[:])
		if length > 512 {
			return 0, errors.New("invalid metalen")
		}

		_, err = io.ReadFull(m.Conn, m.id[:])
		if err != nil {
			return 0, err
		}

		_, err = m.Conn.Read(m.status[:])
		if err != nil {
			return 0, err
		}

		opcode := m.status[0]
		if opcode == SessionStatusKeepAlive {
			continue
		}

		opts := m.status[1]

		if opts != OptionData {
			continue
		}

		_, err = io.ReadFull(m.Conn, m.length[:])
		if err != nil {
			return 0, err
		}
		dataLen := int(binary.BigEndian.Uint16(m.length[:]))
		m.remain = dataLen
		if dataLen > len(b) {
			dataLen = len(b)
		}

		n, err := m.Conn.Read(b[:dataLen])
		m.remain -= n
		return n, err
	}
}

func (m *Mux) Write(b []byte) (int, error) {
	if m.otb != nil {
		// create a sub connection
		if _, err := m.Conn.Write(m.otb); err != nil {
			return 0, err
		}
		m.otb = nil
	}
	m.buf.Reset()
	m.buf.PutUint16be(4)
	m.buf.PutSlice(m.id[:])
	m.buf.PutUint8(SessionStatusKeep)
	m.buf.PutUint8(OptionData)
	m.buf.PutUint16be(uint16(len(b)))
	m.buf.Write(b)

	return m.Conn.Write(m.buf.Bytes())
}

func (m *Mux) Close() error {
	_, err := m.Conn.Write([]byte{0x0, 0x4, m.id[0], m.id[1], SessionStatusEnd, OptionNone})
	if err != nil {
		return err
	}
	return m.Conn.Close()
}

func NewMux(conn net.Conn, option MuxOption) *Mux {
	buf := protobytes.BytesWriter{}

	// fill empty length
	buf.PutSlice([]byte{0x0, 0x0})
	buf.PutSlice(option.ID[:])
	buf.PutUint8(SessionStatusNew)
	buf.PutUint8(OptionNone)

	// tcp
	netType := byte(0x1)
	if option.Type == "udp" {
		netType = byte(0x2)
	}
	buf.PutUint8(netType)

	// port
	buf.PutUint16be(option.Port)

	// address
	ip, err := netip.ParseAddr(option.Host)
	if err != nil {
		buf.PutUint8(0x2)
		buf.PutString(option.Host)
	} else if ip.Is4() {
		buf.PutUint8(0x1)
		buf.PutSlice(ip.AsSlice())
	} else {
		buf.PutUint8(0x3)
		buf.PutSlice(ip.AsSlice())
	}

	metadata := buf.Bytes()
	binary.BigEndian.PutUint16(metadata[:2], uint16(len(metadata)-2))

	return &Mux{
		Conn: conn,
		id:   option.ID,
		otb:  metadata,
	}
}
