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
	"regexp"
	"strings"
	"time"

	"github.com/Seagate/seagate-exos-x-csi/pkg/common"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

const (
	SASAddressFilePath = "/etc/kubernetes/sas-addresses"
	FCAddressFilePath  = "/etc/kubernetes/fc-addresses"
)

type StorageOperations interface {
	csi.NodeServer
	AttachStorage(ctx context.Context, req *csi.NodePublishVolumeRequest) (string, error)
	DetachStorage(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) error
}

type commonService struct {
	storagePoolIdName map[int64]string
	driverVersion     string
}

type fcStorage struct {
	cs                commonService
	connectorInfoPath string
}

type iscsiStorage struct {
	cs                commonService
	connectorInfoPath string
}

type sasStorage struct {
	cs                commonService
	connectorInfoPath string
}

// Map of device WWNs to timestamp of when they were unpublished from the node
var GlobalRemovedDevicesMap = map[string]time.Time{}

// buildCommonService:
func buildCommonService(config map[string]string) (commonService, error) {
	commonserv := commonService{}
	commonserv.driverVersion = config["driverversion"]
	klog.V(2).Infof("buildCommonService commonservice configuration done.")
	return commonserv, nil
}

// NewStorageNode : To return specific implementation of storage
func NewStorageNode(storageProtocol string, config map[string]string) (StorageOperations, error) {
	comnserv, err := buildCommonService(config)
	if err == nil {
		storageProtocol = strings.TrimSpace(storageProtocol)
		klog.V(2).Infof("NewStorageNode for (%s)", storageProtocol)
		if storageProtocol == common.StorageProtocolFC {
			return &fcStorage{cs: comnserv, connectorInfoPath: config["connectorInfoPath"]}, nil
		} else if storageProtocol == common.StorageProtocolSAS {
			return &sasStorage{cs: comnserv, connectorInfoPath: config["connectorInfoPath"]}, nil
		} else if storageProtocol == common.StorageProtocolISCSI {
			return &iscsiStorage{cs: comnserv, connectorInfoPath: config["connectorInfoPath"]}, nil
		} else {
			klog.Warningf("Invalid or no storage protocol specified (%s)", storageProtocol)
			klog.Warningf("Expecting storageProtocol (iscsi, fc, sas, etc) in StorageClass YAML. Default of (%s) used.", common.StorageProtocolISCSI)
			return &iscsiStorage{cs: comnserv, connectorInfoPath: config["connectorInfoPath"]}, nil
		}
	}
	return nil, err
}

// ValidateStorageProtocol: Verifies that a correct protocol is chosen or returns a valid default storage protocol.
func ValidateStorageProtocol(storageProtocol string) string {
	if storageProtocol == common.StorageProtocolFC || storageProtocol == common.StorageProtocolISCSI || storageProtocol == common.StorageProtocolSAS {
		return storageProtocol
	} else {
		klog.Warningf("Invalid or no storage protocol specified (%s)", storageProtocol)
		klog.Warningf("Expecting storageProtocol (iscsi, fc, sas, etc) in StorageClass YAML. Default of (%s) used.", common.StorageProtocolISCSI)
		return common.StorageProtocolISCSI
	}
}

// gateKeepers is a thread safe map indexed by volume name.
var gatekeepers = common.NewStringLock()

// addGatekeeper: Ensure that NodePublishVolume and NodeUnpublishVolume are only called once per volume
func AddGatekeeper(volumeName string) {
	klog.V(4).Infof("[LOCK] volume (%s) gatekeeper", volumeName)
	gatekeepers.Lock(volumeName)
}

// removeGatekeeper: Unlock the volume function mutex when the Publish/Unpublish is complete
func RemoveGatekeeper(volumeName string) {
	klog.V(4).Infof("[UNLOCK] volume (%s) gatekeeper", volumeName)
	gatekeepers.Unlock(volumeName)
}

// wrap the new FS type specification and fall back to the old parameter if necessary
func GetFsType(req *csi.NodePublishVolumeRequest) string {
	fsType := ""
	if fsType = req.GetVolumeCapability().GetMount().GetFsType(); fsType == "" {
		fsType = req.GetVolumeContext()[common.FsTypeConfigKey]
	}
	return fsType
}

// CheckFs: Perform a file system validation
func CheckFs(path string, fstype string, context string) error {

	if IsVolumeInUse(path) {
		klog.Infof("Volume already mounted, not performing FS check")
		return nil
	}

	fsRepairCommand := "e2fsck"
	if fstype == "xfs" {
		fsRepairCommand = "xfs_repair"
	}
	klog.Infof("Checking filesystem (%s -n %s) [%s]", fsRepairCommand, path, context)
	if out, err := exec.Command(fsRepairCommand, "-n", path).CombinedOutput(); err != nil {
		return errors.New(string(out))
	}
	return nil
}

