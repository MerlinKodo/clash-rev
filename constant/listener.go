package constant

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
)

type Listener interface {
	RawAddress() string
	Address() string
	Close() error
}

type InboundType string

const (
	InboundTypeSocks  InboundType = "socks"
	InboundTypeRedir  InboundType = "redir"
	InboundTypeTproxy InboundType = "tproxy"
	InboundTypeHTTP   InboundType = "http"
	InboundTypeMixed  InboundType = "mixed"
)

var supportInboundTypes = map[InboundType]bool{
	InboundTypeSocks:  true,
	InboundTypeRedir:  true,
	InboundTypeTproxy: true,
	InboundTypeHTTP:   true,
	InboundTypeMixed:  true,
}

type inbound struct {
	Type          InboundType `json:"type" yaml:"type"`
	BindAddress   string      `json:"bind-address" yaml:"bind-address"`
	IsFromPortCfg bool        `json:"-" yaml:"-"`
}

// Inbound
type Inbound inbound

// UnmarshalYAML implements yaml.Unmarshaler
func (i *Inbound) UnmarshalYAML(unmarshal func(any) error) error {
	var tp string
	if err := unmarshal(&tp); err != nil {
		var inner inbound
		if err := unmarshal(&inner); err != nil {
			return err
		}

		*i = Inbound(inner)
	} else {
		inner, err := parseInbound(tp)
		if err != nil {
			return err
		}

		*i = Inbound(*inner)
	}

	if !supportInboundTypes[i.Type] {
		return fmt.Errorf("not support inbound type: %s", i.Type)
	}
	_, portStr, err := net.SplitHostPort(i.BindAddress)
	if err != nil {
		return fmt.Errorf("bind address parse error. addr: %s, err: %w", i.BindAddress, err)
	}
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil || port == 0 {
		return fmt.Errorf("invalid bind port. addr: %s", i.BindAddress)
	}
	return nil
}

func parseInbound(alias string) (*inbound, error) {
	u, err := url.Parse(alias)
	if err != nil {
		return nil, err
	}
	listenerType := InboundType(u.Scheme)
	return &inbound{
		Type:        listenerType,
		BindAddress: u.Host,
	}, nil
}

func (i *Inbound) ToAlias() string {
	return string(i.Type) + "://" + i.BindAddress
}
