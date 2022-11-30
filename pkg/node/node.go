package node

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Seagate/csi-lib-iscsi/iscsi"
	"github.com/Seagate/seagate-exos-x-csi/pkg/common"
	"github.com/Seagate/seagate-exos-x-csi/pkg/storage"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes/wrappers"
	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

// Node is the implementation of csi.NodeServer
type Node struct {
	*common.Driver

	semaphore *semaphore.Weighted
	runPath   string
}

// New is a convenience function for creating a node driver
func New() *Node {
	if klog.V(8).Enabled() {
		iscsi.EnableDebugLogging(os.Stderr)
	}

	node := &Node{
		Driver:    common.NewDriver(),
		semaphore: semaphore.NewWeighted(1),
		runPath:   fmt.Sprintf("/var/run/%s", common.PluginName),
	}

	if err := os.MkdirAll(node.runPath, 0755); err != nil {
		panic(err)
	}

	klog.Infof("Node initializing with path: %s", node.runPath)

	requiredBinaries := []string{
		"blkid",      // command-line utility to locate/print block device attributes
		"findmnt",    // find a filesystem
		"iscsiadm",   // iscsi administration
		"mount",      // mount a filesystem
		"mountpoint", // see if a directory or file is a mountpoint
		"multipath",  // device mapping multipathing
		"multipathd", // device mapping multipathing
		"umount",     // unmount file systems
		"dmsetup",    // device-mapper to remove/clean dm entries

		// "blockdev",    // call block device ioctls from the command line
		// "lsblk",       // list block devices
		// "scsi_id",     // retrieve and generate a unique SCSI identifier
		//	"e2fsck",     // check a Linux ext2/ext3/ext4 file system
		//	"mkfs.ext4",  // create an ext2/ext3/ext4 filesystem
		//	"resize2fs",  // ext2/ext3/ext4 file system resizer
	}

	klog.Infof("Checking (%d) binaries", len(requiredBinaries))

	for _, binaryName := range requiredBinaries {
		if err := checkHostBinary(binaryName); err != nil {
			klog.Warningf("Error locating binary %q", binaryName)
		}
	}

	node.InitServer(
		func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			if info.FullMethod == "/csi.v1.Node/NodePublishVolume" {
				if !node.semaphore.TryAcquire(1) {
					return nil, status.Error(codes.Aborted, "node busy: too many concurrent volume publication, try again later")
				}
				defer node.semaphore.Release(1)
			}
			return handler(ctx, req)
		},
		common.NewLogRoutineServerInterceptor(func(fullMethod string) bool {
			return fullMethod == "/csi.v1.Node/NodePublishVolume" ||
				fullMethod == "/csi.v1.Node/NodeUnpublishVolume" ||
				fullMethod == "/csi.v1.Node/NodeExpandVolume"
		}),
	)

	csi.RegisterIdentityServer(node.Server, node)
	csi.RegisterNodeServer(node.Server, node)

	return node
}

// NodeGetInfo returns info about the node
func (node *Node) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	initiatorName, err := readInitiatorName()
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	topology := map[string]string{}
	sasAddresses, err := getSASAddresses()
	if err != nil {
		klog.Infof("Error while searching for SAS HBA Addresses: %s", err)
	}

	for i, sasAddr := range sasAddresses {
		//maximum value length 63 chars
		topoKey := fmt.Sprintf("%s/%s-%d", common.TopologyInitiatorPrefix, common.TopologySASInitiatorLabel, i)
		topology[topoKey] = sasAddr
	}

	return &csi.NodeGetInfoResponse{
		NodeId:            initiatorName,
		MaxVolumesPerNode: 255,
		AccessibleTopology: &csi.Topology{
			Segments: topology,
		},
	}, nil
}

// NodeGetCapabilities returns the supported capabilities of the node server
func (node *Node) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	var csc []*csi.NodeServiceCapability
	cl := []csi.NodeServiceCapability_RPC_Type{
		csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
	}

	for _, cap := range cl {
		// klog.V(4).Infof("enabled node service capability: %v", cap.String())
		csc = append(csc, &csi.NodeServiceCapability{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: cap,
				},
			},
		})
	}

	return &csi.NodeGetCapabilitiesResponse{Capabilities: csc}, nil
}

// NodePublishVolume mounts the volume mounted to the staging path to the target path
func (node *Node) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {

	// Extract the volume name and the storage protocol from the augmented volume id
	volumeName, _ := common.VolumeIdGetName(req.GetVolumeId())
	storageProtocol, _ := common.VolumeIdGetStorageProtocol(req.GetVolumeId())

	klog.Infof("NodePublishVolume called with volume name %s", volumeName)

	config := make(map[string]string)
	config["connectorInfoPath"] = node.getConnectorInfoPath(storageProtocol, volumeName)
	klog.V(2).Infof("NodePublishVolume connectorInfoPath (%v)", config["connectorInfoPath"])

	// Get storage handler
	storageNode, err := storage.NewStorageNode(storageProtocol, config)
	if storageNode != nil {
		return storageNode.NodePublishVolume(ctx, req)
	}

	klog.Errorf("NodePublishVolume error for storage protocol (%v): %v", storageProtocol, err)
	return nil, status.Errorf(codes.Internal, "Unable to process for storage protocol (%v)", storageProtocol)
}

