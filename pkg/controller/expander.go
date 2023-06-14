package controller

import (
	"context"
	"fmt"

	"github.com/Seagate/seagate-exos-x-csi/pkg/common"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

// ControllerExpandVolume expands a volume to the given new size
func (controller *Controller) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	volumeName, _ := common.VolumeIdGetName(req.GetVolumeId())
	if volumeName == "" {
		return nil, status.Error(codes.InvalidArgument, "cannot expand a volume with an empty ID")
	}
	klog.Infof("expanding volume %q", volumeName)

	newSize := req.GetCapacityRange().GetRequiredBytes()
	if newSize == 0 {
		newSize = req.GetCapacityRange().GetLimitBytes()
	}
	klog.V(2).Infof("requested size: %d bytes", newSize)

	response, _, err := controller.client.ShowVolumes(volumeName)
	var expansionSize int64
	if err != nil {
		return nil, err
	} else if len(response) == 0 {
		return nil, fmt.Errorf("volume %q not found", volumeName)
	} else if response[0].SizeNumeric == 0 {
		return nil, fmt.Errorf("could not get current volume size, thus volume expansion is not possible")
	} else if response[0].Blocks == 0 {
		return nil, fmt.Errorf("could not parse volume size: %v", err)
	} else {
		currentSize := response[0].Blocks * response[0].BlockSize
		klog.V(2).Infof("current size: %d bytes", currentSize)
		expansionSize = newSize - currentSize
		klog.V(2).Infof("expanding volume by %d bytes", expansionSize)
	}

	expansionSizeStr := getSizeStr(expansionSize)
	if _, err := controller.client.ExpandVolume(volumeName, expansionSizeStr); err != nil {
		return nil, err
	}

	klog.Infof("volume %q successfully expanded", volumeName)

	return &csi.ControllerExpandVolumeResponse{
		CapacityBytes:         newSize,
		NodeExpansionRequired: true,
	}, nil
}
