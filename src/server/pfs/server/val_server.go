package server

import (
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	"golang.org/x/net/context"
)

var _ APIServer = &valAPIServer{}

type valAPIServer struct {
	APIServer
	env *serviceenv.ServiceEnv
}

func newValidated(inner APIServer, env *serviceenv.ServiceEnv) *valAPIServer {
	return &valAPIServer{
		APIServer: inner,
		env:       env,
	}
}

// InspectFileV2 returns info about a file.
func (a *valAPIServer) InspectFileV2(ctx context.Context, req *pfs.InspectFileRequest) (*pfs.FileInfoV2, error) {
	if err := validateFile(req.File); err != nil {
		return nil, err
	}
	if err := a.checkIsAuthorized(ctx, req.File.Commit.Repo, auth.Scope_READER); err != nil {
		return nil, err
	}
	return a.APIServer.InspectFileV2(ctx, req)
}

// WalkFileV2 walks over all the files under a directory, including children of children.
func (a *valAPIServer) WalkFileV2(req *pfs.WalkFileRequest, server pfs.API_WalkFileV2Server) error {
	file := req.File
	// Validate arguments
	if file == nil {
		return errors.New("file cannot be nil")
	}
	if file.Commit == nil {
		return errors.New("file commit cannot be nil")
	}
	if file.Commit.Repo == nil {
		return errors.New("file commit repo cannot be nil")
	}
	if err := a.checkIsAuthorized(server.Context(), file.Commit.Repo, auth.Scope_READER); err != nil {
		return err
	}
	return a.APIServer.WalkFileV2(req, server)
}

func (a *valAPIServer) GlobFileV2(request *pfs.GlobFileRequest, server pfs.API_GlobFileV2Server) (retErr error) {
	commit := request.Commit
	// Validate arguments
	if commit == nil {
		return errors.New("commit cannot be nil")
	}
	if commit.Repo == nil {
		return errors.New("commit repo cannot be nil")
	}
	if err := a.checkIsAuthorized(server.Context(), commit.Repo, auth.Scope_READER); err != nil {
		return err
	}
	return a.APIServer.GlobFileV2(request, server)
}

func (a *valAPIServer) ListFileV2(req *pfs.ListFileRequest, server pfs.API_ListFileV2Server) error {
	if err := validateFile(req.File); err != nil {
		return err
	}
	if err := a.checkIsAuthorized(server.Context(), req.File.Commit.Repo, auth.Scope_READER); err != nil {
		return err
	}
	return a.APIServer.ListFileV2(req, server)
}

func (a *valAPIServer) getAuth(ctx context.Context) client.AuthAPIClient {
	return a.env.GetPachClient(ctx)
}

func (a *valAPIServer) checkIsAuthorized(ctx context.Context, r *pfs.Repo, s auth.Scope) error {
	client := a.getAuth(ctx)
	me, err := client.WhoAmI(ctx, &auth.WhoAmIRequest{})
	if auth.IsErrNotActivated(err) {
		return nil
	}
	req := &auth.AuthorizeRequest{Repo: r.Name, Scope: s}
	resp, err := client.Authorize(ctx, req)
	if err != nil {
		return errors.Wrapf(grpcutil.ScrubGRPC(err), "error during authorization check for operation on \"%s\"", r.Name)
	}
	if !resp.Authorized {
		return &auth.ErrNotAuthorized{Subject: me.Username, Repo: r.Name, Required: s}
	}
	return nil
}
