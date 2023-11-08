package process

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"
	"os"
	"unsafe"

	"github.com/MerlinKodo/clash-rev/common/pool"

	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"
)

type inetDiagRequest struct {
	Family   byte
	Protocol byte
	Ext      byte
	Pad      byte
	States   uint32

	SrcPort [2]byte
	DstPort [2]byte
	Src     [16]byte
	Dst     [16]byte
	If      uint32
	Cookie  [2]uint32
}

type inetDiagResponse struct {
	Family  byte
	State   byte
	Timer   byte
	ReTrans byte

	SrcPort [2]byte
	DstPort [2]byte
	Src     [16]byte
	Dst     [16]byte
	If      uint32
	Cookie  [2]uint32

	Expires uint32
	RQueue  uint32
	WQueue  uint32
	UID     uint32
	INode   uint32
}

func findProcessPath(network string, from netip.AddrPort, to netip.AddrPort) (string, error) {
	inode, uid, err := resolveSocketByNetlink(network, from, to)
	if err != nil {
		return "", err
	}

	return resolveProcessPathByProcSearch(inode, uid)
}

func resolveSocketByNetlink(network string, from netip.AddrPort, to netip.AddrPort) (inode uint32, uid uint32, err error) {
	var families []byte
	if from.Addr().Unmap().Is4() {
		families = []byte{unix.AF_INET, unix.AF_INET6}
	} else {
		families = []byte{unix.AF_INET6, unix.AF_INET}
	}

	var protocol byte
	switch network {
	case TCP:
		protocol = unix.IPPROTO_TCP
	case UDP:
		protocol = unix.IPPROTO_UDP
	default:
		return 0, 0, ErrInvalidNetwork
	}

	if protocol == unix.IPPROTO_UDP {
		// Swap from & to for udp
		// See also https://www.mail-archive.com/netdev@vger.kernel.org/msg248638.html
		from, to = to, from
	}

	for _, family := range families {
		inode, uid, err = resolveSocketByNetlinkExact(family, protocol, from, to, netlink.Request)
		if err == nil {
			return inode, uid, err
		}
	}

	return 0, 0, ErrNotFound
}

func resolveSocketByNetlinkExact(family byte, protocol byte, from netip.AddrPort, to netip.AddrPort, flags netlink.HeaderFlags) (inode uint32, uid uint32, err error) {
	request := &inetDiagRequest{
		Family:   family,
		Protocol: protocol,
		States:   0xffffffff,
		Cookie:   [2]uint32{0xffffffff, 0xffffffff},
	}

	var (
		fromAddr []byte
		toAddr   []byte
	)
	if family == unix.AF_INET {
		fromAddr = net.IP(from.Addr().AsSlice()).To4()
		toAddr = net.IP(to.Addr().AsSlice()).To4()
	} else {
		fromAddr = net.IP(from.Addr().AsSlice()).To16()
		toAddr = net.IP(to.Addr().AsSlice()).To16()
	}

	copy(request.Src[:], fromAddr)
	copy(request.Dst[:], toAddr)

	binary.BigEndian.PutUint16(request.SrcPort[:], from.Port())
	binary.BigEndian.PutUint16(request.DstPort[:], to.Port())

	conn, err := netlink.Dial(unix.NETLINK_INET_DIAG, nil)
	if err != nil {
		return 0, 0, err
	}
	defer conn.Close()

	message := netlink.Message{
		Header: netlink.Header{
			Type:  20, // SOCK_DIAG_BY_FAMILY
			Flags: flags,
		},
		Data: (*(*[unsafe.Sizeof(*request)]byte)(unsafe.Pointer(request)))[:],
	}

	messages, err := conn.Execute(message)
	if err != nil {
		return 0, 0, err
	}

	for _, msg := range messages {
		if len(msg.Data) < int(unsafe.Sizeof(inetDiagResponse{})) {
			continue
		}

		response := (*inetDiagResponse)(unsafe.Pointer(&msg.Data[0]))

		return response.INode, response.UID, nil
	}

	return 0, 0, ErrNotFound
}

func resolveProcessPathByProcSearch(inode, uid uint32) (string, error) {
	procDir, err := os.Open("/proc")
	if err != nil {
		return "", err
	}
	defer procDir.Close()

	pids, err := procDir.Readdirnames(-1)
	if err != nil {
		return "", err
	}

	expectedSocketName := fmt.Appendf(nil, "socket:[%d]", inode)

	pathBuffer := pool.Get(64)
	defer pool.Put(pathBuffer)

	readlinkBuffer := pool.Get(32)
	defer pool.Put(readlinkBuffer)

	copy(pathBuffer, "/proc/")

	for _, pid := range pids {
		if !isPid(pid) {
			continue
		}

		pathBuffer = append(pathBuffer[:len("/proc/")], pid...)

		stat := &unix.Stat_t{}
		err = unix.Stat(string(pathBuffer), stat)
		if err != nil {
			continue
		} else if stat.Uid != uid {
			continue
		}

		pathBuffer = append(pathBuffer, "/fd/"...)
		fdsPrefixLength := len(pathBuffer)

		fdDir, err := os.Open(string(pathBuffer))
		if err != nil {
			continue
		}

		fds, err := fdDir.Readdirnames(-1)
		fdDir.Close()
		if err != nil {
			continue
		}

		for _, fd := range fds {
			pathBuffer = pathBuffer[:fdsPrefixLength]

			pathBuffer = append(pathBuffer, fd...)

			n, err := unix.Readlink(string(pathBuffer), readlinkBuffer)
			if err != nil {
				continue
			}

			if bytes.Equal(readlinkBuffer[:n], expectedSocketName) {
				return os.Readlink("/proc/" + pid + "/exe")
			}
		}
	}

	return "", fmt.Errorf("inode %d of uid %d not found", inode, uid)
}

func isPid(name string) bool {
	for _, c := range name {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}
