package process

import (
	"encoding/binary"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

type Xinpgen12 [64]byte // size 64

type InEndpoints12 struct {
	FPort  [2]byte
	LPort  [2]byte
	FAddr  [16]byte
	LAddr  [16]byte
	ZoneID uint32
} // size 40

type XTcpcb12 struct {
	Len         uint32        // offset 0
	_           [20]byte      // offset 4
	SocketAddr  uint64        // offset 24
	_           [84]byte      // offset 32
	Family      uint32        // offset 116
	_           [140]byte     // offset 120
	InEndpoints InEndpoints12 // offset 260
	_           [444]byte     // offset 300
} // size 744

type XInpcb12 struct {
	Len         uint32        // offset 0
	_           [12]byte      // offset 4
	SocketAddr  uint64        // offset 16
	_           [84]byte      // offset 24
	Family      uint32        // offset 108
	_           [140]byte     // offset 112
	InEndpoints InEndpoints12 // offset 252
	_           [108]byte     // offset 292
} // size 400

type XFile12 struct {
	Size     uint64   // offset 0
	Pid      uint32   // offset 8
	_        [44]byte // offset 12
	DataAddr uint64   // offset 56
	_        [64]byte // offset 64
} // size 128

var majorVersion = func() int {
	releaseVersion, err := unix.Sysctl("kern.osrelease")
	if err != nil {
		return 0
	}

	majorVersionText, _, _ := strings.Cut(releaseVersion, ".")

	majorVersion, err := strconv.Atoi(majorVersionText)
	if err != nil {
		return 0
	}

	return majorVersion
}()

func findProcessPath(network string, from netip.AddrPort, to netip.AddrPort) (string, error) {
	switch majorVersion {
	case 12, 13:
		return findProcessPath12(network, from, to)
	}

	return "", ErrPlatformNotSupport
}

func findProcessPath12(network string, from netip.AddrPort, to netip.AddrPort) (string, error) {
	switch network {
	case TCP:
		data, err := unix.SysctlRaw("net.inet.tcp.pcblist")
		if err != nil {
			return "", err
		}

		if len(data) < int(unsafe.Sizeof(Xinpgen12{})) {
			return "", fmt.Errorf("invalid sysctl data len: %d", len(data))
		}

		data = data[unsafe.Sizeof(Xinpgen12{}):]

		for len(data) > int(unsafe.Sizeof(XTcpcb12{}.Len)) {
			tcb := (*XTcpcb12)(unsafe.Pointer(&data[0]))
			if tcb.Len < uint32(unsafe.Sizeof(XTcpcb12{})) || uint32(len(data)) < tcb.Len {
				break
			}

			data = data[tcb.Len:]

			var (
				connFromAddr netip.Addr
				connToAddr   netip.Addr
			)
			if tcb.Family == unix.AF_INET {
				connFromAddr = netip.AddrFrom4([4]byte(tcb.InEndpoints.LAddr[12:16]))
				connToAddr = netip.AddrFrom4([4]byte(tcb.InEndpoints.FAddr[12:16]))
			} else if tcb.Family == unix.AF_INET6 {
				connFromAddr = netip.AddrFrom16(tcb.InEndpoints.LAddr)
				connToAddr = netip.AddrFrom16(tcb.InEndpoints.FAddr)
			} else {
				continue
			}

			connFrom := netip.AddrPortFrom(connFromAddr, binary.BigEndian.Uint16(tcb.InEndpoints.LPort[:]))
			connTo := netip.AddrPortFrom(connToAddr, binary.BigEndian.Uint16(tcb.InEndpoints.FPort[:]))

			if connFrom == from && connTo == to {
				pid, err := findPidBySocketAddr12(tcb.SocketAddr)
				if err != nil {
					return "", err
				}

				return findExecutableByPid(pid)
			}
		}
	case UDP:
		data, err := unix.SysctlRaw("net.inet.udp.pcblist")
		if err != nil {
			return "", err
		}

		if len(data) < int(unsafe.Sizeof(Xinpgen12{})) {
			return "", fmt.Errorf("invalid sysctl data len: %d", len(data))
		}

		data = data[unsafe.Sizeof(Xinpgen12{}):]

		for len(data) > int(unsafe.Sizeof(XInpcb12{}.Len)) {
			icb := (*XInpcb12)(unsafe.Pointer(&data[0]))
			if icb.Len < uint32(unsafe.Sizeof(XInpcb12{})) || uint32(len(data)) < icb.Len {
				break
			}
			data = data[icb.Len:]

			var connFromAddr netip.Addr
			if icb.Family == unix.AF_INET {
				connFromAddr = netip.AddrFrom4([4]byte(icb.InEndpoints.LAddr[12:16]))
			} else if icb.Family == unix.AF_INET6 {
				connFromAddr = netip.AddrFrom16(icb.InEndpoints.LAddr)
			} else {
				continue
			}

			connFromPort := binary.BigEndian.Uint16(icb.InEndpoints.LPort[:])

			if (connFromAddr == from.Addr() || connFromAddr.IsUnspecified()) && connFromPort == from.Port() {
				pid, err := findPidBySocketAddr12(icb.SocketAddr)
				if err != nil {
					return "", err
				}

				return findExecutableByPid(pid)
			}
		}
	}

	return "", ErrNotFound
}

func findPidBySocketAddr12(socketAddr uint64) (uint32, error) {
	buf, err := unix.SysctlRaw("kern.file")
	if err != nil {
		return 0, err
	}

	filesLen := len(buf) / int(unsafe.Sizeof(XFile12{}))
	files := unsafe.Slice((*XFile12)(unsafe.Pointer(&buf[0])), filesLen)

	for _, file := range files {
		if file.Size != uint64(unsafe.Sizeof(XFile12{})) {
			return 0, fmt.Errorf("invalid xfile size: %d", file.Size)
		}

		if file.DataAddr == socketAddr {
			return file.Pid, nil
		}
	}

	return 0, ErrNotFound
}

func findExecutableByPid(pid uint32) (string, error) {
	buf := make([]byte, unix.PathMax)
	size := uint64(len(buf))
	mib := [4]uint32{
		unix.CTL_KERN,
		14, // KERN_PROC
		12, // KERN_PROC_PATHNAME
		pid,
	}

	_, _, errno := unix.Syscall6(
		unix.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		uintptr(len(mib)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
		0,
		0,
	)
	if errno != 0 || size == 0 {
		return "", fmt.Errorf("sysctl: get proc name: %w", errno)
	}

	return string(buf[:size-1]), nil
}
