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
	volume := req.GetVolumeId()
	klog.Infof("attach request for initiator %s, volume id: %s", initiatorName, volume)

	vmaps, _, err := driver.client.GetVolumeMaps2(volume)

	if err != nil {
		return nil, err
	}

	for _, vm := range vmaps.Mappings {
		if vm.InitiatorId != initiatorName {
			klog.Infof("volume (%s) is already attached to another initiator (%s)", volume, vm.InitiatorId)
			return nil, status.Errorf(codes.FailedPrecondition, "volume (%s) is already attached to another initiator (%s)", volume, vm.InitiatorId)
		}
	}

	imaps, _, err := driver.client.GetInitiatorMaps(initiatorName)

	// for _, im := range imaps.Mappings {
	// 	if im.Volume == volume {
	// 		klog.Infof("volume (%s) is already mapped to initiator (%s) using LUN %s", volume, initiatorName, im.LUN)
	// 		publishInfo := map[string]string{"lun": im.LUN}
	// 		return &csi.ControllerPublishVolumeResponse{PublishContext: publishInfo}, nil
	// 	}
	// }

	var lun int
	lun, err = driver.client.NextLUN(imaps)
	if err != nil {
		return nil, err
	}
	klog.Infof("using LUN %d", lun)

	if err = driver.client.MapVolumeProcess(volume, initiatorName, lun); err != nil {
		return nil, err
	}

	klog.Infof("successfully mapped volume %s for initiator %s using lun (%d)", req.GetVolumeId(), initiatorName, lun)

	// Build CSI controller publish info from volume publish info
	publishInfo := map[string]string{
		"lun": strconv.Itoa(lun),
	}

	response := csi.ControllerPublishVolumeResponse{PublishContext: publishInfo}
	klog.Infof("response: %v", response)

	return &response, nil
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
