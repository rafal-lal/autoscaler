package server

import (
	"context"
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/externalgrpc/protos"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klog "k8s.io/klog/v2"
)

type BaremetalProvider struct {
	protos.UnimplementedCloudProviderServer
	nodeGroups []cloudprovider.NodeGroup
}

// NewBaremetalProvider creates a grpc wrapper for a cloud provider implementation.
func NewBaremetalProvider(nodeGroupConfig *BaremetalNodeGroupConfig) *BaremetalProvider {
	return &BaremetalProvider{
		nodeGroups: []cloudprovider.NodeGroup{
			NewBaremetalNodeGroup(nodeGroupConfig),
		},
	}
}

// apiv1Node converts a protos.ExternalGrpcNode to a apiv1.Node.
func apiv1Node(pbNode *protos.ExternalGrpcNode) *apiv1.Node {
	apiv1Node := &apiv1.Node{}
	apiv1Node.ObjectMeta = metav1.ObjectMeta{
		Name:        pbNode.GetName(),
		Annotations: pbNode.GetAnnotations(),
		Labels:      pbNode.GetLabels(),
	}
	apiv1Node.Spec = apiv1.NodeSpec{
		ProviderID: pbNode.GetProviderID(),
	}
	return apiv1Node
}

// apiv1Node converts an apiv1.Node to a protos.ExternalGrpcNode.
func pbNodeGroup(ng cloudprovider.NodeGroup) *protos.NodeGroup {
	return &protos.NodeGroup{
		Id:      ng.Id(),
		MaxSize: int32(ng.MaxSize()),
		MinSize: int32(ng.MinSize()),
		Debug:   ng.Debug(),
	}
}

func debug(req fmt.Stringer) {
	klog.V(5).Infof("got gRPC request: %T %s", req, req)
}

// TODO: implement the rest

func (bp *BaremetalProvider) NodeGroups(_ context.Context,
	req *protos.NodeGroupsRequest) (*protos.NodeGroupsResponse, error) {
	debug(req)

	return nil, nil
}

func (bp *BaremetalProvider) GPULabel(_ context.Context, req *protos.GPULabelRequest) (*protos.GPULabelResponse, error) {
	debug(req)
	return &protos.GPULabelResponse{
		Label: "hello world",
	}, nil
}
