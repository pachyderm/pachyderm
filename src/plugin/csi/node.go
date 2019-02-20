package csi

import (
	"context"
	"fmt"
	"log"
	"os"
	pathlib "path"
	"sync"
	"syscall"

	"github.com/pachyderm/pachyderm/src/client"
	filesync "github.com/pachyderm/pachyderm/src/server/pkg/sync"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The parameters below are used to pass information from the Controller
// service (which has access to fields in the original PVC) to the NodeService
// (which only has access to information passed from the Controller API)
const (
	RepoParam   = "repo"
	CommitParam = "commit"
	PathParam   = "path"
)

const (
	// provisionRoot indicates the directory on the node in which this plugin
	// will store pachyderm data before exposing it to pods
	provisionRoot = "/tmp/pachyderm/"
	concurrency   = 100 // max concurrent Pachyderm download/upload operations
)

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

var (
	// _pachClient is a Pachyderm client for gRPC endpoints to share. Don't
	// access it directly, as it may still be 'nil' when the first few RPCs
	// arrive
	_pachClient    *client.APIClient
	pachClientOnce sync.Once
)

// Initialize 'pachClient' for RPC endpoints
// TODO(msteffen): handle auth-is-activated case
func getPachClient() *client.APIClient {
	pachClientOnce.Do(func() {
		var err error
		_pachClient, err = client.NewInCluster()
		if err != nil {
			panic(fmt.Sprintf("could not connect to Pachyderm: %v", err))
		}
	})
	return _pachClient
}

func init() {
	getPachClient() // start connecting to Pachd on startup
}

// NodeStageVolume implements the corresponding endpoint in the CSI service
// definition. In general, it's used to partition and format the disk and mount
// the disk on a node global directory. We use it to download Pachyderm data
// into a node global directory
func (s *NodeSvc) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	// Check arguments (including publish context parameters)
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetStagingTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}
	if len(req.PublishContext[RepoParam]) == 0 {
		return nil, status.Error(codes.InvalidArgument, "\"repo\" missing from request's publish context")
	}
	repo := req.PublishContext[RepoParam]
	if len(req.PublishContext[CommitParam]) == 0 {
		return nil, status.Error(codes.InvalidArgument, "\"commit\" missing from request's publish context")
	}
	commit := req.PublishContext[CommitParam]

	// If non-default 'path'is set, make sure it exists in Pachyderm
	c := getPachClient()
	commitPath := "/"
	if len(req.PublishContext[PathParam]) > 0 {
		commitPath = req.PublishContext[PathParam]
		if _, err := c.ListFile(repo, commit, commitPath); err != nil {
			return nil, status.Error(codes.InvalidArgument,
				fmt.Sprintf("could not stat Pachyderm data at %s@%s:%s: %v", repo, commit, commitPath, err))
		}
	}

	path := pathlib.Join(provisionRoot + req.VolumeId)
	// see if staging path exists
	var fi os.FileInfo
	for {
		var err error
		fi, err = os.Stat(path)
		if err == nil {
			// TODO(msteffen): handle case where data is partially-synced.
			if err := os.RemoveAll(path); err != nil {
				return nil, status.Error(codes.Internal,
					fmt.Sprintf("could not remove existing dir at %q: %v", path, err.Error()))
			}
			continue // retry stat()
		} else if !os.IsNotExist(err) {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	if err := os.MkdirAll(path, 0750); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !fi.IsDir() {
		return nil, status.Error(codes.Internal, fmt.Sprintf(
			"cannot stage volume at %q (%s): it is not a directory",
			path, fi.Mode()))
	}

	// sync pachyderm data to staging volume
	p := filesync.NewPuller()
	if err := p.Pull(c, path, repo, commit, commitPath,
		false /*lazy*/, false, /*empty*/
		concurrency,
		nil /*statsTree*/, "" /*statsRoot*/); err != nil {
		return nil, status.Error(codes.Internal,
			fmt.Sprintf("could not download Pachyderm data at %s@%s:%s: %v", repo, commit, path, err))
	}
	return &csi.NodeStageVolumeResponse{}, nil
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

	// Create mount point (or re-use existing mount point if it's there)
	fi, err := os.Stat(req.TargetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(req.TargetPath, 0750); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	if !fi.IsDir() {
		return nil, status.Error(codes.Internal, fmt.Sprintf("%q (%s) is not a directory", req.TargetPath, fi.Mode()))
	}

	// log mount information
	log.Printf("target %v\nfstype %v\nreadonly %v\nvolumeId %v\nattributes %v\nmountflags %v\n",
		req.TargetPath, req.VolumeCapability.GetMount().FsType, req.Readonly,
		req.VolumeId, req.VolumeContext, req.VolumeCapability.GetMount().MountFlags)

	// finally, mount 'path' to 'targetPath', exposing pach data to pod
	path := pathlib.Join(provisionRoot + req.VolumeId)
	var mntOpts uintptr = syscall.MS_BIND
	if req.Readonly {
		mntOpts |= syscall.MS_RDONLY
	}
	if err := syscall.Mount(path, req.TargetPath, "", mntOpts, ""); err != nil {
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
			// Support Stage/Unstage volume, which we use to download Pachyderm data
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
