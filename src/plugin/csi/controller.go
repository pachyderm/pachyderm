package csi

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type controllerSvc struct{}

// CreateVolume implements the corresponding method of the CSI controller
// interface
func (s *controllerSvc) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId: uuid.New(),
			VolumeContext: map[string]string{
				RepoParam:   req.Parameters[RepoParam],
				CommitParam: req.Parameters[CommitParam],
				PathParam:   req.Parameters[PathParam],
			},
		},
	}, nil
}

// DeleteVolume implements the corresponding method of the CSI controller
// interface
func (s *controllerSvc) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	// nothing needs to be done-- this is handled by the Node service. This RPC
	// is only implemented because this controller supports the
	// CREATE_DELETE_VOLUME capability, so that we can pass information to the
	// node service via CreateVolume.
	return &csi.DeleteVolumeResponse{}, nil
}

// ControllerPublishVolume implements the corresponding method of the CSI
// controller interface
func (s *controllerSvc) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerUnpublishVolume implements the corresponding method of the CSI
// controller interface
func (s *controllerSvc) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ValidateVolumeCapabilities implements the corresponding method of the CSI
// controller interface
func (s *controllerSvc) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ListVolumes implements the corresponding method of the CSI controller
// interface
func (s *controllerSvc) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// GetCapacity implements the corresponding method of the CSI controller
// interface
func (s *controllerSvc) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerGetCapabilities implements the corresponding method of the CSI
// controller interface
func (s *controllerSvc) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: []*csi.ControllerServiceCapability{
			{
				// We need CreateVolume to pass information in the PVC to the Node
				// service, so we support that capability but nothing else
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
					},
				},
			},
		},
	}, nil
}

// ControllerExpandVolume implements the corresponding method of the CSI controller
// interface
func (s *controllerSvc) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// CreateSnapshot implements the corresponding method of the CSI controller
// interface
func (s *controllerSvc) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// DeleteSnapshot implements the corresponding method of the CSI controller
// interface
func (s *controllerSvc) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ListSnapshots implements the corresponding method of the CSI controller
// interface
func (s *controllerSvc) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
