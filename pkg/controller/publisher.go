package controller

import (
	"context"
	"strconv"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
)

// ControllerPublishVolume attaches the given volume to the node
func (driver *Controller) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {

	if klog.V(8) {
		klog.Infof("context: %v", ctx)
		klog.Infof("csi.ControllerPublishVolumeRequest: %v", req)
	}

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "cannot publish volume with empty ID")
	}
	if len(req.GetNodeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "cannot publish volume to a node with empty ID")
	}
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "cannot publish volume without capabilities")
	}

	initiatorName := req.GetNodeId()
	klog.Infof("attach request for initiator %s, volume id: %s", initiatorName, req.GetVolumeId())

	maps, luns, _, err := driver.client.GetVolumeMaps(req.GetVolumeId())
	klog.Infof("maps = %v", maps)

	if err != nil {
		return nil, err
	}
	for i, hostName := range maps {
		if hostName == initiatorName {
			klog.Infof("volume %s is already mapped to (%s) [%s]", req.GetVolumeId(), initiatorName, luns[i])
			return &csi.ControllerPublishVolumeResponse{
				PublishContext: map[string]string{"lun": luns[i]},
			}, nil
		}
	}

	lun, err := driver.client.ChooseLUN(initiatorName)
	if err != nil {
		return nil, err
	}
	klog.Infof("using LUN %d", lun)

	if err = driver.client.MapVolumeProcess(req.GetVolumeId(), initiatorName, lun); err != nil {
		return nil, err
	}

	klog.Infof("successfully mapped volume %s for initiator %s", req.GetVolumeId(), initiatorName)
	return &csi.ControllerPublishVolumeResponse{
		PublishContext: map[string]string{"lun": strconv.Itoa(lun)},
	}, nil
}

// ControllerUnpublishVolume deattaches the given volume from the node
func (driver *Controller) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "cannot unpublish volume with empty ID")
	}

	klog.Infof("unmapping volume %s from initiator %s", req.GetVolumeId(), req.GetNodeId())
	_, status, err := driver.client.UnmapVolume(req.GetVolumeId(), req.GetNodeId())
	if err != nil {
		if status != nil && status.ReturnCode == unmapFailedErrorCode {
			klog.Info("unmap failed, assuming volume is already unmapped")
			return &csi.ControllerUnpublishVolumeResponse{}, nil
		}

		return nil, err
	}

	klog.Infof("successfully unmapped volume %s from all initiators", req.GetVolumeId())
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}
