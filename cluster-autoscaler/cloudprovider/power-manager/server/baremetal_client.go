package server

import (
	"fmt"
	"net/netip"
	"os/exec"
	"strings"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	klog "k8s.io/klog/v2"
)

// Proper client should use k8s.io/client-go
type BaremetalClient struct{}

func (c *BaremetalClient) AddNewNode(labels []string) (*BaremetalNode, error) {
	newBaremetalNode := &BaremetalNode{Status: cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating}}

	exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "-i", "/root/minikube-node-key", "vscode@172.17.0.2",
		"minikube -p test node add --worker=true").
		Run()

	klog.V(5).Info("added new node")

	cmd := exec.Command("kubectl", "get", "nodes", "-o",
		"jsonpath={.items[-1].metadata.labels.kubernetes\\.io/hostname}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		klog.Error("error getting nodename", err)
	}
	nodeName := string(output)

	cmd = exec.Command("kubectl", "get", "nodes", nodeName, "-o",
		"jsonpath=\"{.status.addresses[?(@.type=='InternalIP')].address}\"")
	output, err = cmd.CombinedOutput()
	if err != nil {
		klog.Error("error getting ip", err)
	}
	ipAddr, err := netip.ParseAddr(strings.Trim(string(output), "\""))
	if err != nil {
		klog.Error("error parsing ip", err)
	}

	for _, label := range labels {
		cmd = exec.Command("kubectl", "label", "node", nodeName, label)
		if err := cmd.Run(); err != nil {
			klog.Error("error while labeling new node, node: ", nodeName, " label: ", label)
		}
	}
	klog.V(5).Info("labeled new node")
	newBaremetalNode.mutex.Lock()
	newBaremetalNode.ID = nodeName
	newBaremetalNode.IP = ipAddr
	newBaremetalNode.Status = cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}
	newBaremetalNode.mutex.Unlock()
	klog.V(5).Info("new node ready")

	return newBaremetalNode, nil
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
	go func() {
		time.Sleep(10 * time.Second)
		exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "-i", "/root/minikube-node-key", "vscode@172.17.0.2",
			fmt.Sprintf("minikube -p test node delete %s", nodeID)).
			Run()
	}()

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
