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
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

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

func (sas *sasStorage) AttachStorage(ctx context.Context, req *csi.NodePublishVolumeRequest) (string, error) {
	klog.InfoS("initiating SAS connection...")
	wwn, _ := common.VolumeIdGetWwn(req.GetVolumeId())
	connector := saslib.Connector{VolumeWWN: wwn}
	path, err := saslib.Attach(ctx, &connector, &saslib.OSioHandler{})
	if err != nil {
		return path, status.Error(codes.Unavailable, err.Error())
	}
	klog.InfoS("attached device", "path", path)
	err = connector.Persist(ctx, sas.connectorInfoPath)
	return path, err
}

func (sas *sasStorage) DetachStorage(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) error {
	klog.InfoS("loading SAS connection info from file", "connectorInfoPath", sas.connectorInfoPath)
	connector, err := saslib.GetConnectorFromFile(sas.connectorInfoPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			klog.ErrorS(err, "assuming that SAS connection was already closed")
			return nil
		}
		return status.Error(codes.Internal, err.Error())
	}
	klog.InfoS("connector.OSPathName", "connector.OSPathName", connector.OSPathName)

	if IsVolumeInUse(connector.OSPathName) {
		klog.InfoS("volume is still in use on the node, thus it will not be detached")
		return nil
	}

	_, err = os.Stat(connector.OSPathName)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		klog.ErrorS(err, "assuming that volume is already disconnected")
		return nil
	}

	wwn, _ := common.VolumeIdGetWwn(req.GetVolumeId())
	diskByIdPath := fmt.Sprintf("/dev/disk/by-id/dm-name-3%s", wwn)
	out, err := exec.Command("ls", "-l", diskByIdPath).CombinedOutput()
	klog.InfoS("check for dm-name", "command", fmt.Sprintf("ls -l %s, err = %v, out = \n%s", diskByIdPath, err, string(out)))

	if !connector.Multipath {
		// If we didn't discover the multipath device initially, double check that we didn't just miss it
		// Detach the discovered devices if they are found
		klog.V(3).InfoS("Device saved as non-multipath. Searching for additional devices before Detach")
		if connector.IoHandler == nil {
			connector.IoHandler = &saslib.OSioHandler{}
		}
		discoveredMpathName, devices := saslib.FindDiskById(klog.FromContext(ctx), wwn, connector.IoHandler)
		if (discoveredMpathName != connector.OSPathName) && (len(devices) > 0) {
			klog.V(0).InfoS("Found additional linked devices", "discoveredMpathName", discoveredMpathName, "devices", devices)
			klog.V(0).InfoS("Replacing original connector info prior to Detach",
				"originalDevice", connector.OSPathName, "newDevice", discoveredMpathName,
				"originalDevicePaths", connector.OSDevicePaths, "newDevicePaths", devices)
			connector.OSPathName = discoveredMpathName
			connector.OSDevicePaths = devices
			connector.Multipath = true
		}
	}

	klog.Info("DisconnectVolume, detaching SAS device")
	err = saslib.Detach(ctx, connector.OSPathName, connector.IoHandler)
	if err != nil {
		return err
	}

	klog.InfoS("deleting SAS connection info file", "sas.connectorInfoPath", sas.connectorInfoPath)
	os.Remove(sas.connectorInfoPath)
	return nil
}

func (sas *sasStorage) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "SAS specific NodePublishVolume not implemented")
}

// NodeUnpublishVolume unmounts the volume from the target path
func (sas *sasStorage) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "SAS specific NodeUnpublishVolume not implemented")
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
