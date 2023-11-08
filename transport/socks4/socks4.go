package socks4

import (
	"errors"
	"io"
	"net"
	"net/netip"
	"strconv"

	"github.com/MerlinKodo/clash-rev/component/auth"

	"github.com/MerlinKodo/protobytes"
)

const Version = 0x04

type Command = uint8

const (
	CmdConnect Command = 0x01
	CmdBind    Command = 0x02
)

type Code = uint8

const (
	RequestGranted          Code = 90
	RequestRejected         Code = 91
	RequestIdentdFailed     Code = 92
	RequestIdentdMismatched Code = 93
)

var (
	errVersionMismatched   = errors.New("version code mismatched")
	errCommandNotSupported = errors.New("command not supported")
	errIPv6NotSupported    = errors.New("IPv6 not supported")

	ErrRequestRejected         = errors.New("request rejected or failed")
	ErrRequestIdentdFailed     = errors.New("request rejected because SOCKS server cannot connect to identd on the client")
	ErrRequestIdentdMismatched = errors.New("request rejected because the client program and identd report different user-ids")
	ErrRequestUnknownCode      = errors.New("request failed with unknown code")
)

func ServerHandshake(rw io.ReadWriter, authenticator auth.Authenticator) (addr string, command Command, err error) {
	var req [8]byte
	if _, err = io.ReadFull(rw, req[:]); err != nil {
		return
	}

	r := protobytes.BytesReader(req[:])
	if r.ReadUint8() != Version {
		err = errVersionMismatched
		return
	}

	if command = r.ReadUint8(); command != CmdConnect {
		err = errCommandNotSupported
		return
	}

	var (
		host   string
		port   string
		code   uint8
		userID []byte
	)
	if userID, err = readUntilNull(rw); err != nil {
		return
	}

	dstPort := r.ReadUint16be()
	dstAddr := r.ReadIPv4()
	if isReservedIP(dstAddr) {
		var target []byte
		if target, err = readUntilNull(rw); err != nil {
			return
		}
		host = string(target)
	}

	port = strconv.Itoa(int(dstPort))
	if host != "" {
		addr = net.JoinHostPort(host, port)
	} else {
		addr = net.JoinHostPort(dstAddr.String(), port)
	}

	// SOCKS4 only support USERID auth.
	if authenticator == nil || authenticator.Verify(string(userID), "") {
		code = RequestGranted
	} else {
		code = RequestIdentdMismatched
		err = ErrRequestIdentdMismatched
	}

	reply := protobytes.BytesWriter(make([]byte, 0, 8))
	reply.PutUint8(0)    // reply code
	reply.PutUint8(code) // result code
	reply.PutUint16be(dstPort)
	reply.PutSlice(dstAddr.AsSlice())

	_, wErr := rw.Write(reply.Bytes())
	if err == nil {
		err = wErr
	}
	return
}

func ClientHandshake(rw io.ReadWriter, addr string, command Command, userID string) (err error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return err
	}

	ip, err := netip.ParseAddr(host)
	if err != nil { // Host
		ip = netip.AddrFrom4([4]byte{0, 0, 0, 1})
	} else if ip.Is6() { // IPv6
		return errIPv6NotSupported
	}

	req := protobytes.BytesWriter{}
	req.PutUint8(Version)
	req.PutUint8(command)
	req.PutUint16be(uint16(port))
	req.PutSlice(ip.AsSlice())
	req.PutString(userID)
	req.PutUint8(0) /* NULL */

	if isReservedIP(ip) /* SOCKS4A */ {
		req.PutString(host)
		req.PutUint8(0) /* NULL */
	}

	if _, err = rw.Write(req.Bytes()); err != nil {
		return err
	}

	var resp [8]byte
	if _, err = io.ReadFull(rw, resp[:]); err != nil {
		return err
	}

	if resp[0] != 0x00 {
		return errVersionMismatched
	}

	switch resp[1] {
	case RequestGranted:
		return nil
	case RequestRejected:
		return ErrRequestRejected
	case RequestIdentdFailed:
		return ErrRequestIdentdFailed
	case RequestIdentdMismatched:
		return ErrRequestIdentdMismatched
	default:
		return ErrRequestUnknownCode
	}
}

// For version 4A, if the client cannot resolve the destination host's
// domain name to find its IP address, it should set the first three bytes
// of DSTIP to NULL and the last byte to a non-zero value. (This corresponds
// to IP address 0.0.0.x, with x nonzero. As decreed by IANA  -- The
// Internet Assigned Numbers Authority -- such an address is inadmissible
// as a destination IP address and thus should never occur if the client
// can resolve the domain name.)
func isReservedIP(ip netip.Addr) bool {
	subnet := netip.PrefixFrom(
		netip.AddrFrom4([4]byte{0, 0, 0, 0}),
		24,
	)

	return !ip.IsUnspecified() && subnet.Contains(ip)
}

func readUntilNull(r io.Reader) ([]byte, error) {
	buf := protobytes.BytesWriter{}
	var data [1]byte

	for {
		if _, err := r.Read(data[:]); err != nil {
			return nil, err
		}
		if data[0] == 0 {
			return buf.Bytes(), nil
		}
		buf.PutUint8(data[0])
	}
}
