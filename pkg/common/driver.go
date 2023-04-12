package common

import (
	"context"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Seagate/seagate-exos-x-csi/pkg/exporter"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/google/uuid"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

// PluginName is the public name to be used in storage class etc.
const PluginName = "csi-exos-x.seagate.com"

// Configuration constants
const (
	AugmentKey                = "##"
	FsTypeConfigKey           = "fsType"
	PoolConfigKey             = "pool"
	APIAddressConfigKey       = "apiAddress"
	UsernameSecretKey         = "username"
	PasswordSecretKey         = "password"
	StorageClassAnnotationKey = "storageClass"
	VolumePrefixKey           = "volPrefix"
	WWNs                      = "wwns"
	StorageProtocolKey        = "storageProtocol"
	StorageProtocolISCSI      = "iscsi"
	StorageProtocolFC         = "fc"
	StorageProtocolSAS        = "sas"
	TopologyInitiatorPrefix   = "com.seagate-exos-x-csi"
	TopologySASInitiatorLabel = "sas-address"
	TopologyFCInitiatorLabel  = "fc-address"
	TopologyNodeIdentifier    = "node-id"
	TopologyNodeIDKey         = TopologyInitiatorPrefix + "/" + TopologyNodeIdentifier

	MaximumLUN            = 255
	VolumeNameMaxLength   = 31
	VolumePrefixMaxLength = 3

	//If changed, must also be updated in helm charts
	NodeIPEnvVar          = "CSI_NODE_IP"
	NodeNameEnvVar        = "CSI_NODE_NAME"
	NodeServicePortEnvVar = "CSI_NODE_SERVICE_PORT"
)

// Driver contains main resources needed by the driver and references the underlying specific driver
type Driver struct {
	Server *grpc.Server

	socket   net.Listener
	exporter *exporter.Exporter
}

// WithSecrets is an interface for structs with secrets
type WithSecrets interface {
	GetSecrets() map[string]string
}

// WithParameters is an interface for structs with parameters
type WithParameters interface {
	GetParameters() map[string]string
}

// WithVolumeCaps is an interface for structs with volume capabilities
type WithVolumeCaps interface {
	GetVolumeCapabilities() *[]*csi.VolumeCapability
}

// NewDriver is a convenience function for creating an abstract driver
func NewDriver(collectors ...prometheus.Collector) *Driver {
	exporter := exporter.New(9842)

	for _, collector := range collectors {
		exporter.RegisterCollector(collector)
	}

	return &Driver{exporter: exporter}
}

var routineTimers = map[string]time.Time{}

func (driver *Driver) InitServer(unaryServerInterceptors ...grpc.UnaryServerInterceptor) {
	interceptors := append([]grpc.UnaryServerInterceptor{
		func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			start := time.Now()
			resp, err := handler(ctx, req)
			driver.exporter.Collector.IncCSIRPCCall(info.FullMethod, err == nil)
			driver.exporter.Collector.AddCSIRPCCallDuration(info.FullMethod, time.Since(start))
			return resp, err
		},
	}, unaryServerInterceptors...)

	driver.Server = grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(interceptors...)),
	)
}

var routineDepth = 0
var mu sync.Mutex
var useMutex = false

func NewLogRoutineServerInterceptor(shouldLogRoutine func(string) bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if shouldLogRoutine(info.FullMethod) {
			uuid := uuid.New().String()
			shortuuid := uuid[strings.LastIndex(uuid, "-")+1:]
			klog.Infof("=== [ROUTINE REQUEST] [%d] %s (%s) <0s> ===", routineDepth, info.FullMethod, shortuuid)
			routineTimers[shortuuid] = time.Now()
			if useMutex {
				mu.Lock()
			}
			routineDepth++
			duration := time.Since(routineTimers[shortuuid])
			klog.Infof("=== [ROUTINE START] [%d] %s (%s) <%s> ===", routineDepth, info.FullMethod, shortuuid, duration)
			defer func() {
				routineDepth--
				duration := time.Since(routineTimers[shortuuid])
				klog.Infof("=== [ROUTINE END] [%d] %s (%s) <%s> ===", routineDepth, info.FullMethod, shortuuid, duration)
				delete(routineTimers, shortuuid)
				if useMutex {
					mu.Unlock()
				}
			}()
		}

		result, err := handler(ctx, req)
		if err != nil {
			klog.Error(err)
		}

		return result, err
	}
}

// Start does the boilerplate stuff for starting the driver
// it loads its configuration from cli flags
func (driver *Driver) Start(bind string) {

	var ll klog.Level = 0
	for i := 0; i < 10; i++ {
		if klog.V(klog.Level(i)).Enabled() {
			ll = klog.Level(i)
		} else {
			break
		}
	}

	klog.Infof("starting driver on %s (%s) [level %d]\n\n", runtime.GOOS, runtime.GOARCH, ll)

	parts := strings.Split(bind, "://")
	if len(parts) < 2 {
		klog.Fatal("please specify a protocol in your bind URI (e.g. \"tcp://\")")
	}

	if parts[0][:4] == "unix" {
		syscall.Unlink(parts[1])
	}
	socket, err := net.Listen(parts[0], parts[1])
	if err != nil {
		klog.Fatal(err)
	}
	driver.socket = socket

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	go func() {
		_ = <-sigc
		driver.Stop()
	}()

	go func() {
		driver.exporter.ListenAndServe()
	}()

	klog.Infof("driver listening on %s\n\n", bind)
	driver.Server.Serve(socket)
}

// Stop shuts down the driver
func (driver *Driver) Stop() {
	klog.Info("gracefully stopping...")
	driver.Server.GracefulStop()
	driver.socket.Close()
	driver.exporter.Shutdown()
}
