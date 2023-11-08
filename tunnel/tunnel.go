package tunnel

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"runtime"
	"sync"
	"time"

	"github.com/MerlinKodo/clash-rev/adapter/inbound"
	"github.com/MerlinKodo/clash-rev/component/nat"
	P "github.com/MerlinKodo/clash-rev/component/process"
	"github.com/MerlinKodo/clash-rev/component/resolver"
	C "github.com/MerlinKodo/clash-rev/constant"
	"github.com/MerlinKodo/clash-rev/constant/provider"
	icontext "github.com/MerlinKodo/clash-rev/context"
	"github.com/MerlinKodo/clash-rev/log"
	"github.com/MerlinKodo/clash-rev/tunnel/statistic"

	"go.uber.org/atomic"
)

var (
	tcpQueue  = make(chan C.ConnContext, 200)
	udpQueue  = make(chan *inbound.PacketAdapter, 200)
	natTable  = nat.New()
	rules     []C.Rule
	proxies   = make(map[string]C.Proxy)
	providers map[string]provider.ProxyProvider
	configMux sync.RWMutex

	// Outbound Rule
	mode = Rule

	// default timeout for UDP session
	udpTimeout = 60 * time.Second

	// experimental feature
	UDPFallbackMatch = atomic.NewBool(false)
)

func init() {
	go process()
}

// TCPIn return fan-in queue
func TCPIn() chan<- C.ConnContext {
	return tcpQueue
}

// UDPIn return fan-in udp queue
func UDPIn() chan<- *inbound.PacketAdapter {
	return udpQueue
}

// Rules return all rules
func Rules() []C.Rule {
	return rules
}

// UpdateRules handle update rules
func UpdateRules(newRules []C.Rule) {
	configMux.Lock()
	rules = newRules
	configMux.Unlock()
}

// Proxies return all proxies
func Proxies() map[string]C.Proxy {
	return proxies
}

// Providers return all compatible providers
func Providers() map[string]provider.ProxyProvider {
	return providers
}

// UpdateProxies handle update proxies
func UpdateProxies(newProxies map[string]C.Proxy, newProviders map[string]provider.ProxyProvider) {
	configMux.Lock()
	proxies = newProxies
	providers = newProviders
	configMux.Unlock()
}

// Mode return current mode
func Mode() TunnelMode {
	return mode
}

// SetMode change the mode of tunnel
func SetMode(m TunnelMode) {
	mode = m
}

// processUDP starts a loop to handle udp packet
func processUDP() {
	queue := udpQueue
	for conn := range queue {
		handleUDPConn(conn)
	}
}

func process() {
	numUDPWorkers := 4
	if num := runtime.GOMAXPROCS(0); num > numUDPWorkers {
		numUDPWorkers = num
	}
	for i := 0; i < numUDPWorkers; i++ {
		go processUDP()
	}

	queue := tcpQueue
	for conn := range queue {
		go handleTCPConn(conn)
	}
}

func needLookupIP(metadata *C.Metadata) bool {
	return resolver.MappingEnabled() && metadata.Host == "" && metadata.DstIP != nil
}

func preHandleMetadata(metadata *C.Metadata) error {
	// handle IP string on host
	if ip := net.ParseIP(metadata.Host); ip != nil {
		metadata.DstIP = ip
		metadata.Host = ""
	}

	// preprocess enhanced-mode metadata
	if needLookupIP(metadata) {
		host, exist := resolver.FindHostByIP(metadata.DstIP)
		if exist {
			metadata.Host = host
			metadata.DNSMode = C.DNSMapping
			if resolver.FakeIPEnabled() {
				metadata.DstIP = nil
				metadata.DNSMode = C.DNSFakeIP
			} else if node := resolver.DefaultHosts.Search(host); node != nil {
				// redir-host should lookup the hosts
				metadata.DstIP = node.Data.(net.IP)
			}
		} else if resolver.IsFakeIP(metadata.DstIP) {
			return fmt.Errorf("fake DNS record %s missing", metadata.DstIP)
		}
	}

	return nil
}

func resolveMetadata(ctx C.PlainContext, metadata *C.Metadata) (proxy C.Proxy, rule C.Rule, err error) {
	if metadata.SpecialProxy != "" {
		var exist bool
		proxy, exist = proxies[metadata.SpecialProxy]
		if !exist {
			err = fmt.Errorf("proxy %s not found", metadata.SpecialProxy)
		}
		return
	}

	switch mode {
	case Direct:
		proxy = proxies["DIRECT"]
	case Global:
		proxy = proxies["GLOBAL"]
	case Rule:
		proxy, rule, err = match(metadata)
	default:
		panic(fmt.Sprintf("unknown mode: %s", mode))
	}

	return
}

