package controller

import (
	"context"
	"testing"

	"github.com/Seagate/seagate-exos-x-csi/pkg/common"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mountCapability(mode csi.VolumeCapability_AccessMode_Mode) *csi.VolumeCapability {
	return &csi.VolumeCapability{
		AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}},
		AccessMode: &csi.VolumeCapability_AccessMode{Mode: mode},
	}
}

func blockCapability(mode csi.VolumeCapability_AccessMode_Mode) *csi.VolumeCapability {
	return &csi.VolumeCapability{
		AccessType: &csi.VolumeCapability_Block{Block: &csi.VolumeCapability_BlockVolume{}},
		AccessMode: &csi.VolumeCapability_AccessMode{Mode: mode},
	}
}

func TestValidateAccessModesAcceptsSingleNodeWriterAndLiveMigrationBlockAttach(t *testing.T) {
	caps := []*csi.VolumeCapability{
		mountCapability(csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER),
		blockCapability(csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER),
	}

	if err := validateAccessModes(caps); err != nil {
		t.Fatalf("validateAccessModes returned error for SINGLE_NODE_WRITER and live-migration block attach modes: %v", err)
	}
}

func TestIsValidVolumeCapabilitiesRejectsUnsupportedMultiNodeMode(t *testing.T) {
	caps := []*csi.VolumeCapability{
		mountCapability(csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY),
	}

	err := isValidVolumeCapabilities(caps)
	if err == nil {
		t.Fatal("expected unsupported multi-node access mode to be rejected")
	}
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("expected FailedPrecondition, got %v", status.Code(err))
	}
}

func TestValidateVolumeCapabilitiesRejectsUnsupportedMultiNodeModeBeforeLookup(t *testing.T) {
	controller := &Controller{}
	req := &csi.ValidateVolumeCapabilitiesRequest{
		VolumeId:           common.VolumeIdAugment("vol-1", common.StorageProtocolISCSI, "wwn-1"),
		VolumeCapabilities: []*csi.VolumeCapability{mountCapability(csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY)},
	}

	_, err := controller.ValidateVolumeCapabilities(context.Background(), req)
	if err == nil {
		t.Fatal("expected unsupported multi-node access mode to be rejected")
	}
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("expected FailedPrecondition, got %v", status.Code(err))
	}
}
