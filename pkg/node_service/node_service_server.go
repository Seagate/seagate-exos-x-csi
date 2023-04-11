package node_service

import (
	"context"
	"net"

	pb "github.com/Seagate/seagate-exos-x-csi/pkg/node_service/node_servicepb"
	"github.com/Seagate/seagate-exos-x-csi/pkg/storage"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

type server struct {
	pb.UnimplementedNodeServiceServer
}

func (s *server) GetInitiators(ctx context.Context, in *pb.InitiatorRequest) (*pb.Initiators, error) {
	initiators := []string{}
	var err error
	switch in.GetType() {
	case pb.InitiatorType_FC:
		initiators, err = storage.GetFCInitiators()
	case pb.InitiatorType_SAS:
		initiators, err = storage.GetSASInitiators()
	case pb.InitiatorType_ISCSI:
		initiators, err = storage.GetISCSIInitiators()
	case pb.InitiatorType_UNSPECIFIED:
		klog.InfoS("Unspecified Initiator Type in Initiator Request")
	}
	if err != nil {
		return nil, err
	}
	return &pb.Initiators{Initiators: initiators}, nil
}

func (s *server) NotifyUnmap(ctx context.Context, in *pb.UnmappedVolume) (*pb.Ack, error) {
	return &pb.Ack{Ack: 1}, nil
}

func ListenAndServe(port string) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		klog.ErrorS(err, "Node Service gRPC server failed to listen")
	}
	s := grpc.NewServer()
	pb.RegisterNodeServiceServer(s, &server{})
	klog.V(0).InfoS("Node Service gRPC server listening", "address", lis.Addr())
	s.Serve(lis)
}
