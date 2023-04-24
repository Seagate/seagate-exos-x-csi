package controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/Seagate/seagate-exos-x-csi/pkg/common"
	"github.com/Seagate/seagate-exos-x-csi/pkg/storage"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
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

// Extract available SAS addresses for Nodes from topology segments
// This will contain all SAS initiators for all nodes unless the storage class
// has specified allowed or preferred topologies
func parseTopology(topologies []*csi.Topology, storageProtocol string, parameters *map[string]string) ([]*csi.Topology, error) {
	klog.V(5).Infof("parseTopology: %v", topologies)

	accessibleTopology := []*csi.Topology{}
	hasInitiators := false
	for _, topo := range topologies {

		segments := topo.GetSegments()

		nodeID := segments[common.TopologyNodeIDKey]
		hasInitiators = false
		for key, val := range segments {
			if strings.Contains(key, common.TopologySASInitiatorLabel) || strings.Contains(key, common.TopologyFCInitiatorLabel) {
				hasInitiators = true
				newKey := strings.TrimPrefix(key, common.TopologyInitiatorPrefix)
				// insert the node ID into the key so we can retrieve the node specific addresses after scheduling by the CO
				newKey = nodeID + newKey
				(*parameters)[newKey] = val
			}
		}
		if hasInitiators {
			accessibleTopology = append(accessibleTopology, topo)
		}

	}
	if len(accessibleTopology) == 0 {
		accessibleTopology = nil
	}
	return accessibleTopology, nil
}

// CreateVolume creates a new volume from the given request. The function is idempotent.
func (controller *Controller) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {

	parameters := req.GetParameters()

	volumeName, err := common.TranslateName(req.GetName(), parameters[common.VolumePrefixKey])
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "translate volume name contains invalid characters")
	}

	// Extract the storage interface protocol to be used for this volume (iscsi, fc, sas, etc)
	storageProtocol := storage.ValidateStorageProtocol(parameters[common.StorageProtocolKey])

	if !common.ValidateName(volumeName) {
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
	wwn := ""

	if len(poolType) == 0 {
		poolType = "Virtual"
	}

	klog.Infof("creating volume %q (size %s) pool %q [%s] using protocol (%s)", volumeName, sizeStr, pool, poolType, storageProtocol)

	volumeExists, err := controller.client.CheckVolumeExists(volumeName, size)
	if err != nil {
		return nil, err
	}

	if !volumeExists {
		var sourceId string

		if volume := req.VolumeContentSource.GetVolume(); volume != nil {
			sourceId = volume.VolumeId
			klog.Infof("-- GetVolume sourceID %q", sourceId)
		}

		if snapshot := req.VolumeContentSource.GetSnapshot(); sourceId == "" && snapshot != nil {
			sourceId = snapshot.SnapshotId
			klog.Infof("-- GetSnapshot sourceID %q", sourceId)
		}

		if sourceId != "" {
			sourceName, err := common.VolumeIdGetName(sourceId)
			if err != nil {
				return nil, err
			}
			_, apistatus, err2 := controller.client.CopyVolume(sourceName, volumeName, parameters[common.PoolConfigKey])
			if err2 != nil {
				klog.Infof("-- CopyVolume apistatus.ReturnCode %v", apistatus.ReturnCode)
				if apistatus != nil && apistatus.ReturnCode == snapshotNotFoundErrorCode {
					return nil, status.Errorf(codes.NotFound, "Snapshot source (%s) not found", sourceId)
				} else {
					return nil, err2
				}
			}

		} else {
			_, _, err2 := controller.client.CreateVolume(volumeName, sizeStr, parameters[common.PoolConfigKey], poolType)
			if err2 != nil {
				return nil, err
			}
		}
	}

	if storageProtocol == common.StorageProtocolISCSI {
		// Fill iSCSI context parameters
		targetId, err1 := controller.client.Info.GetTargetId("iSCSI")
		if err1 != nil {
			klog.Errorf("++ GetTargetId error: %v", err1)
		}
		req.GetParameters()["iqn"] = targetId
		portals, err2 := controller.client.Info.GetPortals()
		if err2 != nil {
			klog.Errorf("++ GetPortals error: %v", err2)
		}
		req.GetParameters()["portals"] = portals
		klog.V(2).Infof("Storing iSCSI iqn: %s, portals: %v", targetId, portals)
	}

	wwn, _ = controller.client.GetVolumeWwn(volumeName)
	volumeId := common.VolumeIdAugment(volumeName, storageProtocol, wwn)

	volume := &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volumeId,
			VolumeContext: parameters,
			CapacityBytes: req.GetCapacityRange().GetRequiredBytes(),
			ContentSource: req.GetVolumeContentSource(),
		},
	}

	klog.Infof("created volume %s (%s)", volumeId, sizeStr)

	// Log struct with field names
	klog.V(8).Infof("created volume %+v", volume)
	return volume, nil
}

// DeleteVolume deletes the given volume. The function is idempotent.
func (controller *Controller) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "cannot delete volume with empty ID")
	}
	volumeName, _ := common.VolumeIdGetName(req.GetVolumeId())
	klog.Infof("deleting volume %s", volumeName)

	_, respStatus, err := controller.client.DeleteVolume(volumeName)
	if err != nil {
		if respStatus != nil {
			if respStatus.ReturnCode == volumeNotFoundErrorCode {
				klog.Infof("volume %s does not exist, assuming it has already been deleted", volumeName)
				return &csi.DeleteVolumeResponse{}, nil
			} else if respStatus.ReturnCode == volumeHasSnapshot {
				return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("volume %s cannot be deleted since it has snapshots", volumeName))
			}
		}
		return nil, err
	}

	klog.Infof("successfully deleted volume %s", volumeName)
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
