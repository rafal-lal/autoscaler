package server

import (
	"net/netip"
	"sync"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

// UID would be needed for identification
type BaremetalNode struct {
	ID     string
	IP     netip.Addr
	Status cloudprovider.InstanceStatus
	mutex  sync.Mutex
}