// FindDeviceFormat:
func FindDeviceFormat(device string) (string, error) {
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

// EnsureFsType:
func EnsureFsType(fsType string, disk string) error {
	currentFsType, err := FindDeviceFormat(disk)
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

func MountFilesystem(req *csi.NodePublishVolumeRequest, path string) error {
	fsType := GetFsType(req)
	err := EnsureFsType(fsType, path)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	if err = CheckFs(path, fsType, "Publish"); err != nil {
		return err
	}

	out, err := exec.Command("findmnt", "--output", "TARGET", "--noheadings", path).Output()
	mountpoints := strings.Split(strings.Trim(string(out), "\n"), "\n")
	if err != nil || len(mountpoints) == 0 {
		klog.V(1).InfoS("mount", "command", fmt.Sprintf("mount -t %s %s %s", fsType, path, req.GetTargetPath()))
		os.Mkdir(req.GetTargetPath(), 00755)
		if _, err = os.Stat(path); errors.Is(err, os.ErrNotExist) {
			klog.InfoS("targetpath does not exist", "targetPath", req.GetTargetPath())
		}
		out, err = exec.Command("mount", "-t", fsType, path, req.GetTargetPath()).CombinedOutput()
		if err != nil {
			return status.Error(codes.Internal, string(out))
		}
	} else if len(mountpoints) == 1 {
		if mountpoints[0] == req.GetTargetPath() {
			klog.InfoS("volume already mounted", "targetPath", req.GetTargetPath())
		} else {
			errStr := fmt.Sprintf("device has already been mounted somewhere else (%s instead of %s), please unmount first", mountpoints[0], req.GetTargetPath())
			return status.Error(codes.Internal, errStr)
		}
	} else if len(mountpoints) > 1 {
		return errors.New("device has already been mounted in several locations, please unmount first")
	}

	klog.InfoS("successfully mounted volume", "targetPath", req.GetTargetPath())
	return nil
}

func MountDevice(req *csi.NodePublishVolumeRequest, path string) error {
	deviceFile, err := os.Create(req.GetTargetPath())
	if err != nil {
		klog.ErrorS(err, "could not create file", "TargetPath", req.GetTargetPath())
		return err
	}
	deviceFile.Chmod(00755)
	deviceFile.Close()
	out, err := exec.Command("mount", "-o", "bind", path, req.GetTargetPath()).CombinedOutput()
	if err != nil {
		return status.Error(codes.Internal, string(out))
	}
	return nil
}

// Unmount a given path, usually req.GetVolumePath() from NodeUnpublishVolume
// used for both block and filesystem mount types
func Unmount(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		klog.InfoS("unmounting volume", "path", path)
		klog.V(4).InfoS("mountpoint command", "command", "mountpoint "+path)
		out, err := exec.Command("mountpoint", path).CombinedOutput()
		if err == nil {
			klog.V(4).InfoS("umount command", "command", "umount -l "+path)
			out, err := exec.Command("umount", "-l", path).CombinedOutput()
			if err != nil {
				return status.Error(codes.Internal, string(out))
			}
		} else {
			klog.ErrorS(err, "assuming that volume is already unmounted", "mountpoint_output", out)
		}

		err = os.Remove(path)
		if err != nil && !os.IsNotExist(err) {
			return status.Error(codes.Internal, err.Error())
		}
	} else {
		klog.ErrorS(err, "assuming that volume is already unmounted")
	}
	return nil
}

// IsVolumeInUse: Use findmnt to determine if the device path is mounted or not.
func IsVolumeInUse(devicePath string) bool {
	_, err := exec.Command("findmnt", devicePath).CombinedOutput()
	klog.Infof("isVolumeInUse: findmnt %s, err=%v", devicePath, err)
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return false
		}
	}
	return true
}

// DebugCorruption: Display additional information for debugging
func DebugCorruption(prefix, path string) {
	out, err := exec.Command("ls", "-l", path).CombinedOutput()
	klog.Infof("%s ls -l %s, err = %v, out = \n%s", prefix, path, err, string(out))

	out, err = exec.Command("multipath", "-ll", "-v2", path).CombinedOutput()
	klog.Infof("%s multipath -ll -v2 %s, err = %v, out = \n%s", prefix, path, err, string(out))

	out, err = exec.Command("ls", "-lR", "/dev/disk").CombinedOutput()
	klog.Infof("%s ls -lR /dev/disk, err = %v, out = \n%s", prefix, err, string(out))
}
