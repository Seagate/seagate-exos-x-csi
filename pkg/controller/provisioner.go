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

var (
	volumeCaps = []csi.VolumeCapability_AccessMode{
		{
			Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		},
		{
			Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY,
		},
	}
)

// CreateVolume creates a new volume from the given request. The function is idempotent.
func (controller *Controller) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {

	parameters := req.GetParameters()
	volumeID, err := common.TranslateName(req.GetName(), parameters[common.VolumePrefixKey])
	if err != nil {
		return nil, err
	}

	if common.ValidateName(volumeID) == false {
		return nil, status.Error(codes.InvalidArgument, "volume name contains invalid characters")
	}

	volumeCapabilities := req.GetVolumeCapabilities()
	if err := isValidVolumeCapabilities(volumeCapabilities); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("CreateVolume Volume capabilities not valid: %v", err))
	}

	size := req.GetCapacityRange().GetRequiredBytes()
	sizeStr := getSizeStr(size)
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
			_, apistatus, err2 := controller.client.CopyVolume(sourceID, volumeID, parameters[common.PoolConfigKey])
			if err2 != nil {
				klog.Infof("-- CopyVolume apistatus.ReturnCode %v", apistatus.ReturnCode)
				if apistatus != nil && apistatus.ReturnCode == snapshotNotFoundErrorCode {
					return nil, status.Errorf(codes.NotFound, "Snapshot source (%s) not found", sourceID)
				} else {
					return nil, err2
				}
			}

		} else {
			_, _, err2 := controller.client.CreateVolume(volumeID, sizeStr, parameters[common.PoolConfigKey], poolType)
			if err2 != nil {
				return nil, err
			}
		}
	}

	// Fill iSCSI context parameters
	targetid, _ := controller.client.Info.GetTargetId()
	req.GetParameters()["iqn"] = targetid
	portals, _ := controller.client.Info.GetPortals()
	req.GetParameters()["portals"] = portals

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

// isValidVolumeCapabilities validates the given VolumeCapability array is valid
func isValidVolumeCapabilities(volCaps []*csi.VolumeCapability) error {
	if len(volCaps) == 0 {
		return fmt.Errorf("CreateVolume Volume capabilities must be provided")
	}
	hasSupport := func(cap *csi.VolumeCapability) error {
		if blk := cap.GetBlock(); blk != nil {
			return fmt.Errorf("driver only supports mount access type volume capability")
		}
		for _, c := range volumeCaps {
			if c.GetMode() == cap.AccessMode.GetMode() {
				return nil
			}
		}
		return fmt.Errorf("driver does not support access mode %v", cap.AccessMode.GetMode())
	}

	for _, c := range volCaps {
		if err := hasSupport(c); err != nil {
			return err
		}
	}
	return nil
}
