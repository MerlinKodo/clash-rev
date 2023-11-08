package rules

import (
	"fmt"
	"strconv"

	C "github.com/MerlinKodo/clash-rev/constant"
)

type PortType int

const (
	PortTypeSrc PortType = iota
	PortTypeDest
	PortTypeInbound
)

// Implements C.Rule
var _ C.Rule = (*Port)(nil)

type Port struct {
	adapter  string
	port     C.Port
	portType PortType
}

func (p *Port) RuleType() C.RuleType {
	switch p.portType {
	case PortTypeSrc:
		return C.SrcPort
	case PortTypeDest:
		return C.DstPort
	case PortTypeInbound:
		return C.InboundPort
	default:
		panic(fmt.Errorf("unknown port type: %v", p.portType))
	}
}

func (p *Port) Match(metadata *C.Metadata) bool {
	switch p.portType {
	case PortTypeSrc:
		return metadata.SrcPort == p.port
	case PortTypeDest:
		return metadata.DstPort == p.port
	case PortTypeInbound:
		return metadata.OriginDst.Port() == uint16(p.port)
	default:
		panic(fmt.Errorf("unknown port type: %v", p.portType))
	}
}

func (p *Port) Adapter() string {
	return p.adapter
}

func (p *Port) Payload() string {
	return p.port.String()
}

func (p *Port) ShouldResolveIP() bool {
	return false
}

func (p *Port) ShouldFindProcess() bool {
	return false
}

func NewPort(port string, adapter string, portType PortType) (*Port, error) {
	p, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return nil, errPayload
	}
	return &Port{
		adapter:  adapter,
		port:     C.Port(p),
		portType: portType,
	}, nil
}
