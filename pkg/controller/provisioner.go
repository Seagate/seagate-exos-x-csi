package controller

import (
	"context"
	"fmt"

	"github.com/Seagate/seagate-exos-x-csi/pkg/common"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
)

// CreateVolume creates a new volume from the given request. The function is idempotent.
func (controller *Controller) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {

	volumeID, err := common.TranslateVolumeName(req)
	if err != nil {
		return nil, err
	}

	if common.ValidateVolumeName(volumeID) == false {
		return nil, status.Error(codes.InvalidArgument, "volume name contains invalid characters")
	}

	size := req.GetCapacityRange().GetRequiredBytes()
	sizeStr := getSizeStr(size)
	parameters := req.GetParameters()
	pool := parameters[common.PoolConfigKey]
	poolType, _ := controller.client.Info.GetPoolType(pool)

	if len(poolType) == 0 {
		poolType = "Virtual"
	}

	klog.Infof("creating volume %q (size %s) in pool %q [%s]", volumeID, sizeStr, pool, poolType)

	volumeExists, err := controller.client.CheckVolumeExists(volumeID, size)
	if err != nil {
		return nil, err
	}

	if !volumeExists {
		var sourceID string

		if volume := req.VolumeContentSource.GetVolume(); volume != nil {
			sourceID = volume.VolumeId
			klog.Infof("-- GetVolume sourceID %q", sourceID)
		}

		if snapshot := req.VolumeContentSource.GetSnapshot(); sourceID == "" && snapshot != nil {
			sourceID = snapshot.SnapshotId
			klog.Infof("-- GetSnapshot sourceID %q", sourceID)
		}

		if sourceID != "" {
			_, _, err = controller.client.CopyVolume(sourceID, volumeID, parameters[common.PoolConfigKey])
		} else {
			_, _, err = controller.client.CreateVolume(volumeID, sizeStr, parameters[common.PoolConfigKey], poolType)
		}
		if err != nil {
			return nil, err
		}
	}

	volume := &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volumeID,
			VolumeContext: parameters,
			CapacityBytes: req.GetCapacityRange().GetRequiredBytes(),
			ContentSource: req.GetVolumeContentSource(),
		},
	}

	klog.Infof("created volume %s (%s)", volumeID, sizeStr)
	// Log struct with field names
	klog.V(8).Infof("created volume %+v", volume)
	return volume, nil
}

// DeleteVolume deletes the given volume. The function is idempotent.
func (controller *Controller) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "cannot delete volume with empty ID")
	}

	klog.Infof("deleting volume %s", req.GetVolumeId())
	_, respStatus, err := controller.client.DeleteVolume(req.GetVolumeId())
	if err != nil {
		if respStatus != nil {
			if respStatus.ReturnCode == volumeNotFoundErrorCode {
				klog.Infof("volume %s does not exist, assuming it has already been deleted", req.GetVolumeId())
				return &csi.DeleteVolumeResponse{}, nil
			} else if respStatus.ReturnCode == volumeHasSnapshot {
				return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("volume %s cannot be deleted since it has snapshots", req.GetVolumeId()))
			}
		}
		return nil, err
	}

	klog.Infof("successfully deleted volume %s", req.GetVolumeId())
	return &csi.DeleteVolumeResponse{}, nil
}

func getSizeStr(size int64) string {
	if size == 0 {
		size = 4096
	}

	return fmt.Sprintf("%dB", size)
}
