package controller

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	storageapi "github.com/Seagate/seagate-exos-x-api-go/pkg/v1"
	"github.com/Seagate/seagate-exos-x-csi/pkg/common"
	"github.com/Seagate/seagate-exos-x-csi/pkg/node_service"
	pb "github.com/Seagate/seagate-exos-x-csi/pkg/node_service/node_servicepb"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

var volumeCapabilities = []*csi.VolumeCapability{
	{
		AccessType: &csi.VolumeCapability_Mount{
			Mount: &csi.VolumeCapability_MountVolume{},
		},
		AccessMode: &csi.VolumeCapability_AccessMode{
			Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		},
	},
}

var csiMutexes = map[string]*sync.Mutex{
	"/csi.v1.Controller/CreateVolume":              {},
	"/csi.v1.Controller/ControllerPublishVolume":   {},
	"/csi.v1.Controller/DeleteVolume":              {},
	"/csi.v1.Controller/ControllerUnpublishVolume": {},
	"/csi.v1.Controller/ControllerExpandVolume":    {},
}

var nonAuthenticatedMethods = []string{
	"/csi.v1.Controller/ControllerGetCapabilities",
	"/csi.v1.Controller/ListVolumes",
	"/csi.v1.Controller/GetCapacity",
	"/csi.v1.Controller/ControllerGetVolume",
	"/csi.v1.Identity/Probe",
	"/csi.v1.Identity/GetPluginInfo",
	"/csi.v1.Identity/GetPluginCapabilities",
}

// Controller is the implementation of csi.ControllerServer
type Controller struct {
	*common.Driver

	client             *storageapi.Client
	nodeServiceClients map[string]*grpc.ClientConn
	runPath            string
}

// DriverCtx contains data common to most calls
type DriverCtx struct {
	Credentials map[string]string
	Parameters  map[string]string
	VolumeCaps  *[]*csi.VolumeCapability
}

// New is a convenience fn for creating a controller driver
func New() *Controller {
	client := storageapi.NewClient()
	controller := &Controller{
		Driver:             common.NewDriver(client.Collector),
		client:             client,
		runPath:            fmt.Sprintf("/var/run/%s", common.PluginName),
		nodeServiceClients: map[string]*grpc.ClientConn{},
	}

	if err := os.MkdirAll(controller.runPath, 0755); err != nil {
		panic(err)
	}

	controller.InitServer(
		func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			if mutex, exists := csiMutexes[info.FullMethod]; exists {
				mutex.Lock()
				defer mutex.Unlock()
			}
			return handler(ctx, req)
		},
		common.NewLogRoutineServerInterceptor(func(string) bool {
			return true
		}),
		func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			driverContext := DriverCtx{}
			reqWithSecrets, ok := req.(common.WithSecrets)
			if ok {
				driverContext.Credentials = reqWithSecrets.GetSecrets()
			}
			if reqWithParameters, ok := req.(common.WithParameters); ok {
				driverContext.Parameters = reqWithParameters.GetParameters()
			}
			if reqWithVolumeCaps, ok := req.(common.WithVolumeCaps); ok {
				driverContext.VolumeCaps = reqWithVolumeCaps.GetVolumeCapabilities()
			}

			err := controller.beginRoutine(&driverContext, info.FullMethod)
			if err != nil {
				klog.Infof("controller.beginRoutine error for req = %x", reqWithSecrets)
			}
			defer controller.endRoutine()
			if err != nil {
				return nil, err
			}
			return handler(ctx, req)
		},
	)

	csi.RegisterIdentityServer(controller.Server, controller)
	csi.RegisterControllerServer(controller.Server, controller)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	go func() {
		_ = <-sigc
		controller.Stop()
	}()

	return controller
}

// ControllerGetCapabilities returns the capabilities of the controller service.
func (controller *Controller) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	var csc []*csi.ControllerServiceCapability
	cl := []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
		csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS,
		csi.ControllerServiceCapability_RPC_CLONE_VOLUME,
		csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
	}

	for _, cap := range cl {
		klog.V(4).Infof("enabled controller service capability: %v", cap.String())
		csc = append(csc, &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: cap,
				},
			},
		})
	}

	return &csi.ControllerGetCapabilitiesResponse{Capabilities: csc}, nil
}

// ValidateVolumeCapabilities checks whether a provisioned volume supports the capabilities requested
func (controller *Controller) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	volumeName, _ := common.VolumeIdGetName(req.GetVolumeId())

	if len(volumeName) == 0 {
		return nil, status.Error(codes.InvalidArgument, "cannot validate volume with empty ID")
	}
	if len(req.GetVolumeCapabilities()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "cannot validate volume without capabilities")
	}
	_, _, err := controller.client.ShowVolumes(volumeName)
	if err != nil {
		return nil, status.Error(codes.NotFound, "cannot validate volume not found")
	}

	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeCapabilities: req.GetVolumeCapabilities(),
		},
	}, nil
}

// ListVolumes returns a list of all requested volumes
func (controller *Controller) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ListVolumes is unimplemented and should not be called")
}

// GetCapacity returns the capacity of the storage pool
func (controller *Controller) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetCapacity is unimplemented and should not be called")
}

// ControllerGetVolume fetch current information about a volume
func (controller *Controller) ControllerGetVolume(ctx context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ControllerGetVolume is unimplemented and should not be called")
}

// Probe returns the health and readiness of the plugin
func (controller *Controller) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	return &csi.ProbeResponse{}, nil
}

