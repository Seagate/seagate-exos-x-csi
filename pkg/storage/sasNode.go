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
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

// Read the sas address configuration file and return addresses
func readSASAddrFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	foundAddresses := []string{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		foundAddresses = append(foundAddresses, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return foundAddresses, nil
}

func GetSASInitiators() ([]string, error) {
	// look for file specifying addresses. Skip discovery process if it exists
	specifiedSASAddrs, err := readSASAddrFile(SASAddressFilePath)
	if err != nil {
		klog.ErrorS(err, "Error reading sas address config file", "path", SASAddressFilePath)
	}
	if specifiedSASAddrs != nil {
		return specifiedSASAddrs, nil
	}
	klog.InfoS("begin SAS address discovery")
	sasAddrFilename := "host_sas_address"
	scsiHostBasePath := "/sys/class/scsi_host/"

	dirList, err := os.ReadDir(scsiHostBasePath)
	if err != nil {
		return nil, err
	}

	for _, hostDir := range dirList {
		sasAddrFile := filepath.Join(scsiHostBasePath, hostDir.Name(), sasAddrFilename)
		addrBytes, err := os.ReadFile(sasAddrFile)
		address := string(addrBytes)
		address = strings.TrimLeft(strings.TrimRight(address, "\n"), "0x")

		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			} else {
				klog.ErrorS(err, "error searching for SAS HBA addresses", "path", sasAddrFile)
				return nil, err
			}
		} else {
			klog.InfoS("found SAS HBA address", "address", address)
			if firstAddress, err := strconv.ParseInt(address, 16, 0); err != nil {
				return nil, err
			} else {
				secondAddress := strconv.FormatInt(firstAddress+1, 16)
				specifiedSASAddrs = append(specifiedSASAddrs, address, secondAddress)
			}
		}
	}
	return specifiedSASAddrs, nil
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
	if req.GetVolumeCapability().GetBlock() != nil &&
		req.GetVolumeCapability().GetMount() != nil {
		return nil, status.Error(codes.InvalidArgument, "cannot have both block and mount access type")
	}
	if req.GetVolumeCapability().GetBlock() == nil &&
		req.GetVolumeCapability().GetMount() == nil {
		return nil, status.Error(codes.InvalidArgument, "volume access type not specified, must be either block or mount")
	}

	volumeName, _ := common.VolumeIdGetName(req.GetVolumeId())
	wwn, _ := common.VolumeIdGetWwn(req.GetVolumeId())
	lun := req.GetPublishContext()["lun"]

	// Ensure that NodePublishVolume is only called once per volume
	AddGatekeeper(volumeName)
	defer RemoveGatekeeper(volumeName)

	klog.V(1).InfoS("[START] publishing volume", "volumeName", volumeName, "wwn", wwn, "targetPath", req.GetTargetPath(), "lun", lun)

	// Initiate SAS attachment
	klog.InfoS("initiating SAS connection...")
	connector := saslib.Connector{VolumeWWN: wwn}
	path, err := saslib.Attach(ctx, &connector, &saslib.OSioHandler{})
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	klog.InfoS("attached device", "path", path)

	// if current wwn has been published before, remove it from our list of previously unpublished wwns
	delete(GlobalRemovedDevicesMap, wwn)
	// check if previously unpublished devices were rediscovered by the scsi subsystem during Attach
	checkPreviouslyRemovedDevices(ctx)

	if req.GetVolumeCapability().GetMount() != nil {
		fsType := GetFsType(req)
		err = EnsureFsType(fsType, path)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		corrupted := false
		if err = CheckFs(path, fsType, "Publish"); err != nil {
			corrupted = true
		}

		klog.InfoS("device multipath status", "multipath", connector.Multipath, "path", path, "wwn", wwn, "corrupted", corrupted)

		if corrupted {
			klog.InfoS("device corruption (publish)", "device", connector.OSPathName, "volume", volumeName, "multipath", connector.Multipath, " wwn", wwn, "corrupted", corrupted)
			DebugCorruption("$$", path)
			return nil, status.Errorf(codes.DataLoss, "(publish) filesystem (%v) seems to be corrupted: %v", path, err)
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
				return nil, status.Error(codes.Internal, string(out))
			}
		} else if len(mountpoints) == 1 {
			if mountpoints[0] == req.GetTargetPath() {
				klog.InfoS("volume already mounted", "targetPath", req.GetTargetPath())
			} else {
				errStr := fmt.Sprintf("device has already been mounted somewhere else (%s instead of %s), please unmount first", mountpoints[0], req.GetTargetPath())
				return nil, status.Error(codes.Internal, errStr)
			}
		} else if len(mountpoints) > 1 {
			return nil, errors.New("device has already been mounted in several locations, please unmount first")
		}
	} else if req.GetVolumeCapability().GetBlock() != nil {
		deviceFile, err := os.Create(req.GetTargetPath())
		if err != nil {
			klog.ErrorS(err, "could not create file", "TargetPath", req.GetTargetPath())
			return nil, err
		}
		deviceFile.Chmod(00755)
		deviceFile.Close()
		out, err := exec.Command("mount", "-o", "bind", path, req.GetTargetPath()).CombinedOutput()
		if err != nil {
			return nil, status.Error(codes.Internal, string(out))
		}
	}

	klog.Infof("saving SAS connection info in %s", sas.connectorInfoPath)
	if _, err := os.Stat(sas.connectorInfoPath); err == nil {
		klog.ErrorS(err, "sas connection file already exists", "connectorInfoPath", sas.connectorInfoPath)
	}
	err = connector.Persist(ctx, sas.connectorInfoPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	klog.InfoS("successfully mounted volume", "targetPath", req.GetTargetPath())
	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume unmounts the volume from the target path
func (sas *sasStorage) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
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

	klog.Infof("loading SAS connection info from %s", sas.connectorInfoPath)
	connector, err := saslib.GetConnectorFromFile(sas.connectorInfoPath)
	if err != nil {
		if os.IsNotExist(err) {
			klog.Warning(errors.Wrap(err, "assuming that SAS connection was already closed"))
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
	out, err := exec.Command("ls", "-l", fmt.Sprintf("/dev/disk/by-id/dm-name-3%s", wwn)).CombinedOutput()
	klog.Infof("check for dm-name: ls -l %s, err = %v, out = \n%s", fmt.Sprintf("/dev/disk/by-id/dm-name-3%s", wwn), err, string(out))

	if !connector.Multipath {
		// If we didn't discover the multipath device initially, double check that we didn't just miss it
		// Detach the discovered devices if they are found
		klog.V(3).Info("Device saved as non-multipath. Searching for additional devices before Detach")
		if connector.IoHandler == nil {
			connector.IoHandler = &saslib.OSioHandler{}
		}
		discoveredMpathName, devices := saslib.FindDiskById(klog.FromContext(ctx), wwn, connector.IoHandler)
		if (discoveredMpathName != connector.OSPathName) && (len(devices) > 0) {
			klog.V(0).Infof("Found additional linked devices: %s, %v", discoveredMpathName, devices)
			klog.V(0).Infof("Replacing original connector info prior to Detach, device: %s=>%s, linked device paths: %v=>%v", connector.OSPathName, discoveredMpathName, connector.OSDevicePaths, devices)
			connector.OSPathName = discoveredMpathName
			connector.OSDevicePaths = devices
			connector.Multipath = true
		}
	}

	klog.Info("DisconnectVolume, detaching SAS device")
	err = saslib.Detach(ctx, connector.OSPathName, connector.IoHandler)

	if err != nil {
		return nil, err
	}

	klog.Infof("deleting SAS connection info file %s", sas.connectorInfoPath)
	os.Remove(sas.connectorInfoPath)

	GlobalRemovedDevicesMap[connector.VolumeWWN] = time.Now()

	klog.Info("successfully detached SAS device")
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func checkPreviouslyRemovedDevices(ctx context.Context) error {
	klog.Info("Checking previously removed devices")
	for wwn := range GlobalRemovedDevicesMap {
		klog.Infof("Checking for rediscovery of wwn:%s", wwn)

		dm, devices := saslib.FindDiskById(klog.FromContext(ctx), wwn, &saslib.OSioHandler{})
		if dm != "" {
			klog.Infof("Rediscovery found for wwn:%s -- mpath device: %s, devices: %v", wwn, dm, devices)
			saslib.Detach(ctx, dm, &saslib.OSioHandler{})
		}
	}
	return nil
}

// NodeGetVolumeStats return info about a given volume
// Will not be called as the plugin does not have the GET_VOLUME_STATS capability
func (sas *sasStorage) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeGetVolumeStats is not implemented")
}

// NodeExpandVolume finalizes volume expansion on the node
func (sas *sasStorage) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {

	volumeName, _ := common.VolumeIdGetName(req.GetVolumeId())
	volumepath := req.GetVolumePath()
	klog.V(2).Infof("NodeExpandVolume: VolumeId=%v,  VolumePath=%v", volumeName, volumepath)

	if len(volumeName) == 0 {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("node expand volume requires volume id"))
	}

	if len(volumepath) == 0 {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("node expand volume requires volume path"))
	}

	connector, err := saslib.GetConnectorFromFile(sas.connectorInfoPath)
	klog.V(3).Infof("GetConnectorFromFile(%s) connector: %v, err: %v", volumeName, connector, err)

	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("node expand volume path not found for volume id (%s)", volumeName))
	}

	if connector.Multipath {
		klog.V(2).Info("device is using multipath")
		if err := saslib.ResizeMultipathDevice(ctx, connector.OSPathName); err != nil {
			return nil, err
		}
	} else {
		klog.V(2).Info("device is NOT using multipath")
	}

	if req.GetVolumeCapability().GetMount() != nil {
		klog.Infof("expanding filesystem using resize2fs on device %s", connector.OSPathName)
		output, err := exec.Command("resize2fs", connector.OSPathName).CombinedOutput()
		if err != nil {
			klog.V(2).InfoS("could not resize filesystem", "resize2fs output", output)
			return nil, fmt.Errorf("could not resize filesystem: %v", output)
		}
	}

	return &csi.NodeExpandVolumeResponse{}, nil
}

// NodeGetCapabilities returns the supported capabilities of the node server
func (sas *sasStorage) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeGetCapabilities is not implemented")
}

// NodeGetInfo returns info about the node
func (sas *sasStorage) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeGetInfo is not implemented")
}
