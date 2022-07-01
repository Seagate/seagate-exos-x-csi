//
// Copyright (c) 2021 Seagate Technology LLC and/or its Affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// For any questions about this software or licensing,
// please email opensource@seagate.com or cortx-questions@seagate.com.

package storage

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	saslib "github.com/Seagate/csi-lib-sas/sas"
	"github.com/Seagate/seagate-exos-x-csi/pkg/common"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
)

// NodeStageVolume mounts the volume to a staging path on the node. This is
// called by the CO before NodePublishVolume and is used to temporary mount the
// volume to a staging path. Once mounted, NodePublishVolume will make sure to
// mount it to the appropriate path
// Will not be called as the plugin does not have the STAGE_UNSTAGE_VOLUME capability
func (sas *sasStorage) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeStageVolume is not implemented")
}

// NodeUnstageVolume unstages the volume from the staging path
// Will not be called as the plugin does not have the STAGE_UNSTAGE_VOLUME capability
func (sas *sasStorage) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeUnstageVolume is not implemented")
}

// NodePublishVolume mounts the volume mounted to the staging path to the target path
func (sas *sasStorage) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "cannot publish volume with empty id")
	}
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "cannot publish volume at an empty path")
	}
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "cannot publish volume without capabilities")
	}

	volumeName, _ := common.VolumeIdGetName(req.GetVolumeId())
	wwn, _ := common.VolumeIdGetWwn(req.GetVolumeId())
	lun, _ := req.GetPublishContext()["lun"]

	// Ensure that NodePublishVolume is only called once per volume
	addGatekeeper(volumeName)
	defer removeGatekeeper(volumeName)

	klog.V(1).Infof("[START] publish volume (%s) wwn (%s) target (%s) lun (%s)", volumeName, wwn, req.GetTargetPath(), lun)

	// Initiate SAS attachment
	klog.Info("initiating SAS connection...")
	connector := saslib.Connector{Lun: lun, TargetWWNs: []string{wwn}}
	path, err := saslib.Attach(ctx, connector, &saslib.OSioHandler{})
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	klog.Infof("attached device at %s", path)

	fsType := req.GetVolumeContext()[common.FsTypeConfigKey]
	err = ensureFsType(fsType, path)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	corrupted := false
	if err = checkFs(path, "Publish"); err != nil {
		corrupted = true
	}

	if connector.Multipath {
		klog.Infof("device is using multipath, device=%v, wwn=%v, corrupted=%v", path, wwn, corrupted)
	} else {
		klog.Infof("device is NOT using multipath, device=%v, wwn=%v, corrupted=%v", path, wwn, corrupted)
	}

	if corrupted {
		klog.Infof("device corruption (publish), device=%v, volume=%s, multipath=%v, wwn=%v, corrupted=%v", connector.DevicePath, volumeName, connector.Multipath, wwn, corrupted)
		debugCorruption("$$", path)
		return nil, status.Errorf(codes.DataLoss, "(publish) filesystem (%v) seems to be corrupted: %v", path, err)
	}

	out, err := exec.Command("findmnt", "--output", "TARGET", "--noheadings", path).Output()
	mountpoints := strings.Split(strings.Trim(string(out), "\n"), "\n")
	if err != nil || len(mountpoints) == 0 {
		klog.V(1).Infof("mount -t %s %s %s", fsType, path, req.GetTargetPath())
		os.Mkdir(req.GetTargetPath(), 00755)
		if _, err = os.Stat(path); errors.Is(err, os.ErrNotExist) {
			klog.Infof("targetpath does not exist:%s", req.GetTargetPath())
		}
		out, err = exec.Command("mount", "-t", fsType, path, req.GetTargetPath()).CombinedOutput()
		if err != nil {
			return nil, status.Error(codes.Internal, string(out))
		}
	} else if len(mountpoints) == 1 {
		if mountpoints[0] == req.GetTargetPath() {
			klog.Infof("volume %s already mounted", req.GetTargetPath())
		} else {
			errStr := fmt.Sprintf("device has already been mounted somewhere else (%s instead of %s), please unmount first", mountpoints[0], req.GetTargetPath())
			return nil, status.Error(codes.Internal, errStr)
		}
	} else if len(mountpoints) > 1 {
		return nil, errors.New("device has already been mounted in several locations, please unmount first")
	}

	klog.Infof("saving SAS connection info in %s", sas.connectorInfoPath)
	if _, err := os.Stat(sas.connectorInfoPath); err == nil {
		klog.Warningf("sas connection file already exists: %s", sas.connectorInfoPath)
	}
	err = saslib.PersistConnector(ctx, &connector, sas.connectorInfoPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	klog.Infof("successfully mounted volume at %s", req.GetTargetPath())
	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume unmounts the volume from the target path
func (sas *sasStorage) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeUnpublishVolume is not implemented")
}

// NodeGetVolumeStats return info about a given volume
// Will not be called as the plugin does not have the GET_VOLUME_STATS capability
func (sas *sasStorage) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeGetVolumeStats is not implemented")
}

// NodeExpandVolume finalizes volume expansion on the node
func (sas *sasStorage) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeExpandVolume is not implemented")
}

// NodeGetCapabilities returns the supported capabilities of the node server
func (sas *sasStorage) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeGetCapabilities is not implemented")
}

// NodeGetInfo returns info about the node
func (sas *sasStorage) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeGetInfo is not implemented")
}
