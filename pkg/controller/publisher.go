package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Seagate/seagate-exos-x-csi/pkg/common"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
)

// getConnectorInfoPath
func (driver *Controller) getConnectorInfoPath(volumeID string) string {
	return fmt.Sprintf("%s/%s.json", driver.runPath, volumeID)
}

// Read connector json file and return initiator address info for the given volume
func (driver *Controller) readInitiatorMapFromFile(filePath string, volumeID string) ([]string, error) {
	klog.Infof("Reading initiator value for volume %v from file %v", volumeID, filePath)
	f, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	initiatorMap := make(map[string][]string)
	err = json.Unmarshal(f, &initiatorMap)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling initiator info file for specified volume ID %v", volumeID)
	}
	initiators, found := initiatorMap[volumeID]
	if found {
		return initiators, nil
	} else {
		return nil, fmt.Errorf("initiator value for volume ID %v not found", volumeID)
	}
}

// PersistConnector persists the provided Connector to the specified file
func persistInitiatorMap(volumeID string, initiators []string, filePath string) error {
	initiatorMap := map[string][]string{
		volumeID: initiators,
	}
	f, err := os.Create(filePath)
	if err != nil {
		klog.Error("error encoding initiator info: %v", err)
		return fmt.Errorf("error creating initiator map file %s: %s", filePath, err)
	}
	defer f.Close()
	encoder := json.NewEncoder(f)
	if err = encoder.Encode(initiatorMap); err != nil {
		klog.Error("error encoding initiator info: %v", err)
		return fmt.Errorf("error encoding initiator info: %v", err)
	}
	klog.Infof("wrote initiator persistence file at %s", filePath)
	return nil
}

// ControllerPublishVolume attaches the given volume to the node
func (driver *Controller) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "cannot publish volume with empty ID")
	}
	if len(req.GetNodeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "cannot publish volume to a node with empty ID")
	}
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "cannot publish volume without capabilities")
	}
	parameters := req.GetVolumeContext()

	volumeName, _ := common.VolumeIdGetName(req.GetVolumeId())

	var initiatorNames []string
	// Available SAS initiators for the node are provided here through NodeGetInfo
	if parameters[common.StorageProtocolKey] == common.StorageProtocolSAS {
		for key, val := range parameters {
			if strings.Contains(key, common.TopologyInitiatorPrefix) {
				initiatorNames = append(initiatorNames, val)
			}
		}
	} else {
		initiatorNames = []string{req.GetNodeId()}
	}

	persistentInfoFilepath := driver.getConnectorInfoPath(req.GetVolumeId())
	persistInitiatorMap(volumeName, initiatorNames, persistentInfoFilepath)

	klog.Infof("attach request for initiator(s) %v, volume id: %s", initiatorNames, volumeName)

	lun, err := driver.client.PublishVolume(volumeName, initiatorNames)

	if err != nil {
		return nil, err
	}

	return &csi.ControllerPublishVolumeResponse{
		PublishContext: map[string]string{"lun": lun},
	}, err
}

// ControllerUnpublishVolume detaches the given volume from the node
func (driver *Controller) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "cannot unpublish volume with empty ID")
	}

	volumeName, _ := common.VolumeIdGetName(req.GetVolumeId())

	var initiators []string
	var err error
	if protocol, _ := common.VolumeIdGetStorageProtocol(req.GetVolumeId()); protocol == common.StorageProtocolSAS {
		initiators, err = driver.readInitiatorMapFromFile(driver.getConnectorInfoPath(req.GetVolumeId()), volumeName)
		if err != nil {
			return nil, fmt.Errorf("error retrieving initiator! cannot unpublish volume %v", volumeName)
		}
	} else {
		initiators = []string{req.GetNodeId()}
	}

	klog.Infof("unmapping volume %s from initiator %s", volumeName, initiators)
	for _, initiator := range initiators {
		_, status, err := driver.client.UnmapVolume(volumeName, initiator)
		if err != nil {
			if status != nil && status.ReturnCode == unmapFailedErrorCode {
				klog.Info("unmap failed, assuming volume is already unmapped")
			} else {
				klog.Errorf("unknown error while unmapping initiator %s: %v", initiator, err)
			}
		}
	}

	persistentInfoFilepath := driver.getConnectorInfoPath(req.GetVolumeId())
	os.Remove(persistentInfoFilepath)

	klog.Infof("successfully unmapped volume %s from all initiators", volumeName)
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}
