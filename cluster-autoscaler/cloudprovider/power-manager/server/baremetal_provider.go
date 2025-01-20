package server

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"
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
	klog.V(5).Info(ng.Id(), " ", ng.MaxSize(), " ", ng.MinSize(), " ", ng.Debug())
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

func (bp *BaremetalProvider) NodeGroups(_ context.Context, req *protos.NodeGroupsRequest,
) (*protos.NodeGroupsResponse, error) {
	debug(req)

	pbNgs := make([]*protos.NodeGroup, 0)
	for _, ng := range bp.nodeGroups {
		pbNgs = append(pbNgs, pbNodeGroup(ng))
	}

	return &protos.NodeGroupsResponse{
		NodeGroups: pbNgs,
	}, nil
}

func (bp *BaremetalProvider) NodeGroupForNode(_ context.Context, req *protos.NodeGroupForNodeRequest,
) (*protos.NodeGroupForNodeResponse, error) {
	debug(req)

	pbNode := req.GetNode()
	if pbNode == nil {
		return nil, fmt.Errorf("request fields were nil")
	}

	node := apiv1Node(pbNode)
	for _, ng := range bp.nodeGroups {
		if ngNodes, err := ng.Nodes(); err != nil {
			klog.Error(err)
			return nil, err
		} else {
			for _, ngNode := range ngNodes {
				klog.V(5).Info("node name ", node.Name, " cloud node ", ngNode.Id)
				if ngNode.Id == node.Name {
					klog.V(5).Info("node group ", ng)
					return &protos.NodeGroupForNodeResponse{
						NodeGroup: pbNodeGroup(ng),
					}, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("node group for node %s not found", pbNode.Name)
}

func (bp *BaremetalProvider) PricingNodePrice(_ context.Context, req *protos.PricingNodePriceRequest,
) (*protos.PricingNodePriceResponse, error) {
	debug(req)

	return &protos.PricingNodePriceResponse{Price: 1.0}, nil
}

func (bp *BaremetalProvider) PricingPodPrice(_ context.Context, req *protos.PricingPodPriceRequest,
) (*protos.PricingPodPriceResponse, error) {
	debug(req)

	return &protos.PricingPodPriceResponse{Price: 1.0}, nil
}

func (bp *BaremetalProvider) GPULabel(_ context.Context, req *protos.GPULabelRequest) (*protos.GPULabelResponse, error) {
	debug(req)
	return &protos.GPULabelResponse{
		Label: "gpu-present",
	}, nil
}

func (bp *BaremetalProvider) GetAvailableGPUTypes(_ context.Context, req *protos.GetAvailableGPUTypesRequest,
) (*protos.GetAvailableGPUTypesResponse, error) {
	debug(req)

	return &protos.GetAvailableGPUTypesResponse{
		GpuTypes: map[string]*anypb.Any{
			"gpu-type1": nil,
		},
	}, nil
}

func (bp *BaremetalProvider) Cleanup(_ context.Context, req *protos.CleanupRequest) (*protos.CleanupResponse, error) {
	debug(req)

	return &protos.CleanupResponse{}, nil
}

func (bp *BaremetalProvider) Refresh(_ context.Context, req *protos.RefreshRequest) (*protos.RefreshResponse, error) {
	debug(req)

	return &protos.RefreshResponse{}, nil
}

func (bp *BaremetalProvider) getNodeGroup(id string) cloudprovider.NodeGroup {
	for _, ng := range bp.nodeGroups {
		if ng.Id() == id {
			return ng
		}
	}

	return nil
}

func (bp *BaremetalProvider) NodeGroupTargetSize(_ context.Context, req *protos.NodeGroupTargetSizeRequest,
) (*protos.NodeGroupTargetSizeResponse, error) {
	debug(req)

	id := req.GetId()
	ng := bp.getNodeGroup(id)
	if ng == nil {
		return nil, fmt.Errorf("NodeGroup %q, not found", id)
	}

	size, err := ng.TargetSize()
	if err != nil {
		return nil, err
	}
	return &protos.NodeGroupTargetSizeResponse{
		TargetSize: int32(size),
	}, nil
}

func (bp *BaremetalProvider) NodeGroupIncreaseSize(_ context.Context, req *protos.NodeGroupIncreaseSizeRequest,
) (*protos.NodeGroupIncreaseSizeResponse, error) {
	debug(req)

	id := req.GetId()
	ng := bp.getNodeGroup(id)
	if ng == nil {
		return nil, fmt.Errorf("NodeGroup %q, not found", id)
	}
	err := ng.IncreaseSize(int(req.GetDelta()))
	if err != nil {
		return nil, err
	}
	return &protos.NodeGroupIncreaseSizeResponse{}, nil
}

func (bp *BaremetalProvider) NodeGroupDeleteNodes(_ context.Context, req *protos.NodeGroupDeleteNodesRequest,
) (*protos.NodeGroupDeleteNodesResponse, error) {
	debug(req)

	id := req.GetId()
	ng := bp.getNodeGroup(id)
	if ng == nil {
		return nil, fmt.Errorf("NodeGroup %q, not found", id)
	}
	klog.V(5).Info(req.GetNodes())
	nodes := make([]*apiv1.Node, 0)
	for _, n := range req.GetNodes() {
		nodes = append(nodes, apiv1Node(n))
	}
	err := ng.DeleteNodes(nodes)
	if err != nil {
		return nil, err
	}
	return &protos.NodeGroupDeleteNodesResponse{}, nil
}

func (bp *BaremetalProvider) NodeGroupDecreaseTargetSize(_ context.Context,
	req *protos.NodeGroupDecreaseTargetSizeRequest) (*protos.NodeGroupDecreaseTargetSizeResponse, error) {
	debug(req)

	id := req.GetId()
	ng := bp.getNodeGroup(id)
	if ng == nil {
		return nil, fmt.Errorf("NodeGroup %q, not found", id)
	}
	err := ng.DecreaseTargetSize(int(req.GetDelta()))
	if err != nil {
		return nil, err
	}
	return &protos.NodeGroupDecreaseTargetSizeResponse{}, nil
}

func (bp *BaremetalProvider) NodeGroupNodes(_ context.Context, req *protos.NodeGroupNodesRequest,
) (*protos.NodeGroupNodesResponse, error) {
	debug(req)

	id := req.GetId()
	ng := bp.getNodeGroup(id)
	if ng == nil {
		return nil, fmt.Errorf("NodeGroup %q, not found", id)
	}
	instances, err := ng.Nodes()
	if err != nil {
		return nil, err
	}
	pbInstances := make([]*protos.Instance, 0)
	for _, i := range instances {
		pbInstance := new(protos.Instance)
		pbInstance.Id = i.Id
		if i.Status == nil {
			pbInstance.Status = &protos.InstanceStatus{
				InstanceState: protos.InstanceStatus_unspecified,
				ErrorInfo:     &protos.InstanceErrorInfo{},
			}
		} else {
			pbInstance.Status = new(protos.InstanceStatus)
			pbInstance.Status.InstanceState = protos.InstanceStatus_InstanceState(i.Status.State)
			if i.Status.ErrorInfo == nil {
				pbInstance.Status.ErrorInfo = &protos.InstanceErrorInfo{}
			} else {
				pbInstance.Status.ErrorInfo = &protos.InstanceErrorInfo{
					ErrorCode:          i.Status.ErrorInfo.ErrorCode,
					ErrorMessage:       i.Status.ErrorInfo.ErrorMessage,
					InstanceErrorClass: int32(i.Status.ErrorInfo.ErrorClass),
				}
			}
		}
		pbInstances = append(pbInstances, pbInstance)
	}
	return &protos.NodeGroupNodesResponse{
		Instances: pbInstances,
	}, nil
}
