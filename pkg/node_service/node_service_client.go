package node_service

import (
	"context"
	"os"
	"time"

	"github.com/Seagate/seagate-exos-x-csi/pkg/common"
	pb "github.com/Seagate/seagate-exos-x-csi/pkg/node_service/node_servicepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/klog/v2"
)

func InitializeClient(nodeAddress string) (conn *grpc.ClientConn, err error) {
	port, envFound := os.LookupEnv(common.NodeServicePortEnvVar)
	if !envFound {
		port = "978"
		klog.InfoS("no node service port found in environment. using default", "port", port)
	}
	nodeServiceAddr := nodeAddress + ":" + port
	conn, err = grpc.Dial(nodeServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		klog.ErrorS(err, "Error connecting to node service", "node ip", nodeAddress, "port", port)
		return
	}
	return
}

// Connect to the node_service gRPC server at the given address and retrieve initiators
func GetNodeInitiators(ctx context.Context, conn *grpc.ClientConn, reqType pb.InitiatorType) ([]string, error) {
	client := pb.NewNodeServiceClient(conn)
	initiatorReq := pb.InitiatorRequest{Type: reqType}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	initiators, err := client.GetInitiators(ctx, &initiatorReq)
	if err != nil {
		klog.ErrorS(err, "Error during GetInitiators", "reqType", initiatorReq.Type)
		return nil, err
	}
	return initiators.Initiators, nil
}

func NotifyUnmap(ctx context.Context, conn *grpc.ClientConn, volumeName string) (err error) {
	client := pb.NewNodeServiceClient(conn)
	unmappedVolumePb := pb.UnmappedVolume{VolumeName: volumeName}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_, err = client.NotifyUnmap(ctx, &unmappedVolumePb)
	if err != nil {
		klog.ErrorS(err, "Error during unmap notification", "unmappedVolumeName", volumeName)
	}
	return
}
