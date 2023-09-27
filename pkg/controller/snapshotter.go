package controller

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"

	storageapitypes "github.com/Seagate/seagate-exos-x-api-go/pkg/common"
	"github.com/Seagate/seagate-exos-x-csi/pkg/common"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

// CreateSnapshot creates a snapshot of the given volume
func (controller *Controller) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {

	parameters := req.GetParameters()
	snapshotName, err := common.TranslateName(req.GetName(), parameters[common.VolumePrefixKey])
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "translate snapshot name contains invalid characters")
	}

	if common.ValidateName(snapshotName) == false {
		return nil, status.Error(codes.InvalidArgument, "snapshot name contains invalid characters")
	}

	sourceVolumeId, err := common.VolumeIdGetName(req.GetSourceVolumeId())
	if sourceVolumeId == "" || err != nil {
		return nil, status.Error(codes.InvalidArgument, "snapshot SourceVolumeId is not valid")
	}

	respStatus, err := controller.client.CreateSnapshot(sourceVolumeId, snapshotName)
	if err != nil && respStatus.ReturnCode != storageapitypes.SnapshotAlreadyExists {
		return nil, err
	}

	// The expectation is that show snapshots will return a single array item for the snapshot created
	snapshots, _, err := controller.client.ShowSnapshots(snapshotName, "")
	if err != nil {
		return nil, err
	}

	var snapshot *csi.Snapshot
	for _, ss := range snapshots {
		if ss.ObjectName != "snapshot" {
			continue
		}

		snapshot, err = newSnapshotFromResponse(&ss)
		if err != nil {
			return nil, err
		}
	}

	if snapshot == nil {
		return nil, errors.New("snapshot not found")
	}

	if snapshot.SourceVolumeId != sourceVolumeId {
		return nil, status.Error(codes.AlreadyExists, "cannot validate volume with empty ID")
	}

	return &csi.CreateSnapshotResponse{Snapshot: snapshot}, nil
}

// DeleteSnapshot deletes a snapshot of the given volume
func (controller *Controller) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {

	if req.SnapshotId == "" {
		return nil, status.Error(codes.InvalidArgument, "DeleteSnapshot snapshot id is required")
	}

	status, err := controller.client.DeleteSnapshot(req.SnapshotId)
	if err != nil {
		if status != nil && status.ReturnCode == storageapitypes.SnapshotNotFoundErrorCode {
			klog.Infof("snapshot %s does not exist, assuming it has already been deleted", req.SnapshotId)
			return &csi.DeleteSnapshotResponse{}, nil
		}
		return nil, err
	}
	return &csi.DeleteSnapshotResponse{}, nil
}

// ListSnapshots: list existing snapshots up to MaxEntries
func (controller *Controller) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	sourceVolumeId, err := common.VolumeIdGetName(req.GetSourceVolumeId())

	response, respStatus, err := controller.client.ShowSnapshots(req.SnapshotId, sourceVolumeId)
	// BadInputParam is returned from the controller when an invalid volume is specified,
	// so return an empty response object in this case
	if err != nil {
		if respStatus.ReturnCode == storageapitypes.BadInputParam {
			return &csi.ListSnapshotsResponse{
				Entries:   []*csi.ListSnapshotsResponse_Entry{},
				NextToken: "",
			}, nil
		} else {
			return nil, err
		}
	}

	// StartingToken is an index from 1 to maximum, "" returns 0
	startingToken, err := strconv.Atoi(req.StartingToken)
	klog.V(2).Infof("ListSnapshots: MaxEntries=%v, StartingToken=%q|%d", req.MaxEntries, req.StartingToken, startingToken)

	snapshots := []*csi.ListSnapshotsResponse_Entry{}
	var count, total, next int32 = 0, 0, math.MaxInt32

	for _, object := range response {

		// Convert raw object into csi.Snapshot object
		snapshot, err := newSnapshotFromResponse(&object)

		// Only store snapshot objects
		if err == nil {
			total++
			klog.V(2).Infof("snapshot[%d]: SnapshotId=%v, SourceVolumeId=%v", total, snapshot.SnapshotId, snapshot.SourceVolumeId)

			// Filter entries if StartingToken is provided
			if (req.StartingToken == "") || (req.StartingToken != "" && total >= int32(startingToken)) {

				// Only add entries up to the maximum
				if (req.MaxEntries == 0) || (count < req.MaxEntries) {
					snapshots = append(snapshots, &csi.ListSnapshotsResponse_Entry{Snapshot: snapshot})
					count++
					klog.V(2).Infof("   added[%d]: SnapshotId=%v, SourceVolumeId=%v", count, snapshot.SnapshotId, snapshot.SourceVolumeId)
				}
				// When needed, store the next index which is returned to the caller
				if (req.MaxEntries != 0) && (count == req.MaxEntries) && (next == math.MaxInt32) {
					next = total + 1
					klog.V(2).Infof("next=%v", next)
				}
			}
		}
	}

	klog.V(2).Infof("ListSnapshots[%d]: %v", count, snapshots)

	// Mark the next token if there are snapshot entries remaining
	nextToken := ""
	if (req.MaxEntries != 0) && (next <= total) {
		nextToken = strconv.FormatInt(int64(next), 10)
		klog.V(2).Infof("next=%v, nextToken=%q", next, nextToken)
	}

	return &csi.ListSnapshotsResponse{
		Entries:   snapshots,
		NextToken: nextToken,
	}, nil
}

func newSnapshotFromResponse(snapshot *storageapitypes.SnapshotObject) (*csi.Snapshot, error) {
	if snapshot.ObjectName != "snapshot" {
		return nil, fmt.Errorf("not a snapshot object, type is %v", snapshot.ObjectName)
	}

	klog.InfoS("csi snapshot info", "snapshot", snapshot.Name, "volume", snapshot.MasterVolumeName, "creationTime", snapshot.CreationTime)

	return &csi.Snapshot{
		SizeBytes:      snapshot.TotalSizeNumeric,
		SnapshotId:     snapshot.Name,
		SourceVolumeId: snapshot.MasterVolumeName,
		CreationTime:   snapshot.CreationTime,
		ReadyToUse:     true,
	}, nil
}