// NodeUnpublishVolume unmounts the volume from the target path
func (node *Node) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {

	// Extract the volume name and the storage protocol from the augmented volume id
	volumeName, _ := common.VolumeIdGetName(req.GetVolumeId())
	storageProtocol, _ := common.VolumeIdGetStorageProtocol(req.GetVolumeId())

	klog.Infof("NodeUnpublishVolume volume %s at target path %s", volumeName, req.GetTargetPath())

	config := make(map[string]string)
	config["connectorInfoPath"] = node.getConnectorInfoPath(storageProtocol, volumeName)
	klog.V(2).Infof("NodeUnpublishVolume connectorInfoPath (%v)", config["connectorInfoPath"])

	// Get storage handler
	storageNode, err := storage.NewStorageNode(storageProtocol, config)
	if storageNode != nil {
		return storageNode.NodeUnpublishVolume(ctx, req)
	}

	klog.Errorf("NodeUnpublishVolume error for storage protocol (%v): %v", storageProtocol, err)
	return nil, status.Errorf(codes.Internal, "Unable to process for storage protocol (%v)", storageProtocol)
}

// NodeExpandVolume finalizes volume expansion on the node
func (node *Node) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {

	// Extract the volume name and the storage protocol from the augmented volume id
	volumeName, _ := common.VolumeIdGetName(req.GetVolumeId())
	storageProtocol, _ := common.VolumeIdGetStorageProtocol(req.GetVolumeId())

	klog.Infof("NodeExpandVolume volume %s at volume path %s", volumeName, req.GetVolumePath())

	config := make(map[string]string)
	config["connectorInfoPath"] = node.getConnectorInfoPath(storageProtocol, volumeName)
	klog.V(2).Infof("NodeExpandVolume connectorInfoPath (%v)", config["connectorInfoPath"])

	// Get storage handler
	storageNode, err := storage.NewStorageNode(storageProtocol, config)
	if storageNode != nil {
		return storageNode.NodeExpandVolume(ctx, req)
	}

	klog.Errorf("NodeExpandVolume error for storage protocol (%v): %v", storageProtocol, err)
	return nil, status.Errorf(codes.Internal, "Unable to process for storage protocol (%v)", storageProtocol)
}

// NodeGetVolumeStats return info about a given volume
// Will not be called as the plugin does not have the GET_VOLUME_STATS capability
func (node *Node) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeGetVolumeStats is not implemented")
}

// NodeStageVolume mounts the volume to a staging path on the node. This is
// called by the CO before NodePublishVolume and is used to temporary mount the
// volume to a staging path. Once mounted, NodePublishVolume will make sure to
// mount it to the appropriate path
// Will not be called as the plugin does not have the STAGE_UNSTAGE_VOLUME capability
func (node *Node) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeStageVolume is not implemented")
}

// NodeUnstageVolume unstages the volume from the staging path
// Will not be called as the plugin does not have the STAGE_UNSTAGE_VOLUME capability
func (node *Node) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeUnstageVolume is not implemented")
}

// Probe returns the health and readiness of the plugin
func (node *Node) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	// klog.V(4).Infof("Probe called with args: %#v", req)
	return &csi.ProbeResponse{Ready: &wrappers.BoolValue{Value: true}}, nil
}

// getConnectorInfoPath
func (node *Node) getConnectorInfoPath(storageProtocol, volumeID string) string {
	return fmt.Sprintf("%s/%s-%s.json", node.runPath, storageProtocol, volumeID)
}

// checkHostBinary: Determine if a binary image is installed or not
func checkHostBinary(name string) error {
	if path, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("binary %q not found", name)
	} else {
		klog.V(5).Infof("found binary %q in host PATH at %q", name, path)
	}

	return nil
}

// readInitiatorName: Extract the initiator name from /etc/iscsi file
func readInitiatorName() (string, error) {

	initiatorNameFilePath := "/etc/iscsi/initiatorname.iscsi"
	file, err := os.Open(initiatorNameFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if equal := strings.Index(line, "="); equal >= 0 {
			if strings.TrimSpace(line[:equal]) == "InitiatorName" {
				return strings.TrimSpace(line[equal+1:]), nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("InitiatorName key is missing from %s", initiatorNameFilePath)
}

// Read the sas address configuration file and return addresses
func readSASAddrFile() ([]string, error) {

	SASFilePath := "/etc/kubernetes/sas-addresses"
	file, err := os.Open(SASFilePath)
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

// Retrieve SAS Host addresses for the node
func getSASAddresses() ([]string, error) {

	// look for file specifying addresses. Skip discovery process if it exists
	foundAddresses, err := readSASAddrFile()
	if err == nil && foundAddresses != nil {
		klog.Infof("using SAS addresses from config file: %v", foundAddresses)
		return foundAddresses, nil
	} else if errors.Is(err, os.ErrNotExist) {
		klog.Infof("SAS address config file not found")
	} else if err != nil {
		klog.Warningf("error attempting to read SAS host address file: %s", err)
	}

	klog.Infof("begin SAS address discovery")
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
				klog.Warningf("error searching for HBA addresses: %v", err)
				return nil, err
			}
		} else {
			klog.Infof("found HBA address %s", address)
			if firstAddress, err := strconv.ParseInt(address, 16, 0); err != nil {
				return nil, err
			} else {
				secondAddress := strconv.FormatInt(firstAddress+1, 16)
				foundAddresses = append(foundAddresses, address, secondAddress)
			}
		}
	}
	return foundAddresses, nil
}
