//go:build freebsd

package process

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestEnforceStructValid12(t *testing.T) {
	if majorVersion != 12 && majorVersion != 13 {
		t.Skipf("Unsupported freebsd version: %d", majorVersion)

		return
	}

	assert.Equal(t, 0, int(unsafe.Offsetof(XTcpcb12{}.Len)))
	assert.Equal(t, 24, int(unsafe.Offsetof(XTcpcb12{}.SocketAddr)))
	assert.Equal(t, 116, int(unsafe.Offsetof(XTcpcb12{}.Family)))
	assert.Equal(t, 260, int(unsafe.Offsetof(XTcpcb12{}.InEndpoints)))
	assert.Equal(t, 0, int(unsafe.Offsetof(XInpcb12{}.Len)))
	assert.Equal(t, 16, int(unsafe.Offsetof(XInpcb12{}.SocketAddr)))
	assert.Equal(t, 108, int(unsafe.Offsetof(XInpcb12{}.Family)))
	assert.Equal(t, 252, int(unsafe.Offsetof(XInpcb12{}.InEndpoints)))
	assert.Equal(t, 0, int(unsafe.Offsetof(XFile12{}.Size)))
	assert.Equal(t, 8, int(unsafe.Offsetof(XFile12{}.Pid)))
	assert.Equal(t, 56, int(unsafe.Offsetof(XFile12{}.DataAddr)))
	assert.Equal(t, 64, int(unsafe.Sizeof(Xinpgen12{})))
	assert.Equal(t, 744, int(unsafe.Sizeof(XTcpcb12{})))
	assert.Equal(t, 400, int(unsafe.Sizeof(XInpcb12{})))
	assert.Equal(t, 40, int(unsafe.Sizeof(InEndpoints12{})))
	assert.Equal(t, 128, int(unsafe.Sizeof(XFile12{})))
}
