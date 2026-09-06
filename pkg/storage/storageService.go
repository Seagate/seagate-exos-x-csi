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
	"path/filepath"
	"regexp"
	"strings"
	"time"

	saslib "github.com/Seagate/csi-lib-sas/sas"
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

var (
	execCommand        = exec.Command
	execCommandContext = exec.CommandContext
	readDir            = os.ReadDir
	readFile           = os.ReadFile
	evalSymlinks       = filepath.EvalSymlinks
)

const diskByIDPath = "/dev/disk/by-id"

func normalizeWWN(wwn string) string {
	wwn = strings.ToLower(strings.TrimSpace(wwn))
	wwn = strings.TrimPrefix(wwn, "mpath-")
	wwn = strings.TrimPrefix(wwn, "dm-name-")
	if strings.HasPrefix(wwn, "3") {
		wwn = strings.TrimPrefix(wwn, "3")
	}
	return wwn
}

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
var SASandFCRemovedDevicesMap = map[string]time.Time{}

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

// filesystemValidator returns the validation command used before trusting a detected filesystem.
func filesystemValidator(fsType string) string {
	if fsType == "xfs" {
		return "xfs_repair"
	}
	if strings.HasPrefix(fsType, "ext") {
		return "e2fsck"
	}
	return ""
}

func isValidFilesystem(fsType, devicePath string) bool {
	validator := filesystemValidator(fsType)
	if validator == "" {
		klog.Infof("No filesystem validator available for %s on device %s; treating detected filesystem as valid", fsType, devicePath)
		return true
	}

	klog.Infof("Validating detected %s filesystem on device %s with %s -n", fsType, devicePath, validator)
	out, err := execCommand(validator, "-n", devicePath).CombinedOutput()
	if err != nil {
		klog.Infof("Filesystem validation failed for %s on device %s: %s", fsType, devicePath, strings.TrimSpace(string(out)))
		return false
	}

	klog.Infof("Filesystem validation succeeded for %s on device %s", fsType, devicePath)
	return true
}

// Check for and remove any rediscovered iscsi devices that were previously unmapped
// This is a common function for SAS and FC
func CheckPreviouslyRemovedDevices(ctx context.Context) error {
	klog.Info("Checking previously removed devices")
	for wwn := range SASandFCRemovedDevicesMap {
		klog.Infof("Checking for rediscovery of wwn:%s", wwn)

		dm, devices := saslib.FindDiskById(klog.FromContext(ctx), wwn, &saslib.OSioHandler{})
		if dm != "" {
			klog.Infof("Rediscovery found for wwn:%s -- mpath device: %s, devices: %v", wwn, dm, devices)
			saslib.Detach(ctx, dm, &saslib.OSioHandler{})
		}
	}
	return nil
}

func resolveDeviceWWN(devicePath string) (string, error) {
	resolvedDevicePath, err := evalSymlinks(devicePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve device path %s: %w", devicePath, err)
	}

	dmName := filepath.Base(resolvedDevicePath)
	if strings.HasPrefix(dmName, "dm-") {
		sysfsUUIDPath := filepath.Join("/sys/block", dmName, "dm", "uuid")
		uuidBytes, err := readFile(sysfsUUIDPath)
		if err == nil {
			return normalizeWWN(string(uuidBytes)), nil
		}
	}

	entries, err := readDir(diskByIDPath)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", diskByIDPath, err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, "dm-name-3") {
			continue
		}

		byIDPath := filepath.Join(diskByIDPath, name)
		resolvedByIDPath, err := evalSymlinks(byIDPath)
		if err != nil {
			continue
		}

		if resolvedByIDPath == resolvedDevicePath {
			return normalizeWWN(name), nil
		}
	}

	return "", fmt.Errorf("failed to resolve WWN for device path %s", devicePath)
}

