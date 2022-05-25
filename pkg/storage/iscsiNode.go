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
	"regexp"
	"strconv"
	"strings"
	"time"

	iscsilib "github.com/Seagate/csi-lib-iscsi/iscsi"
	"github.com/Seagate/seagate-exos-x-csi/pkg/common"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
)

// Configuration constants
const (
	BlkidTimeout      = 10
	maxDmnameAttempts = 10
	dmnameDelay       = 10
)

// NodeStageVolume mounts the volume to a staging path on the node. This is
// called by the CO before NodePublishVolume and is used to temporary mount the
// volume to a staging path. Once mounted, NodePublishVolume will make sure to
// mount it to the appropriate path
// Will not be called as the plugin does not have the STAGE_UNSTAGE_VOLUME capability
func (iscsi *iscsiStorage) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeStageVolume is not implemented")
}

// NodeUnstageVolume unstages the volume from the staging path
// Will not be called as the plugin does not have the STAGE_UNSTAGE_VOLUME capability
func (iscsi *iscsiStorage) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeUnstageVolume is not implemented")
}

// NodePublishVolume mounts the volume mounted to the staging path to the target path
func (iscsi *iscsiStorage) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
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
	klog.Infof("publishing volume %s, wwn (%s)", volumeName, wwn)
	klog.Infof("target path: %s", req.GetTargetPath())

	iqn := req.GetVolumeContext()["iqn"]
	portals := strings.Split(req.GetVolumeContext()["portals"], ",")
	klog.Infof("iSCSI iqn: %s, portals: %v", iqn, portals)

	lun, _ := strconv.ParseInt(req.GetPublishContext()["lun"], 10, 32)
	klog.Infof("lun-%d, LUN: %d", lun, lun)

	klog.Info("initiating ISCSI connection...")
	targets := make([]iscsilib.TargetInfo, 0)
	for _, portal := range portals {
		if portal != "" {
			klog.V(1).Infof("-- add iqn (%v) portal (%v)", iqn, portal)
			targets = append(targets, iscsilib.TargetInfo{
				Iqn:    iqn,
				Portal: portal,
			})
		}
	}
	connector := iscsilib.Connector{
		Targets:     targets,
		Lun:         int32(lun),
		DoDiscovery: true,
		RetryCount:  20,
	}

	path, err := iscsilib.Connect(&connector)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	klog.Infof("attached device at %s", path)

	exists := true
	out, err := exec.Command("ls", "-l", fmt.Sprintf("/dev/disk/by-id/dm-name-3%s", wwn)).CombinedOutput()
	klog.Infof("ls -l %s, err = %v, out = \n%s", fmt.Sprintf("/dev/disk/by-id/dm-name-3%s", wwn), err, string(out))
	if err != nil {
		exists = false
	}

	// wait here until the dm-name exists, for debugging
	if exists == false {
		// Force a reload of all existing multipath maps
		output, err := exec.Command("multipath", "-r").CombinedOutput()
		klog.Infof("## (publish) multipath -r: err=%v, output=\n%v", output, err)

		attempts := 1
		for attempts < (maxDmnameAttempts + 1) {
			out, err := exec.Command("ls", "-l", fmt.Sprintf("/dev/disk/by-id/dm-name-3%s", wwn)).CombinedOutput()
			klog.Infof("[%d] check for dm-name exists: ls -l %s, err = %v, out = \n%s", attempts, fmt.Sprintf("/dev/disk/by-id/dm-name-3%s", wwn), err, string(out))
			if err == nil {
				break
			}
			time.Sleep(dmnameDelay * time.Second)
			attempts++
		}
	}

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
		klog.Infof("device is using multipath, device=%v, wwn=%v, exists=%v, corrupted=%v", path, wwn, exists, corrupted)
	} else {
		klog.Infof("device is NOT using multipath, device=%v, wwn=%v, exists=%v, corrupted=%v", path, wwn, exists, corrupted)
	}

	if corrupted {
		klog.Infof("device corruption (publish), device=%v, multipath=%v, wwn=%v, exists=%v, corrupted=%v", connector.DevicePath, connector.Multipath, wwn, exists, corrupted)
		debugCorruption("$$", path)
		return nil, status.Errorf(codes.DataLoss, "(publish) filesystem (%v) seems to be corrupted: %v", path, err)
	}

	out, err = exec.Command("findmnt", "--output", "TARGET", "--noheadings", path).Output()
	mountpoints := strings.Split(strings.Trim(string(out), "\n"), "\n")
	if err != nil || len(mountpoints) == 0 {
		klog.Infof("mount -t %s %s %s", fsType, path, req.GetTargetPath())
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

	klog.Infof("saving ISCSI connection info in %s", iscsi.iscsiInfoPath)
	if _, err := os.Stat(iscsi.iscsiInfoPath); err == nil {
		klog.Warningf("@@ File Exists: %s", iscsi.iscsiInfoPath)
	}
	err = iscsilib.PersistConnector(&connector, iscsi.iscsiInfoPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	klog.Infof("successfully mounted volume at %s", req.GetTargetPath())
	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume unmounts the volume from the target path
func (iscsi *iscsiStorage) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "cannot unpublish volume with an empty volume id")
	}
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "cannot unpublish volume with an empty target path")
	}

	volumeName, _ := common.VolumeIdGetName(req.GetVolumeId())
	klog.Infof("unpublishing volume %s at target path %s", volumeName, req.GetTargetPath())

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

	klog.Infof("loading ISCSI connection info from %s", iscsi.iscsiInfoPath)
	connector, err := iscsilib.GetConnectorFromFile(iscsi.iscsiInfoPath)
	if err != nil {
		if os.IsNotExist(err) {
			klog.Warning(errors.Wrap(err, "assuming that ISCSI connection is already closed"))
			return &csi.NodeUnpublishVolumeResponse{}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	klog.Infof("connector.DevicePath (%s)", connector.DevicePath)

	if isVolumeInUse(connector.DevicePath) {
		klog.Info("volume is still in use on the node, thus it will not be detached")
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	_, err = os.Stat(connector.DevicePath)
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

	if err = checkFs(connector.DevicePath, "Unpublish"); err != nil {
		klog.Infof("device corruption (unpublish), device=%v, multipath=%v, wwn=%v, exists=%v, corrupted=%v", connector.DevicePath, connector.Multipath, wwn, exists, true)
		debugCorruption("!!", connector.DevicePath)
		return nil, status.Errorf(codes.DataLoss, "(unpublish) filesystem seems to be corrupted: %v", err)
	}

	klog.Info("DisconnectVolume, detaching ISCSI device")
	err = iscsilib.DisconnectVolume(*connector)
	if err != nil {
		return nil, err
	}

	klog.Infof("deleting ISCSI connection info file %s", iscsi.iscsiInfoPath)
	os.Remove(iscsi.iscsiInfoPath)

	klog.Info("successfully detached ISCSI device")
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeGetVolumeStats return info about a given volume
// Will not be called as the plugin does not have the GET_VOLUME_STATS capability
func (iscsi *iscsiStorage) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeGetVolumeStats is not implemented")
}

// NodeExpandVolume finalizes volume expansion on the node
func (iscsi *iscsiStorage) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {

	volumeName, _ := common.VolumeIdGetName(req.GetVolumeId())
	volumepath := req.GetVolumePath()
	klog.V(2).Infof("NodeExpandVolume: VolumeId=%v,  VolumePath=%v", volumeName, volumepath)

	if len(volumeName) == 0 {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("node expand volume requires volume id"))
	}

	if len(volumepath) == 0 {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("node expand volume requires volume path"))
	}

	connector, err := iscsilib.GetConnectorFromFile(iscsi.iscsiInfoPath)
	klog.V(3).Infof("GetConnectorFromFile(%s) connector: %v, err: %v", volumeName, connector, err)

	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("node expand volume path not found for volume id (%s)", volumeName))
	}

	// TODO: Is a rescan needed - rescan a scsi device by writing 1 in /sys/class/scsi_device/h:c:t:l/device/rescan
	// for i := range connector.Devices {
	// 	connector.Devices[i].Rescan()
	// }

	if connector.Multipath {
		klog.V(2).Info("device is using multipath")
		if err := iscsilib.ResizeMultipathDevice(connector.DevicePath); err != nil {
			return nil, err
		}
	} else {
		klog.V(2).Info("device is NOT using multipath")
	}

	klog.Infof("expanding filesystem using resize2fs on device %s", connector.DevicePath)
	output, err := exec.Command("resize2fs", connector.DevicePath).CombinedOutput()
	if err != nil {
		klog.V(2).Info("could not resize filesystem: %v", output)
		return nil, fmt.Errorf("could not resize filesystem: %v", output)
	}

	return &csi.NodeExpandVolumeResponse{}, nil
}

