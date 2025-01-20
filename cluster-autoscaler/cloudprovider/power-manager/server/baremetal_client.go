package server

import (
	"fmt"
	"net/netip"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	klog "k8s.io/klog/v2"
)

// Proper client should use k8s.io/client-go
type BaremetalClient struct{}

func (c *BaremetalClient) AddNewNode(labels []string) (*BaremetalNode, error) {
	cmd := exec.Command("minikube", "-p", "test", "node", "add")
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	klog.V(5).Info("added new node")

	cmd = exec.Command("kubectl", "get", "nodes", "-o",
		"jsonpath='{.items[-1].metadata.labels.kubernetes\\.io/hostname}'")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	if !regexp.MustCompile(`^\S*$`).Match(output) {
		return nil, fmt.Errorf("retrieved node name in bad format, node: %s", output)
	}
	nodeName := string(output)

	cmd = exec.Command("kubectl", "get", "nodes", nodeName, "-o",
		"jsonpath=\"{.status.addresses[?(@.type=='InternalIP')].address}\"")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	ipAddr, err := netip.ParseAddr(string(output))
	if err != nil {
		return nil, err
	}

	cmdStr := "kubectl label node %s"
	for _, label := range labels {
		cmd = exec.Command(fmt.Sprintf(cmdStr, label))
		if err := cmd.Run(); err != nil {
			klog.Error("error while labeling new node, node: ", nodeName, " label: ", label)
		}
	}
	klog.V(5).Info("labeled new node")

	return &BaremetalNode{ID: nodeName, IP: ipAddr, Status: cloudprovider.InstanceStatus{
		State:     cloudprovider.InstanceRunning,
		ErrorInfo: nil,
	},
	}, nil
}

func (c *BaremetalClient) AddExistingNodeToNodeGroup(nodeName string) (*BaremetalNode, error) {
	klog.V(5).Info("nodename ", nodeName)
	time.Sleep(1 * time.Second)

	cmd := exec.Command("kubectl", "get", "nodes", nodeName, "-o",
		"jsonpath=\"{.status.addresses[?(@.type=='InternalIP')].address}\"")
	output, err := cmd.CombinedOutput()
	klog.V(5).Info("ip output ", string(output))
	if err != nil {
		klog.Error("not nil err ", err)
		return nil, err
	}
	ipAddr, err := netip.ParseAddr(strings.Trim(string(output), "\""))
	if err != nil {
		return nil, err
	}
	klog.V(5).Info("added existing node to node group")

	return &BaremetalNode{ID: nodeName, IP: ipAddr, Status: cloudprovider.InstanceStatus{
		State:     cloudprovider.InstanceRunning,
		ErrorInfo: nil,
	},
	}, nil
}

func (c *BaremetalClient) DeleteNode(nodeID string) error {
	// cmd := exec.Command(fmt.Sprintf("minikube -p test node delete %s", nodeID))
	// if err := cmd.Run(); err != nil {
	// 	return err
	// }

	klog.V(5).Info("deleted node: ", nodeID)

	return nil
}

func (c *BaremetalClient) GetNodesNames() ([]string, error) {
	cmd := exec.Command("kubectl", "get", "nodes", "-o", "jsonpath={.items[*].metadata.name}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	nodes := strings.Split(string(output), " ")

	klog.V(5).Info("retrieved nodes: ", nodes)

	return nodes, nil
}
