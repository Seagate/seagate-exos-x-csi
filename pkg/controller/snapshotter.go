package controller

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	storageapi "github.com/Seagate/seagate-exos-x-api-go"
	"github.com/Seagate/seagate-exos-x-csi/pkg/common"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
)

// CreateSnapshot creates a snapshot of the given volume
func (controller *Controller) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {

	parameters := req.GetParameters()
	name, err := common.TranslateName(req.GetName(), parameters[common.VolumePrefixKey])
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "trnaslate snapshot name contains invalid characters")
	}

	if common.ValidateName(name) == false {
		return nil, status.Error(codes.InvalidArgument, "snapshot name contains invalid characters")
	}

	if req.SourceVolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "snapshot SourceVolumeId is not valid")
	}

	_, respStatus, err := controller.client.CreateSnapshot(req.SourceVolumeId, name)
	if err != nil && respStatus.ReturnCode != snapshotAlreadyExists {
		return nil, err
	}

	response, _, err := controller.client.ShowSnapshots(name)
	if err != nil {
		return nil, err
	}

	var snapshot *csi.Snapshot
	for _, object := range response.Objects {
		if object.Typ != "snapshots" {
			continue
		}

		snapshot, err = newSnapshotFromResponse(&object)
		if err != nil {
			return nil, err
		}
	}

	if snapshot == nil {
		return nil, errors.New("snapshot not found")
	}

	if snapshot.SourceVolumeId != req.SourceVolumeId {
		return nil, status.Error(codes.AlreadyExists, "cannot validate volume with empty ID")
	}

	return &csi.CreateSnapshotResponse{Snapshot: snapshot}, nil
}

// DeleteSnapshot deletes a snapshot of the given volume
func (controller *Controller) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	_, status, err := controller.client.DeleteSnapshot(req.SnapshotId)
	if err != nil {
		if status != nil && status.ReturnCode == snapshotNotFoundErrorCode {
			klog.Infof("snapshot %s does not exist, assuming it has already been deleted", req.SnapshotId)
			return &csi.DeleteSnapshotResponse{}, nil
		}
		return nil, err
	}
	return &csi.DeleteSnapshotResponse{}, nil
}

// ListSnapshots list existing snapshots
func (controller *Controller) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	response, _, err := controller.client.ShowSnapshots()
	if err != nil {
		return nil, err
	}

	snapshots := []*csi.ListSnapshotsResponse_Entry{}
	for _, object := range response.Objects {
		if object.Typ != "snapshots" {
			continue
		}

		snapshot, err := newSnapshotFromResponse(&object)
		if err != nil {
			return nil, err
		}

		snapshots = append(snapshots, &csi.ListSnapshotsResponse_Entry{
			Snapshot: snapshot,
		})
	}

	return &csi.ListSnapshotsResponse{
		Entries: snapshots,
	}, nil
}

func newSnapshotFromResponse(object *storageapi.Object) (*csi.Snapshot, error) {
	properties, err := object.GetProperties("total-size-numeric", "name", "master-volume-name", "creation-date-time-numeric")
	if err != nil {
		return nil, fmt.Errorf("could not read snapshot %v", err)
	}

	sizeBytes, err := strconv.ParseInt(properties[0].Data, 10, 64)
	snapshotId := properties[1].Data
	sourceVolumeId := properties[2].Data
	creationTime, err := creationTimeFromString(properties[3].Data)

	return &csi.Snapshot{
		SizeBytes:      sizeBytes,
		SnapshotId:     snapshotId,
		SourceVolumeId: sourceVolumeId,
		CreationTime:   creationTime,
		ReadyToUse:     true,
	}, nil
}

func creationTimeFromString(creationTime string) (*timestamp.Timestamp, error) {
	creationTimestamp, err := strconv.ParseInt(creationTime, 10, 64)
	if err != nil {
		return nil, err
	}

	return ptypes.TimestampProto(time.Unix(creationTimestamp, 0))
}