// NodeGetCapabilities returns the supported capabilities of the node server
func (iscsi *iscsiStorage) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeGetCapabilities is not implemented")
}

// NodeGetInfo returns info about the node
func (iscsi *iscsiStorage) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeGetInfo is not implemented")
}

// checkFs:
func checkFs(path string, context string) error {
	klog.Infof("Checking filesystem (e2fsck -n %s) [%s]", path, context)
	if out, err := exec.Command("e2fsck", "-n", path).CombinedOutput(); err != nil {
		return errors.New(string(out))
	}
	return nil
}

// findDeviceFormat:
func findDeviceFormat(device string) (string, error) {
	klog.V(2).Infof("Trying to find filesystem format on device %q", device)

	ctx, cancel := context.WithTimeout(context.Background(), BlkidTimeout*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, "blkid",
		"-p",
		"-s", "TYPE",
		"-s", "PTTYPE",
		"-o", "export",
		device).CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		err = fmt.Errorf("command timed out after %d seconds", BlkidTimeout)
	}

	klog.V(2).Infof("blkid output: %q, err=%v", output, err)

	if err != nil {
		// blkid exit with code 2 if the specified token (TYPE/PTTYPE, etc) could not be found or if device could not be identified.
		if exit, ok := err.(*exec.ExitError); ok && exit.ExitCode() == 2 {
			klog.V(2).Infof("Device seems to be is unformatted (%v)", err)
			return "", nil
		}
		return "", fmt.Errorf("could not not find format for device %q (%v)", device, err)
	}

	re := regexp.MustCompile(`([A-Z]+)="?([^"\n]+)"?`) // Handles alpine and debian outputs
	matches := re.FindAllSubmatch(output, -1)

	var filesystemType, partitionType string
	for _, match := range matches {
		if len(match) != 3 {
			return "", fmt.Errorf("invalid blkid output: %s", output)
		}
		key := string(match[1])
		value := string(match[2])

		if key == "TYPE" {
			filesystemType = value
		} else if key == "PTTYPE" {
			partitionType = value
		}
	}

	if partitionType != "" {
		klog.V(2).Infof("Device %q seems to have a partition table type: %s", partitionType)
		return "OTHER/PARTITIONS", nil
	}

	return filesystemType, nil
}

