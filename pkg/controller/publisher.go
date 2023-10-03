package controller

import (
	"context"
	"fmt"

	storageapitypes "github.com/Seagate/seagate-exos-x-api-go/pkg/common"

	"github.com/Seagate/seagate-exos-x-csi/pkg/common"
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

	initiators, err := driver.GetNodeInitiators(ctx, nodeIP, parameters[common.StorageProtocolKey])
	if err != nil {
		klog.ErrorS(err, "error getting node initiators", "node-ip", nodeIP, "storage-protocol", parameters[common.StorageProtocolKey])
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Could not retrieve initiators for scheduled node(%s)", nodeIP))
	}

	volumeName, _ := common.VolumeIdGetName(req.GetVolumeId())

	klog.InfoS("attach request", "initiator(s)", initiators, "volume", volumeName)

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
	volumeWWN, _ := common.VolumeIdGetWwn(req.GetVolumeId())
	nodeIP := req.GetNodeId()
	storageProtocol, err := common.VolumeIdGetStorageProtocol(req.GetVolumeId())
	if err != nil {
		klog.ErrorS(err, "No storage protocol found in ControllerUnpublishVolume", "storage protocol", storageProtocol, "volume ID:", req.GetVolumeId())
		return nil, err
	}

	initiators, err := driver.GetNodeInitiators(ctx, nodeIP, storageProtocol)
	if err != nil {
		klog.ErrorS(err, "error getting initiators from the node", "nodeIP", nodeIP, "storageProtocol", storageProtocol)
	}

	klog.InfoS("unmapping volume from initiator", "volumeName", volumeName, "initiators", initiators)
	for _, initiator := range initiators {
		status, err := driver.client.UnmapVolume(volumeName, initiator)
		if err != nil {
			if status != nil && status.ReturnCode == storageapitypes.UnmapFailedErrorCode {
				klog.Info("unmap failed, assuming volume is already unmapped")
			} else {
				klog.Errorf("unknown error while unmapping initiator %s: %v", initiator, err)
			}
		} else {
			driver.NotifyUnmap(ctx, nodeIP, volumeWWN)
		}
	}

	klog.Infof("successfully unmapped volume %s from all initiators", volumeName)
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}
