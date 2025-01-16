package server

import (
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"

	apiv1 "k8s.io/api/core/v1"
)

type BaremetalNodeGroupConfig struct {
	Name        string
	MinSizeConf int
	MaxSizeConf int
	Labels      []string
}

type BaremetalNodeGroup struct {
	nodes  map[string]*BaremetalNode
	client BaremetalClient

	BaremetalNodeGroupConfig
}

func NewBaremetalNodeGroup(config *BaremetalNodeGroupConfig) cloudprovider.NodeGroup {
	// TODO: implement creation of new NodeGroup
	return &BaremetalNodeGroup{}
}

func (bng *BaremetalNodeGroup) MaxSize() int {
	return bng.MaxSizeConf
}

func (bng *BaremetalNodeGroup) MinSize() int {
	return bng.MinSizeConf
}

func (bng *BaremetalNodeGroup) TargetSize() (int, error) {
	return (len(bng.nodes)), nil
}

func (bng *BaremetalNodeGroup) IncreaseSize(delta int) error {
	if len(bng.nodes) >= bng.MaxSize() {
		return fmt.Errorf("NodeGroup already has maximum number of members")
	}

	newNode, err := bng.client.AddNode(bng.Labels)
	if err != nil {
		return err
	}
	bng.nodes[newNode.ip.String()] = newNode

	return nil
}

func (bng *BaremetalNodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

func (bng *BaremetalNodeGroup) DeleteNodes(nodesToDel []*apiv1.Node) error {
	if len(bng.nodes)-len(nodesToDel) < bng.MaxSize() {
		return fmt.Errorf("NodeGroup would be smaller than minimum number of members")
	}

	for _, node := range bng.nodes {
		if err := bng.client.DeleteNode(node.name); err != nil {
			return err
		}
		delete(bng.nodes, node.name)
	}

	return nil
}

func (bng *BaremetalNodeGroup) DecreaseTargetSize(delta int) error {
	return nil
}

func (bng *BaremetalNodeGroup) ForceDeleteNodes([]*apiv1.Node) error {
	return cloudprovider.ErrNotImplemented
}

func (bng *BaremetalNodeGroup) Id() string {
	return bng.Name
}

func (bng *BaremetalNodeGroup) Debug() string {
	debug := fmt.Sprintf("name: %s, number of active nodes: %d, min size: %s, max size: %s",
		bng.Name, len(bng.nodes), bng.MinSizeConf, bng.MaxSizeConf)

	return debug
}

func (bng *BaremetalNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	return []cloudprovider.Instance{}, cloudprovider.ErrNotImplemented
}

func (bng *BaremetalNodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func (bng *BaremetalNodeGroup) Exist() bool {
	return true
}

func (bng *BaremetalNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func (bng *BaremetalNodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

func (bng *BaremetalNodeGroup) Autoprovisioned() bool {
	return false
}

func (bng *BaremetalNodeGroup) GetOptions(
	defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}
