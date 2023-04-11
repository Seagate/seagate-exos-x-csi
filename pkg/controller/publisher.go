package controller

import (
	"context"

	"github.com/Seagate/seagate-exos-x-csi/pkg/common"
	"github.com/Seagate/seagate-exos-x-csi/pkg/node_service"
	pb "github.com/Seagate/seagate-exos-x-csi/pkg/node_service/node_servicepb"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

// ControllerPublishVolume attaches the given volume to the node
func (driver *Controller) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "cannot publish volume with empty ID")
	}
	if len(req.GetNodeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "cannot publish volume to a node with empty ID")
	}
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "cannot publish volume without capabilities")
	}

	nodeIP := req.GetNodeId()
	parameters := req.GetVolumeContext()

	var reqType pb.InitiatorType
	switch parameters[common.StorageProtocolKey] {
	case common.StorageProtocolSAS:
		reqType = pb.InitiatorType_SAS
	case common.StorageProtocolFC:
		reqType = pb.InitiatorType_FC
	case common.StorageProtocolISCSI:
		reqType = pb.InitiatorType_ISCSI
	}

	initiators, err := node_service.GetNodeInitiators(nodeIP, reqType)
	if err != nil {
		klog.ErrorS(err, "error getting node initiators", "node-ip", nodeIP, "storage-protocol", reqType)
		return nil, err
	}

	volumeName, _ := common.VolumeIdGetName(req.GetVolumeId())

	klog.Infof("attach request for initiator(s) %v, volume id: %s", initiators, volumeName)

	lun, err := driver.client.PublishVolume(volumeName, initiators)

	if err != nil {
		return nil, err
	}

	return &csi.ControllerPublishVolumeResponse{
		PublishContext: map[string]string{"lun": lun},
	}, err
}

// ControllerUnpublishVolume detaches the given volume from the node
func (driver *Controller) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "cannot unpublish volume with empty ID")
	}

	volumeName, _ := common.VolumeIdGetName(req.GetVolumeId())

	var initiators []string

	nodeIP := req.GetNodeId()
	storageProtocol, err := common.VolumeIdGetStorageProtocol(req.GetVolumeId())
	if err != nil {
		klog.ErrorS(err, "No storage protocol found in ControllerUnpublishVolume", "storage protocol", storageProtocol, "volume ID:", req.GetVolumeId())
		return nil, err
	}

	var reqType pb.InitiatorType
	switch storageProtocol {
	case common.StorageProtocolSAS:
		reqType = pb.InitiatorType_SAS
	case common.StorageProtocolFC:
		reqType = pb.InitiatorType_FC
	case common.StorageProtocolISCSI:
		reqType = pb.InitiatorType_ISCSI
	}

	initiators, err = node_service.GetNodeInitiators(nodeIP, reqType)
	if err != nil {
		klog.ErrorS(err, "error getting initiators from the node", "nodeIP", nodeIP, "storage-protocol", reqType)
	}

	klog.InfoS("unmapping volume from initiator", "volumeName", volumeName, "initiators", initiators)
	for _, initiator := range initiators {
		_, status, err := driver.client.UnmapVolume(volumeName, initiator)
		if err != nil {
			if status != nil && status.ReturnCode == unmapFailedErrorCode {
				klog.Info("unmap failed, assuming volume is already unmapped")
			} else {
				klog.Errorf("unknown error while unmapping initiator %s: %v", initiator, err)
			}
		}
	}

	klog.Infof("successfully unmapped volume %s from all initiators", volumeName)
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}
