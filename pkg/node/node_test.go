package node

import (
	"testing"

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

func TestValidateNodePublishVolumeCapabilityAllowsSingleNodeWriterMount(t *testing.T) {
	if err := validateNodePublishVolumeCapability(mountCapability(csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER)); err != nil {
		t.Fatalf("expected SINGLE_NODE_WRITER mount to be allowed: %v", err)
	}
}

func TestValidateNodePublishVolumeCapabilityAllowsLiveMigrationMultiNodeBlockAttach(t *testing.T) {
	if err := validateNodePublishVolumeCapability(blockCapability(csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER)); err != nil {
		t.Fatalf("expected live-migration multi-node block attach to be allowed: %v", err)
	}
}

func TestValidateNodePublishVolumeCapabilityRejectsLiveMigrationMountVolume(t *testing.T) {
	err := validateNodePublishVolumeCapability(mountCapability(csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER))
	if err == nil {
		t.Fatal("expected live-migration multi-node mount volume to be rejected")
	}
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("expected FailedPrecondition, got %v", status.Code(err))
	}
}
