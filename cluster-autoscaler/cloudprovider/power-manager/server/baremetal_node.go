package server

import (
	"net/netip"
)

// UID would be needed for identification
type BaremetalNode struct {
	name string
	ip   netip.Addr
}