func handleUDPConn(packet *inbound.PacketAdapter) {
	metadata := packet.Metadata()
	if !metadata.Valid() {
		packet.Drop()
		log.Warnln("[Metadata] not valid: %#v", metadata)
		return
	}

	// make a fAddr if request ip is fakeip
	var fAddr netip.Addr
	if resolver.IsExistFakeIP(metadata.DstIP) {
		fAddr, _ = netip.AddrFromSlice(metadata.DstIP)
		fAddr = fAddr.Unmap()
	}

	if err := preHandleMetadata(metadata); err != nil {
		packet.Drop()
		log.Debugln("[Metadata PreHandle] error: %s", err)
		return
	}

	// local resolve UDP dns
	if !metadata.Resolved() {
		ips, err := resolver.LookupIP(context.Background(), metadata.Host)
		if err != nil {
			packet.Drop()
			return
		} else if len(ips) == 0 {
			packet.Drop()
			return
		}
		metadata.DstIP = ips[0]
	}

	key := packet.LocalAddr().String()

	handle := func() bool {
		pc := natTable.Get(key)
		if pc != nil {
			handleUDPToRemote(packet, pc, metadata)
			return true
		}
		return false
	}

	if handle() {
		packet.Drop()
		return
	}

	lockKey := key + "-lock"
	cond, loaded := natTable.GetOrCreateLock(lockKey)

	go func() {
		defer packet.Drop()

		if loaded {
			cond.L.Lock()
			cond.Wait()
			handle()
			cond.L.Unlock()
			return
		}

		defer func() {
			natTable.Delete(lockKey)
			cond.Broadcast()
		}()

		pCtx := icontext.NewPacketConnContext(metadata)
		proxy, rule, err := resolveMetadata(pCtx, metadata)
		if err != nil {
			log.Warnln("[UDP] Parse metadata failed: %s", err.Error())
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), C.DefaultUDPTimeout)
		defer cancel()
		rawPc, err := proxy.ListenPacketContext(ctx, metadata.Pure())
		if err != nil {
			if rule == nil {
				log.Warnln(
					"[UDP] dial %s %s --> %s error: %s",
					proxy.Name(),
					metadata.SourceAddress(),
					metadata.RemoteAddress(),
					err.Error(),
				)
			} else {
				log.Warnln("[UDP] dial %s (match %s/%s) %s --> %s error: %s", proxy.Name(), rule.RuleType().String(), rule.Payload(), metadata.SourceAddress(), metadata.RemoteAddress(), err.Error())
			}
			return
		}
		pCtx.InjectPacketConn(rawPc)
		pc := statistic.NewUDPTracker(rawPc, statistic.DefaultManager, metadata, rule)

		switch true {
		case metadata.SpecialProxy != "":
			log.Infoln("[UDP] %s --> %s using %s", metadata.SourceAddress(), metadata.RemoteAddress(), metadata.SpecialProxy)
		case rule != nil:
			log.Infoln(
				"[UDP] %s --> %s match %s(%s) using %s",
				metadata.SourceAddress(),
				metadata.RemoteAddress(),
				rule.RuleType().String(),
				rule.Payload(),
				rawPc.Chains().String(),
			)
		case mode == Global:
			log.Infoln("[UDP] %s --> %s using GLOBAL", metadata.SourceAddress(), metadata.RemoteAddress())
		case mode == Direct:
			log.Infoln("[UDP] %s --> %s using DIRECT", metadata.SourceAddress(), metadata.RemoteAddress())
		default:
			log.Infoln(
				"[UDP] %s --> %s doesn't match any rule using DIRECT",
				metadata.SourceAddress(),
				metadata.RemoteAddress(),
			)
		}

		oAddr, _ := netip.AddrFromSlice(metadata.DstIP)
		oAddr = oAddr.Unmap()
		go handleUDPToLocal(packet.UDPPacket, pc, key, oAddr, fAddr)

		natTable.Set(key, pc)
		handle()
	}()
}

