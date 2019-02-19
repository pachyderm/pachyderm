package csi

import (
	"context"
	"os"
	pathlib "path"

	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/kubernetes/pkg/util/mount"
)

// The CSI spec provides examples in which plugin names are reversed domain
// names, but more recent examples of CSI plugins use regular domain names (see
// e.g. http://pach-links.s3-website-us-west-1.amazonaws.com/csi-example)
const pluginName = "csi.pachyderm.com"

// The Pachyderm CSI plugin is versioned separately from Pachyderm itself in
// case we need to release bug fixes for the CSI plugin without releasing pachd
// itself
const pluginVersion = "v1.0"

type identitySvc struct{}

func (s *identitySvc) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	return &csi.GetPluginInfoResponse{
		Name:          pluginName,
		VendorVersion: pluginVersion,
	}, nil
}

func (s *identitySvc) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	return &csi.GetPluginCapabilitiesResponse{
		// as this CSI driver doesn't export a controller, we advertise no
		// capabilities at all
	}, nil
}

func (s *identitySvc) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	return &csi.ProbeResponse{
		Ready: &wrappers.BoolValue{Value: true},
	}, nil
}

// NodeSvc implements this CSI NodeService GRPC service interface
type NodeSvc struct {
	id string
}

// NewNodeSvc constructs a NodeSvc instance
func NewNodeSvc() *NodeSvc {
	return &NodeSvc{
		id: uuid.New(),
	}
}

// NodeStageVolume implements the corresponding endpoint in the CSI service
// definition. It's used to partition and format the disk and mount the disk on
// a node global directory
func (s *NodeSvc) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// NodePublishVolume implements the corresponding endpoint in the CSI service
// definition. It's used to bind mount the global directory on a container
// directory.
func (s *NodeSvc) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	// Check arguments
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability missing in request")
	}
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	fi, err := os.Stat(req.TargetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(targetPath, 0750); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	if !fi.IsDir() {
		return nil, status.Error(codes.Internal, fmt.Sprintf("%q (%s) is not a directory", req.TargetPath, fi.Mode()))
	}

	// deviceID := ""
	// if req.GetPublishContext() != nil {
	// 	deviceID = req.GetPublishContext()[deviceID]
	// }

	// volumeID := req.GetVolumeId()
	// attrib := req.GetVolumeContext()
	// mountFlags := req.GetVolumeCapability().GetMount().GetMountFlags()

	options := []string{"bind"}
	if req.Readonly {
		options = append(options, "ro")
	}
	mounter := mount.New("")
	path := pathlib.Join(provisionRoot + req.VolumeId)
	if err := mounter.Mount(path, targetPath, "", options); err != nil {
		return nil, err
	}

	return nil, status.Error(codes.Unimplemented, "")
}

// NodeUnpublishVolume implements the corresponding endpoint in the CSI service
// definition.
func (s *NodeSvc) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// NodeUnstageVolume implements the corresponding endpoint in the CSI service
// definition.
func (s *NodeSvc) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// NodeGetVolumeStats implements the corresponding endpoint in the CSI service
// definition.
func (s *NodeSvc) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// NodeExpandVolume implements the corresponding endpoint in the CSI service
// definition.
func (s *NodeSvc) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// NodeGetCapabilities implements the corresponding endpoint in the CSI service
// definition.
func (s *NodeSvc) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
					},
				},
			},
		},
	}, nil
}

// NodeGetInfo implements the correspoding endpoint in the CSI service
// definition.
func (s *NodeSvc) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return &csi.NodeGetInfoResponse{
		NodeId: s.id,
	}, nil
}
