package profile

import (
	"github.com/MerlinKodo/clash-rev/common/atomic"
)

// StoreSelected is a global switch for storing selected proxy to cache
var StoreSelected = atomic.NewBool(true)
