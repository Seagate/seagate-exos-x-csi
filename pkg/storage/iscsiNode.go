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
	"strconv"
	"strings"
	"time"

	iscsilib "github.com/Seagate/csi-lib-iscsi/iscsi"
	"github.com/Seagate/seagate-exos-x-csi/pkg/common"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

// Configuration constants
const (
	BlkidTimeout      = 10
	maxDmnameAttempts = 18
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

func (iscsi *iscsiStorage) AttachStorage(ctx context.Context, req *csi.NodePublishVolumeRequest) (string, error) {
	wwn, _ := common.VolumeIdGetWwn(req.GetVolumeId())
	iqn := req.GetVolumeContext()["iqn"]
	portals := strings.Split(req.GetVolumeContext()["portals"], ",")
	klog.InfoS("iSCSI connection info:", "iqn", iqn, "portals", portals)

	lun, _ := strconv.ParseInt(req.GetPublishContext()["lun"], 10, 32)
	klog.InfoS("LUN:", "lun", lun)

	klog.InfoS("initiating ISCSI connection...")
	targets := make([]iscsilib.TargetInfo, 0)
	for _, portal := range portals {
		if portal != "" {
			klog.V(1).InfoS("-- add iqn and portal targets", "iqn", iqn, "portal", portal)
			targets = append(targets, iscsilib.TargetInfo{
				Iqn:    iqn,
				Portal: portal,
			})
			// test and produce a warning if path already exists before iscsi login
			devicePath := fmt.Sprintf("/dev/disk/by-path/ip-%s:3260-iscsi-%s-lun-%d", portal, iqn, lun)
			_, err := os.Stat(devicePath)
			klog.V(4).InfoS("[TEST] os stat device:", "exist", !os.IsNotExist(err), "device", devicePath)
			if !os.IsNotExist(err) {
				_, err := os.Stat(devicePath)
				klog.V(4).InfoS("WARNING: device exists before iscsi login:", "devicePath", devicePath, "os.Stat error", err)
			}
		}
	}

	// If CHAP secrets have been specified, include them in the iscsilib Connector
	doCHAPAuth := false
	authType := ""
	var iscsiSecrets iscsilib.Secrets
	if reqSecrets := req.GetSecrets(); reqSecrets != nil {
		CHAPusername := reqSecrets[common.CHAPUsernameKey]
		CHAPpassword := reqSecrets[common.CHAPSecretKey]
		CHAPusernameIn := reqSecrets[common.CHAPUsernameInKey]
		CHAPpasswordIn := reqSecrets[common.CHAPPasswordInKey]
		if CHAPusername != "" && CHAPpassword != "" {
			doCHAPAuth = true
			authType = "chap"
			iscsiSecrets = iscsilib.Secrets{
				SecretsType: "chap",
				UserName:    CHAPusername,
				Password:    CHAPpassword,
				UserNameIn:  CHAPusernameIn,
				PasswordIn:  CHAPpasswordIn,
			}
		}
	}

	klog.V(4).InfoS("iscsi connector setup", "AuthType", authType, "Targets", targets, "Lun", lun)
	connector := &iscsilib.Connector{
		AuthType:         authType,
		Targets:          targets,
		Lun:              int32(lun),
		DoDiscovery:      true,
		DoCHAPDiscovery:  doCHAPAuth,
		DiscoverySecrets: iscsiSecrets,
		SessionSecrets:   iscsiSecrets,
		RetryCount:       20,
	}

	path, err := iscsilib.Connect(connector)
	if err != nil {
		return "", err
	}
	klog.InfoS("attached device:", "path", path)

	exists := true
	out, err := exec.Command("ls", "-l", fmt.Sprintf("/dev/disk/by-id/dm-name-3%s", wwn)).CombinedOutput()
	klog.V(1).InfoS("ls command output", "command", fmt.Sprintf("ls -l /dev/disk/by-id/dm-name-3%s", wwn), "err", err, "out", out)
	if err != nil {
		exists = false
	}

	// wait here until the dm-name exists, for debugging
	if !exists {
		attempts := 1
		for attempts < (maxDmnameAttempts + 1) {
			// Force a reload of all existing multipath maps
			output, err := exec.Command("multipath", "-r").CombinedOutput()
			klog.V(4).InfoS("## (publish) multipath -r output", "err", err, "output", output)

			out, err := exec.Command("ls", "-l", fmt.Sprintf("/dev/disk/by-id/dm-name-3%s", wwn)).CombinedOutput()
			klog.V(1).InfoS("check for dm-name exists", "attempt", attempts, "command", fmt.Sprintf("ls -l /dev/disk/by-id/dm-name-3%s", wwn), "err", err, "out", out)
			if err == nil {
				exists = true
				break
			}
			time.Sleep(dmnameDelay * time.Second)
			attempts++
		}
	}
	if _, err := os.Stat(iscsi.connectorInfoPath); err == nil {
		klog.InfoS("iscsi connection file already exists", "connectorInfoPath", iscsi.connectorInfoPath)
	}

	klog.InfoS("saving ISCSI connection info", "connectorInfoPath", iscsi.connectorInfoPath)
	if _, err := os.Stat(iscsi.connectorInfoPath); err == nil {
		klog.InfoS("iscsi connection file already exists", "connectorInfoPath", iscsi.connectorInfoPath)
	}
	err = iscsilib.PersistConnector(connector, iscsi.connectorInfoPath)
	if err != nil {
		return "", err
	}

	return path, nil
}

func (iscsi *iscsiStorage) DetachStorage(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) error {
	klog.Infof("loading ISCSI connection info from %s", iscsi.connectorInfoPath)
	connector, err := iscsilib.GetConnectorFromFile(iscsi.connectorInfoPath)
	if err != nil {
		if os.IsNotExist(err) {
			klog.InfoS("assuming that ISCSI connection is already closed")
			return nil
		}
		return status.Error(codes.Internal, err.Error())
	}
	klog.InfoS("connector.DevicePath", "connector.DevicePath", connector.DevicePath)

	if IsVolumeInUse(connector.DevicePath) {
		klog.Info("volume is still in use on the node, thus it will not be detached")
		return nil
	}

	_, err = os.Stat(connector.DevicePath)
	if err != nil && os.IsNotExist(err) {
		klog.InfoS("connector.devicePath does not exist, assuming that volume is already disconnected")
		return nil
	}

	wwn, _ := common.VolumeIdGetWwn(req.GetVolumeId())
	out, err := exec.Command("ls", "-l", fmt.Sprintf("/dev/disk/by-id/dm-name-3%s", wwn)).CombinedOutput()
	klog.Infof("check for dm-name: ls -l %s, err = %v, out = \n%s", fmt.Sprintf("/dev/disk/by-id/dm-name-3%s", wwn), err, string(out))

	klog.Info("DisconnectVolume, detaching ISCSI device")
	err = iscsilib.DisconnectVolume(*connector)
	if err != nil {
		return err
	}

	klog.Infof("deleting ISCSI connection info file %s", iscsi.connectorInfoPath)
	os.Remove(iscsi.connectorInfoPath)
	return nil
}

func (iscsi *iscsiStorage) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "iSCSI specific NodePublishVolume not implemented")
}

