package controller

import (
	"context"

	"github.com/Seagate/seagate-exos-x-csi/pkg/common"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
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

	volumeName, _ := common.VolumeIdGetName(req.GetVolumeId())
	initiatorName := req.GetNodeId()
	klog.Infof("attach request for initiator %s, volume id: %s", initiatorName, volumeName)

	lun, err := driver.client.PublishVolume(volumeName, initiatorName)

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
	klog.Infof("unmapping volume %s from initiator %s", volumeName, req.GetNodeId())
	_, status, err := driver.client.UnmapVolume(volumeName, req.GetNodeId())
	if err != nil {
		if status != nil && status.ReturnCode == unmapFailedErrorCode {
			klog.Info("unmap failed, assuming volume is already unmapped")
			return &csi.ControllerUnpublishVolumeResponse{}, nil
		}

		return nil, err
	}

	klog.Infof("successfully unmapped volume %s from all initiators", volumeName)
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}
