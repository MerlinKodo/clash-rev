package context

import (
	"net"

	C "github.com/MerlinKodo/clash-rev/constant"

	"github.com/gofrs/uuid/v5"
)

type ConnContext struct {
	id       uuid.UUID
	metadata *C.Metadata
	conn     net.Conn
}

func NewConnContext(conn net.Conn, metadata *C.Metadata) *ConnContext {
	id, _ := uuid.NewV4()
	return &ConnContext{
		id:       id,
		metadata: metadata,
		conn:     conn,
	}
}

// ID implement C.ConnContext ID
func (c *ConnContext) ID() uuid.UUID {
	return c.id
}

// Metadata implement C.ConnContext Metadata
func (c *ConnContext) Metadata() *C.Metadata {
	return c.metadata
}

// Conn implement C.ConnContext Conn
func (c *ConnContext) Conn() net.Conn {
	return c.conn
}