func handleTCPConn(connCtx C.ConnContext) {
	defer connCtx.Conn().Close()

	metadata := connCtx.Metadata()
	if !metadata.Valid() {
		log.Warnln("[Metadata] not valid: %#v", metadata)
		return
	}

	if err := preHandleMetadata(metadata); err != nil {
		log.Debugln("[Metadata PreHandle] error: %s", err)
		return
	}

	proxy, rule, err := resolveMetadata(connCtx, metadata)
	if err != nil {
		log.Warnln("[Metadata] parse failed: %s", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), C.DefaultTCPTimeout)
	defer cancel()
	remoteConn, err := proxy.DialContext(ctx, metadata.Pure())
	if err != nil {
		if rule == nil {
			log.Warnln(
				"[TCP] dial %s %s --> %s error: %s",
				proxy.Name(),
				metadata.SourceAddress(),
				metadata.RemoteAddress(),
				err.Error(),
			)
		} else {
			log.Warnln("[TCP] dial %s (match %s/%s) %s --> %s error: %s", proxy.Name(), rule.RuleType().String(), rule.Payload(), metadata.SourceAddress(), metadata.RemoteAddress(), err.Error())
		}
		return
	}
	remoteConn = statistic.NewTCPTracker(remoteConn, statistic.DefaultManager, metadata, rule)
	defer remoteConn.Close()

	switch true {
	case metadata.SpecialProxy != "":
		log.Infoln("[TCP] %s --> %s using %s", metadata.SourceAddress(), metadata.RemoteAddress(), metadata.SpecialProxy)
	case rule != nil:
		log.Infoln(
			"[TCP] %s --> %s match %s(%s) using %s",
			metadata.SourceAddress(),
			metadata.RemoteAddress(),
			rule.RuleType().String(),
			rule.Payload(),
			remoteConn.Chains().String(),
		)
	case mode == Global:
		log.Infoln("[TCP] %s --> %s using GLOBAL", metadata.SourceAddress(), metadata.RemoteAddress())
	case mode == Direct:
		log.Infoln("[TCP] %s --> %s using DIRECT", metadata.SourceAddress(), metadata.RemoteAddress())
	default:
		log.Infoln(
			"[TCP] %s --> %s doesn't match any rule using DIRECT",
			metadata.SourceAddress(),
			metadata.RemoteAddress(),
		)
	}

	handleSocket(connCtx, remoteConn)
}

func shouldResolveIP(rule C.Rule, metadata *C.Metadata) bool {
	return rule.ShouldResolveIP() && metadata.Host != "" && metadata.DstIP == nil
}

func match(metadata *C.Metadata) (C.Proxy, C.Rule, error) {
	configMux.RLock()
	defer configMux.RUnlock()

	var resolved bool
	var processFound bool

	if node := resolver.DefaultHosts.Search(metadata.Host); node != nil {
		ip := node.Data.(net.IP)
		metadata.DstIP = ip
		resolved = true
	}

	for _, rule := range rules {
		if !resolved && shouldResolveIP(rule, metadata) {
			ip, err := resolver.ResolveIP(metadata.Host)
			if err != nil {
				log.Debugln("[DNS] resolve %s error: %s", metadata.Host, err.Error())
			} else {
				log.Debugln("[DNS] %s --> %s", metadata.Host, ip.String())
				metadata.DstIP = ip
			}
			resolved = true
		}

		if !processFound && rule.ShouldFindProcess() {
			processFound = true

			srcIP, ok := netip.AddrFromSlice(metadata.SrcIP)
			if ok && metadata.OriginDst.IsValid() {
				srcIP = srcIP.Unmap()
				path, err := P.FindProcessPath(metadata.NetWork.String(), netip.AddrPortFrom(srcIP, uint16(metadata.SrcPort)), metadata.OriginDst)
				if err != nil {
					log.Debugln("[Process] find process %s: %v", metadata.String(), err)
				} else {
					log.Debugln("[Process] %s from process %s", metadata.String(), path)
					metadata.ProcessPath = path
				}
			}
		}

		if rule.Match(metadata) {
			adapter, ok := proxies[rule.Adapter()]
			if !ok {
				continue
			}

			if metadata.NetWork == C.UDP && !adapter.SupportUDP() && UDPFallbackMatch.Load() {
				log.Debugln("[Matcher] %s UDP is not supported, skip match", adapter.Name())
				continue
			}
			return adapter, rule, nil
		}
	}

	return proxies["DIRECT"], nil, nil
}