func (controller *Controller) beginRoutine(ctx *DriverCtx, methodName string) error {
	if err := runPreflightChecks(ctx.Parameters, ctx.VolumeCaps); err != nil {
		return err
	}

	needsAuthentication := true
	for _, name := range nonAuthenticatedMethods {
		if methodName == name {
			needsAuthentication = false
			break
		}
	}

	if !needsAuthentication {
		return nil
	}

	if ctx.Credentials == nil {
		return errors.New("missing API credentials")
	}

	return controller.configureClient(ctx.Credentials)
}

func (controller *Controller) endRoutine() {
	controller.client.CloseConnections()
}

func (controller *Controller) configureClient(credentials map[string]string) error {
	username := string(credentials[common.UsernameSecretKey])
	password := string(credentials[common.PasswordSecretKey])
	apiAddr := string(credentials[common.APIAddressConfigKey])

	if len(username) == 0 {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("(%s) is missing from secrets", common.UsernameSecretKey))
	}

	if len(password) == 0 {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("(%s) is missing from secrets", common.PasswordSecretKey))
	}

	if len(apiAddr) == 0 {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("(%s) is missing from secrets", common.APIAddressConfigKey))
	}

	klog.Infof("using API at address (%s)", apiAddr)
	if controller.client.SessionValid(apiAddr, username) {
		return nil
	}

	klog.Infof("login to API address %q as user %q", apiAddr, username)
	controller.client.StoreCredentials(apiAddr, username, password)
	err := controller.client.Login()
	if err != nil {
		return status.Error(codes.Unauthenticated, err.Error())
	}

	klog.Info("login was successful")
	err = controller.client.InitSystemInfo()

	return err
}

func runPreflightChecks(parameters map[string]string, capabilities *[]*csi.VolumeCapability) error {
	checkIfKeyExistsInConfig := func(key string) error {
		if parameters == nil {
			return nil
		}

		klog.V(2).Infof("checking for %s in storage class parameters", key)
		_, ok := parameters[key]
		if !ok {
			return status.Errorf(codes.InvalidArgument, "'%s' is missing from configuration", key)
		}
		return nil
	}

	if err := checkIfKeyExistsInConfig(common.PoolConfigKey); err != nil {
		return err
	}

	if capabilities != nil {
		if len(*capabilities) == 0 {
			return status.Error(codes.InvalidArgument, "missing volume capabilities")
		}
		for _, capability := range *capabilities {
			accessMode := capability.GetAccessMode().GetMode()
			accessModeSupported := false
			for _, mode := range common.SupportedAccessModes {
				if accessMode == mode {
					accessModeSupported = true
				}
			}
			if !accessModeSupported {
				return status.Errorf(codes.FailedPrecondition, "driver does not support access mode %v", accessMode)
			}
			if mount := capability.GetMount(); mount != nil {
				if mount.GetFsType() == "" {
					if err := checkIfKeyExistsInConfig(common.FsTypeConfigKey); err != nil {
						return status.Error(codes.FailedPrecondition, "no fstype specified in storage class")
					} else {
						klog.InfoS("storage class parameter "+common.FsTypeConfigKey+" is deprecated. Please migrate to 'csi.storage.k8s.io/fstype'", "parameter", common.FsTypeConfigKey)
					}
				}
			}
		}
	}
	return nil
}

// Makes an RPC call to the specified node to retrieve initiators of the specified type (iSCSI,FC,SAS)
// Handles re-use of the relatively expensive grpc Channel(grpc.ClientConn)
// The gRPC stub is created and destroyed on each call
func (controller *Controller) GetNodeInitiators(ctx context.Context, nodeAddress string, protocol string) ([]string, error) {
	var reqType pb.InitiatorType
	switch protocol {
	case common.StorageProtocolSAS:
		reqType = pb.InitiatorType_SAS
	case common.StorageProtocolFC:
		reqType = pb.InitiatorType_FC
	case common.StorageProtocolISCSI:
		reqType = pb.InitiatorType_ISCSI
	}

	clientConnection := controller.nodeServiceClients[nodeAddress]
	if clientConnection == nil {
		klog.V(3).InfoS("node grpc client not found, establishing...", "nodeAddress", nodeAddress)
		var err error
		clientConnection, err = node_service.InitializeClient(nodeAddress)
		if err != nil {
			return nil, err
		}
		controller.nodeServiceClients[nodeAddress] = clientConnection
	}
	initiators, err := node_service.GetNodeInitiators(ctx, clientConnection, reqType)
	return initiators, err
}

func (controller *Controller) NotifyUnmap(ctx context.Context, nodeAddress string, volumeName string) error {
	clientConnection := controller.nodeServiceClients[nodeAddress]
	if clientConnection == nil {
		klog.V(3).InfoS("node grpc client not found, establishing...", "nodeAddress", nodeAddress)
		var err error
		clientConnection, err = node_service.InitializeClient(nodeAddress)
		if err != nil {
			return err
		}
		controller.nodeServiceClients[nodeAddress] = clientConnection
	}
	return node_service.NotifyUnmap(ctx, clientConnection, volumeName)
}

// Graceful shutdown of Node-Controller RPC Clients
func (controller *Controller) Stop() {
	klog.V(3).InfoS("Controller code graceful shutdown..")
	for nodeIP, clientConn := range controller.nodeServiceClients {
		klog.V(3).InfoS("Closing node client", "nodeIP", nodeIP)
		clientConn.Close()
	}
	controller.Driver.Stop()
}
