package server

import (
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"

	apiv1 "k8s.io/api/core/v1"
	klog "k8s.io/klog/v2"
)

type BaremetalNodeGroupConfig struct {
	ID          string
	MinSizeConf int
	MaxSizeConf int
	Labels      []string
}

type BaremetalNodeGroup struct {
	nodes  map[string]*BaremetalNode
	client BaremetalClient

	*BaremetalNodeGroupConfig
}

func NewBaremetalNodeGroup(config *BaremetalNodeGroupConfig) cloudprovider.NodeGroup {
	nodes := make(map[string]*BaremetalNode)
	client := BaremetalClient{}

	if nodesNames, err := client.GetNodesNames(); err == nil {
		for _, nodeName := range nodesNames {
			node, err := client.AddExistingNodeToNodeGroup(nodeName)
			if err != nil {
				klog.Error(err)
				continue
			}
			nodes[node.ID] = node
		}
	} else {
		klog.Error(err)
		klog.Exit("couldn't add existing nodes to default nodegroup")
	}

	return &BaremetalNodeGroup{
		nodes:                    nodes,
		client:                   BaremetalClient{},
		BaremetalNodeGroupConfig: config,
	}
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

	for range delta {
		newNode, err := bng.client.AddNewNode(bng.Labels)
		if err != nil {
			return err
		}
		bng.nodes[newNode.ID] = newNode
	}

	return nil
}

func (bng *BaremetalNodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

func (bng *BaremetalNodeGroup) DeleteNodes(nodesToDel []*apiv1.Node) error {
	if len(bng.nodes)-len(nodesToDel) < bng.MinSize() {
		return fmt.Errorf("NodeGroup would be smaller than minimum number of members")
	}

	for _, node := range nodesToDel {
		if err := bng.client.DeleteNode(node.Name); err != nil {
			return err
		}
		delete(bng.nodes, node.Name)
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
	return bng.ID
}

func (bng *BaremetalNodeGroup) Debug() string {
	return fmt.Sprintf("id: %s, number of active nodes: %d, min size: %d, max size: %d",
		bng.ID, len(bng.nodes), bng.MinSizeConf, bng.MaxSizeConf)
}

func (bng *BaremetalNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	instances := make([]cloudprovider.Instance, 0)
	for _, node := range bng.nodes {
		instances = append(instances, cloudprovider.Instance{Id: node.ID, Status: &node.Status})
	}

	return instances, nil
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