func ValidateAttachedDeviceWWN(volumeName, devicePath, expectedWWN string) error {
	resolvedWWN, err := resolveDeviceWWN(devicePath)
	if err != nil {
		klog.Errorf("Failed to resolve attached device WWN: volumeName=%s devicePath=%s expectedWWN=%s err=%v", volumeName, devicePath, expectedWWN, err)
		return err
	}

	expectedWWN = normalizeWWN(expectedWWN)
	klog.Infof("Attached device validation: volumeName=%s devicePath=%s resolvedWWN=%s expectedWWN=%s", volumeName, devicePath, resolvedWWN, expectedWWN)
	if resolvedWWN != expectedWWN {
		err := fmt.Errorf("attached device WWN mismatch for volume %s: expected %s, got %s for %s", volumeName, expectedWWN, resolvedWWN, devicePath)
		klog.Errorf("%v", err)
		return err
	}

	return nil
}

// FindDeviceFormat:
func FindDeviceFormat(device string) (string, error) {
	klog.V(2).Infof("Trying to find filesystem format on device %q", device)

	ctx, cancel := context.WithTimeout(context.Background(), BlkidTimeout*time.Second)
	defer cancel()
	output, err := execCommandContext(ctx, "blkid",
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

func resizeCommandForFs(fsType, devicePath, volumePath string) (string, []string, error) {
	switch {
	case fsType == "xfs":
		return "xfs_growfs", []string{volumePath}, nil
	case strings.HasPrefix(fsType, "ext"):
		return "resize2fs", []string{devicePath}, nil
	case fsType == "":
		return "", nil, fmt.Errorf("could not determine filesystem type for device %s", devicePath)
	default:
		return "", nil, fmt.Errorf("unsupported filesystem %q for resize on device %s", fsType, devicePath)
	}
}

func exitCodeFromError(err error) int {
	if err == nil {
		return 0
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode()
	}
	return -1
}

func runLoggedCommand(name string, args ...string) ([]byte, error) {
	klog.Infof("Running command: %s %s", name, strings.Join(args, " "))
	output, err := execCommand(name, args...).CombinedOutput()
	exitCode := exitCodeFromError(err)
	klog.Infof("Command output: %s", string(output))
	klog.Infof("Command exit code: %d", exitCode)
	if err != nil {
		klog.Errorf("Command failed: %s %s", name, strings.Join(args, " "))
		return output, fmt.Errorf("command %s %s failed with exit code %d: %s", name, strings.Join(args, " "), exitCode, string(output))
	}
	return output, nil
}

func ResizeFilesystem(devicePath, volumePath string) error {
	fsType, err := FindDeviceFormat(devicePath)
	if err != nil {
		klog.Errorf("Failed to detect filesystem for device %s: %v", devicePath, err)
		return err
	}
	klog.Infof("Detected filesystem: %s", fsType)
	klog.Infof("Resize target device path: %s", devicePath)
	klog.Infof("Resize target mount path: %s", volumePath)

	command, args, err := resizeCommandForFs(fsType, devicePath, volumePath)
	if err != nil {
		klog.Errorf("Failed to determine resize command for filesystem %s on device %s with mount path %s: %v", fsType, devicePath, volumePath, err)
		return err
	}

	dfBefore, err := runLoggedCommand("df", "-T", volumePath)
	if err != nil {
		klog.Errorf("Failed to collect filesystem state before resize for mount path %s: %v", volumePath, err)
		return fmt.Errorf("failed to collect filesystem state before resize for %s: %w", volumePath, err)
	}
	klog.Infof("Filesystem before resize: %s", string(dfBefore))

	klog.Infof("Expanding %s filesystem using %s", fsType, command)
	if _, err := runLoggedCommand(command, args...); err != nil {
		klog.Errorf("Filesystem resize failed for fsType=%s devicePath=%s volumePath=%s: %v", fsType, devicePath, volumePath, err)
		return fmt.Errorf("could not resize filesystem for fsType=%s devicePath=%s volumePath=%s: %w", fsType, devicePath, volumePath, err)
	}

	dfAfter, err := runLoggedCommand("df", "-T", volumePath)
	if err != nil {
		klog.Errorf("Failed to collect filesystem state after resize for mount path %s: %v", volumePath, err)
		return fmt.Errorf("failed to collect filesystem state after resize for %s: %w", volumePath, err)
	}
	klog.Infof("Filesystem after resize: %s", string(dfAfter))

	return nil
}

func mountTarget(fsType, devicePath, targetPath string) error {
	args := []string{"-t", fsType, devicePath, targetPath}
	klog.V(4).Infof("Running mount command: mount %s", strings.Join(args, " "))
	out, err := execCommand("mount", args...).CombinedOutput()
	if err == nil {
		return nil
	}
	klog.Errorf("Mount failed: cmd='mount %s' err='%v' output='%s'",
		strings.Join(args, " "),
		err,
		strings.TrimSpace(string(out)),
	)

	if fsType != "xfs" {
		return status.Error(codes.Internal, string(out))
	}

	// XFS preserves filesystem UUIDs across clones and snapshots. Linux refuses
	// to mount multiple XFS filesystems with the same UUID on the same node, so
	// retry with nouuid to allow cloned volumes to mount together.
	klog.Infof("xfs mount failed for %s, retrying with nouuid", devicePath)
	retryArgs := []string{"-t", fsType, "-o", "nouuid", devicePath, targetPath}
	klog.V(4).Infof("Running mount command: mount %s", strings.Join(retryArgs, " "))
	retryOut, retryErr := execCommand("mount", retryArgs...).CombinedOutput()
	if retryErr != nil {
		klog.Errorf("Mount retry failed: cmd='mount %s' err='%v' output='%s'",
			strings.Join(retryArgs, " "),
			retryErr,
			strings.TrimSpace(string(retryOut)),
		)
		errStr := fmt.Sprintf("%s; xfs retry with nouuid failed: %s", strings.TrimSpace(string(out)), strings.TrimSpace(string(retryOut)))
		return status.Error(codes.Internal, errStr)
	}

	klog.InfoS("successfully mounted xfs volume with nouuid", "devicePath", devicePath, "targetPath", targetPath)
	return nil
}

// EnsureFsType:
func EnsureFsType(fsType string, disk string) error {
	currentFsType, err := FindDeviceFormat(disk)
	if err != nil {
		return err
	}

	klog.Infof("Detected filesystem from blkid on device %s: %q", disk, currentFsType)
	switch {
	case currentFsType == "":
		klog.Infof("Device %s is unformatted; creating %s filesystem", disk, fsType)
	case currentFsType != fsType:
		return fmt.Errorf("Could not create %s filesystem on device %s since it already has one (%s)", fsType, disk, currentFsType)
	default:
		valid := isValidFilesystem(fsType, disk)
		klog.Infof("Filesystem validation result for %s on device %s: %t", fsType, disk, valid)
		if valid {
			return nil
		}
		klog.Infof("Detected %s filesystem on device %s is invalid; recreating filesystem", fsType, disk)
	}

	klog.Infof("Creating %s filesystem on device %s", fsType, disk)
	out, err := execCommand(fmt.Sprintf("mkfs.%s", fsType), disk).CombinedOutput()
	if err != nil {
		return errors.New(string(out))
	}

	return nil
}

func MountFilesystem(req *csi.NodePublishVolumeRequest, path string) error {
	fsType := GetFsType(req)
	err := EnsureFsType(fsType, path)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	out, err := execCommand("findmnt", "--output", "TARGET", "--noheadings", path).Output()
	mountpoints := strings.Split(strings.Trim(string(out), "\n"), "\n")
	if err != nil || len(mountpoints) == 0 {
		os.Mkdir(req.GetTargetPath(), 00755)
		if _, err = os.Stat(path); errors.Is(err, os.ErrNotExist) {
			klog.InfoS("targetpath does not exist", "targetPath", req.GetTargetPath())
		}
		if err = mountTarget(fsType, path, req.GetTargetPath()); err != nil {
			return err
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
