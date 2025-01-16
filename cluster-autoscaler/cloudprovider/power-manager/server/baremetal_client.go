package server

import (
	"fmt"
	"net/netip"
	"os/exec"
	"regexp"

	klog "k8s.io/klog/v2"
)

// Proper client should use k8s.io/client-go
type BaremetalClient struct{}

func (c *BaremetalClient) AddNode(labels []string) (*BaremetalNode, error) {
	cmd := exec.Command("minikube -p test node add")
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	klog.V(5).Info("added new node")

	cmd = exec.Command("kubectl get nodes -o jsonpath='{.items[-1].metadata.labels.kubernetes\\.io/hostname}'")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	if !regexp.MustCompile(`^\S*$`).Match(output) {
		return nil, fmt.Errorf("retrieved node name in bad format, node: ", output)
	}
	nodeName := string(output)

	cmd = exec.Command(fmt.Sprintf(
		"kubectl get nodes %s -o jsonpath=\"{.status.addresses[?(@.type=='InternalIP')].address}\"", nodeName))
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
		cmd = exec.Command(fmt.Sprintf(cmdStr), label)
		if err := cmd.Run(); err != nil {
			klog.Error("error while labeling new node, node: ", nodeName, " label: ", label)
		}
	}
	klog.V(5).Info("labeled new node")

	return &BaremetalNode{name: nodeName, ip: ipAddr}, nil
}

func (c *BaremetalClient) DeleteNode(nodeName string) error {
	cmd := exec.Command(fmt.Sprintf("minikube -p test node delete %s", nodeName))
	if err := cmd.Run(); err != nil {
		return err
	}

	klog.V(5).Info("deleted node: ", nodeName)

	return nil
}
