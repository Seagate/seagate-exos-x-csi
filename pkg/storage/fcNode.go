//
// Copyright (c) 2022 Seagate Technology LLC and/or its Affiliates
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

	fclib "github.com/Seagate/csi-lib-sas/sas"
	"github.com/Seagate/seagate-exos-x-csi/pkg/common"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

// NodeStageVolume mounts the volume to a staging path on the node. This is
// called by the CO before NodePublishVolume and is used to temporary mount the
// volume to a staging path. Once mounted, NodePublishVolume will make sure to
// mount it to the appropriate path
// Will not be called as the plugin does not have the STAGE_UNSTAGE_VOLUME capability
func (fc *fcStorage) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeStageVolume is not implemented")
}

// NodeUnstageVolume unstages the volume from the staging path
// Will not be called as the plugin does not have the STAGE_UNSTAGE_VOLUME capability
func (fc *fcStorage) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeUnstageVolume is not implemented")
}

// NodePublishVolume mounts the volume mounted to the staging path to the target path
func (fc *fcStorage) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
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
	AddGatekeeper(volumeName)
	defer RemoveGatekeeper(volumeName)

	klog.V(1).Infof("[START] publish volume (%s) wwn (%s) target (%s) lun (%s)", volumeName, wwn, req.GetTargetPath(), lun)

	// Initiate FC attachment
	klog.Info("initiating FC connection...")
	connector := fclib.Connector{VolumeWWN: wwn}
	path, err := fclib.Attach(ctx, &connector, &fclib.OSioHandler{})
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	klog.Infof("attached device at %s", path)

	fsType := req.GetVolumeContext()[common.FsTypeConfigKey]
	err = EnsureFsType(fsType, path)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	corrupted := false
	if err = CheckFs(path, "Publish"); err != nil {
		corrupted = true
	}

	if connector.Multipath {
		klog.Infof("device is using multipath, device=%v, wwn=%v, corrupted=%v", path, wwn, corrupted)
	} else {
		klog.Infof("device is NOT using multipath, device=%v, wwn=%v, corrupted=%v", path, wwn, corrupted)
	}

	if corrupted {
		klog.Infof("device corruption (publish), device=%v, volume=%s, multipath=%v, wwn=%v, corrupted=%v", connector.OSPathName, volumeName, connector.Multipath, wwn, corrupted)
		DebugCorruption("$$", path)
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

	klog.Infof("saving FC connection info in %s", fc.connectorInfoPath)
	if _, err := os.Stat(fc.connectorInfoPath); err == nil {
		klog.Warningf("fc connection file already exists: %s", fc.connectorInfoPath)
	}
	err = connector.Persist(ctx, fc.connectorInfoPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	klog.Infof("successfully mounted volume at %s", req.GetTargetPath())
	return &csi.NodePublishVolumeResponse{}, nil

}

// NodeUnpublishVolume unmounts the volume from the target path
func (fc *fcStorage) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "cannot unpublish volume with an empty volume id")
	}
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "cannot unpublish volume with an empty target path")
	}

	volumeName, _ := common.VolumeIdGetName(req.GetVolumeId())

	// Ensure that NodeUnpublishVolume is only called once per volume
	AddGatekeeper(volumeName)
	defer RemoveGatekeeper(volumeName)

	klog.Infof("[START] unpublishing volume (%s) at target path %s", volumeName, req.GetTargetPath())

	_, err := os.Stat(req.GetTargetPath())
	if err == nil {
		klog.Infof("unmounting volume at %s", req.GetTargetPath())
		klog.V(4).Infof("command: %s %s", "mountpoint", req.GetTargetPath())
		out, err := exec.Command("mountpoint", req.GetTargetPath()).CombinedOutput()
		if err == nil {
			klog.V(4).Infof("command: %s %s", "umount -l", req.GetTargetPath())
			out, err := exec.Command("umount", "-l", req.GetTargetPath()).CombinedOutput()
			if err != nil {
				return nil, status.Error(codes.Internal, string(out))
			}
		} else {
			klog.Warningf("assuming that volume is already unmounted: %s", out)
		}

		err = os.Remove(req.GetTargetPath())
		if err != nil && !os.IsNotExist(err) {
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		klog.Warningf("assuming that volume is already unmounted: %v", err)
	}

	klog.Infof("loading FC connection info from %s", fc.connectorInfoPath)
	connector, err := fclib.GetConnectorFromFile(fc.connectorInfoPath)
	if err != nil {
		if os.IsNotExist(err) {
			klog.Warning(errors.Wrap(err, "assuming that FC connection was already closed"))
			return &csi.NodeUnpublishVolumeResponse{}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	klog.Infof("connector.OSPathName (%s)", connector.OSPathName)

	if IsVolumeInUse(connector.OSPathName) {
		klog.Info("volume is still in use on the node, thus it will not be detached")
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	_, err = os.Stat(connector.OSPathName)
	if err != nil && os.IsNotExist(err) {
		klog.Warningf("assuming that volume is already disconnected: %s", err)
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	wwn, _ := common.VolumeIdGetWwn(req.GetVolumeId())
	exists := true
	out, err := exec.Command("ls", "-l", fmt.Sprintf("/dev/disk/by-id/dm-name-3%s", wwn)).CombinedOutput()
	klog.Infof("check for dm-name: ls -l %s, err = %v, out = \n%s", fmt.Sprintf("/dev/disk/by-id/dm-name-3%s", wwn), err, string(out))
	if err != nil {
		exists = false
	}

	if err = CheckFs(connector.OSPathName, "Unpublish"); err != nil {
		klog.Infof("device corruption (unpublish), device=%v, volume=%s, multipath=%v, wwn=%v, exists=%v, corrupted=%v", connector.OSPathName, volumeName, connector.Multipath, wwn, exists, true)
		DebugCorruption("!!", connector.OSPathName)
		return nil, status.Errorf(codes.DataLoss, "(unpublish) filesystem seems to be corrupted: %v", err)
	}

	klog.Info("DisconnectVolume, detaching device")
	err = fclib.Detach(ctx, connector.OSPathName, connector.IoHandler)

	if err != nil {
		return nil, err
	}

	klog.Infof("deleting FC connection info file %s", fc.connectorInfoPath)
	os.Remove(fc.connectorInfoPath)

	klog.Info("successfully detached FC device")
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeGetVolumeStats return info about a given volume
// Will not be called as the plugin does not have the GET_VOLUME_STATS capability
func (fc *fcStorage) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeGetVolumeStats is not implemented")
}

// NodeExpandVolume finalizes volume expansion on the node
func (fc *fcStorage) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {

	volumeName, _ := common.VolumeIdGetName(req.GetVolumeId())
	volumepath := req.GetVolumePath()
	klog.V(2).Infof("NodeExpandVolume: VolumeId=%v,  VolumePath=%v", volumeName, volumepath)

	if len(volumeName) == 0 {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("node expand volume requires volume id"))
	}

	if len(volumepath) == 0 {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("node expand volume requires volume path"))
	}

	connector, err := fclib.GetConnectorFromFile(fc.connectorInfoPath)
	klog.V(3).Infof("GetConnectorFromFile(%s) connector: %v, err: %v", volumeName, connector, err)

	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("node expand volume path not found for volume id (%s)", volumeName))
	}

	if connector.Multipath {
		klog.V(2).Info("device is using multipath")
		if err := fclib.ResizeMultipathDevice(ctx, connector.OSPathName); err != nil {
			return nil, err
		}
	} else {
		klog.V(2).Info("device is NOT using multipath")
	}

	klog.Infof("expanding filesystem using resize2fs on device %s", connector.OSPathName)
	output, err := exec.Command("resize2fs", connector.OSPathName).CombinedOutput()
	if err != nil {
		klog.V(2).Info("could not resize filesystem: %v", output)
		return nil, fmt.Errorf("could not resize filesystem: %v", output)
	}

	return &csi.NodeExpandVolumeResponse{}, nil
}

// NodeGetCapabilities returns the supported capabilities of the node server
func (fc *fcStorage) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeGetCapabilities is not implemented")
}

// NodeGetInfo returns info about the node
func (fc *fcStorage) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeGetInfo is not implemented")
}
