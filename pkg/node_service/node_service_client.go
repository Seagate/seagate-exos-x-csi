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

// Connect to the node_service gRPC server at the given address and retrieve initiators
func GetNodeInitiators(nodeAddress string, reqType pb.InitiatorType) ([]string, error) {
	port, envFound := os.LookupEnv(common.NodeServicePortEnvVar)
	if !envFound {
		port = "978"
		klog.InfoS("no node service port found in environment. using default", "port", port)
	}

	nodeServiceAddr := nodeAddress + ":" + port
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.Dial(nodeServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		klog.ErrorS(err, "Error connecting to node service", "node ip", nodeAddress, "port", port)
		return nil, err
	}
	defer conn.Close()

	client := pb.NewNodeServiceClient(conn)
	initiatorReq := pb.InitiatorRequest{Type: reqType}
	initiators, err := client.GetInitiators(ctx, &initiatorReq)
	if err != nil {
		klog.ErrorS(err, "Error during GetInitiators", "initiatorReq", initiatorReq)
		return nil, err
	}
	return initiators.Initiators, nil
}
