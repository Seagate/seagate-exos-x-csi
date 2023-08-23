package node

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/Seagate/csi-lib-iscsi/iscsi"
	"github.com/Seagate/seagate-exos-x-csi/pkg/common"
	"github.com/Seagate/seagate-exos-x-csi/pkg/node_service"
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

	semaphore  *semaphore.Weighted
	runPath    string
	nodeName   string
	nodeIP     string
	nodeServer *grpc.Server
}

// New is a convenience function for creating a node driver
func New() *Node {
	if klog.V(8).Enabled() {
		iscsi.EnableDebugLogging(os.Stderr)
	}

	envNodeName, _ := os.LookupEnv(common.NodeNameEnvVar)
	nodeIP, envFound := os.LookupEnv(common.NodeIPEnvVar)
	if !envFound {
		klog.InfoS("no Node IP found in environment. Using default")
		nodeIP = "127.0.0.1"
	}
	envServicePort, envFound := os.LookupEnv(common.NodeServicePortEnvVar)
	if !envFound {
		klog.InfoS("no node service port found in environment. Using default")
		envServicePort = "978"
	}

	node := &Node{
		Driver:    common.NewDriver(),
		semaphore: semaphore.NewWeighted(1),
		runPath:   fmt.Sprintf("/var/run/%s", common.PluginName),
		nodeName:  envNodeName,
		nodeIP:    nodeIP,
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
			klog.Infof(">>> %s", info.FullMethod)
			if info.FullMethod == "/csi.v1.Node/NodePublishVolume" {
				if err := node.semaphore.Acquire(ctx, 1); err != nil {
					klog.Infof(">>> %s FAILED to acquire semaphore", info.FullMethod)
					return nil, status.Error(codes.Aborted, "node busy: too many concurrent volume publications, try again later")
				}
				defer node.semaphore.Release(1)
				klog.Infof(">>> %s acquired semaphore", info.FullMethod)
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

	// initialize node-controller communication service
	node.nodeServer = grpc.NewServer()
	go node_service.ListenAndServe(node.nodeServer, envServicePort)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	go func() {
		_ = <-sigc
		node.Stop()
	}()

	return node
}

// NodeGetInfo returns info about the node
func (node *Node) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return &csi.NodeGetInfoResponse{
		NodeId:            node.nodeIP,
		MaxVolumesPerNode: 255,
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
	// Extract the volume name and the storage protocol from the augmented volume id
	volumeName, _ := common.VolumeIdGetName(req.GetVolumeId())
	storageProtocol, _ := common.VolumeIdGetStorageProtocol(req.GetVolumeId())

	// Ensure that NodePublishVolume is only called once per volume
	storage.AddGatekeeper(volumeName)
	defer storage.RemoveGatekeeper(volumeName)

	klog.InfoS("NodePublishVolume call", "volumeName", volumeName)

	config := make(map[string]string)
	config["connectorInfoPath"] = node.getConnectorInfoPath(storageProtocol, volumeName)
	klog.V(2).Infof("NodePublishVolume connectorInfoPath (%v)", config["connectorInfoPath"])

	// Get storage handler
	storageNode, err := storage.NewStorageNode(storageProtocol, config)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// Do any required device discovery and return path of the device on the node fs
	path, err := storageNode.AttachStorage(ctx, req)
	if req.GetVolumeCapability().GetMount() != nil {
		err = storage.MountFilesystem(req, path)
	} else if req.GetVolumeCapability().GetBlock() != nil {
		err = storage.MountDevice(req, path)
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &csi.NodePublishVolumeResponse{}, nil
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

// Graceful shutdown of the node-controller RPC server
func (node *Node) Stop() {
	klog.V(3).InfoS("Node graceful shutdown..")
	node.nodeServer.GracefulStop()
	node.Driver.Stop()
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