// ensureFsType:
func ensureFsType(fsType string, disk string) error {
	currentFsType, err := findDeviceFormat(disk)
	if err != nil {
		return err
	}

	klog.V(1).Infof("Detected filesystem: %q", currentFsType)
	if currentFsType != fsType {
		if currentFsType != "" {
			return fmt.Errorf("Could not create %s filesystem on device %s since it already has one (%s)", fsType, disk, currentFsType)
		}

		klog.Infof("Creating %s filesystem on device %s", fsType, disk)
		out, err := exec.Command(fmt.Sprintf("mkfs.%s", fsType), disk).CombinedOutput()
		if err != nil {
			return errors.New(string(out))
		}
	}

	return nil
}

// isVolumeInUse: Use findmnt to determine if the devie path is mounted or not.
func isVolumeInUse(devicePath string) bool {
	_, err := exec.Command("findmnt", devicePath).CombinedOutput()
	klog.Infof("isVolumeInUse: findmnt %s, err=%v", devicePath, err)
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return false
		}
	}
	return true
}

func debugCorruption(prefix, path string) {
	out, err := exec.Command("lsof", "-V", path).Output()
	klog.Infof("%s checkFs ERROR lsof %s, err = %v, out = \n%s", prefix, path, err, string(out))

	out, err = exec.Command("ls", "-l", path).CombinedOutput()
	klog.Infof("%s ls -l %s, err = %v, out = \n%s", prefix, path, err, string(out))

	out, err = exec.Command("multipath", "-ll", "-v2", path).CombinedOutput()
	klog.Infof("%s multipath -ll -v2 %s, err = %v, out = \n%s", prefix, path, err, string(out))

	out, err = exec.Command("ls", "-lR", "/dev/disk").CombinedOutput()
	klog.Infof("%s ls -lR /dev/disk, err = %v, out = \n%s", prefix, err, string(out))
}