// NodeUnpublishVolume unmounts the volume from the target path
func (iscsi *iscsiStorage) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "iSCSI specific NodeUnpublishVolume not implemented")
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

	connector, err := iscsilib.GetConnectorFromFile(iscsi.connectorInfoPath)
	klog.V(3).Infof("GetConnectorFromFile(%s) connector: %v, err: %v", volumeName, connector, err)

	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("node expand volume path not found for volume id (%s)", volumeName))
	}

	if connector.Multipath {
		klog.V(2).Info("device is using multipath")
		if err := iscsilib.ResizeMultipathDevice(connector.DevicePath); err != nil {
			return nil, err
		}
	} else {
		klog.V(2).Info("device is NOT using multipath")
	}

	if req.GetVolumeCapability().GetMount() != nil {
		klog.Infof("expanding filesystem using resize2fs on device %s", connector.DevicePath)
		output, err := exec.Command("resize2fs", connector.DevicePath).CombinedOutput()
		if err != nil {
			klog.V(2).InfoS("could not resize filesystem", "resize2fs output", output)
			return nil, fmt.Errorf("could not resize filesystem: %v", output)
		}
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

func GetISCSIInitiators() ([]string, error) {
	initiatorNameFilePath := "/etc/iscsi/initiatorname.iscsi"
	file, err := os.Open(initiatorNameFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if equal := strings.Index(line, "="); equal >= 0 {
			if strings.TrimSpace(line[:equal]) == "InitiatorName" {
				return []string{strings.TrimSpace(line[equal+1:])}, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("InitiatorName key is missing from %s", initiatorNameFilePath)
}
