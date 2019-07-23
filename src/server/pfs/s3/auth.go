package s3

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/sirupsen/logrus"
)

type authMiddleware struct {
	logger *logrus.Entry
}

func (c authMiddleware) SecretKey(r *http.Request, accessKey string, region *string) (secretKey *string, err error) {
	vars := mux.Vars(r)
	pc, err := pachClient(vars["authAccessKey"])
	if err != nil {
		return nil, err
	}

	// Some S3 clients (like minio) require the use of authenticated requests,
	// but there's no use for auth in clusters that don't have it enabled.
	// This allows the use of empty access and secret keys in the case where
	// auth is not enabled on the cluster.
	if accessKey == "" {
		active, err := pc.IsAuthActive()
		if err != nil {
			return nil, err
		}
		if !active {
			return &accessKey, nil
		}
	}

	// TODO(ys): what happens if auth fails? does this return nil, or an
	// error? currently assuming nil is returned.
	resp, err := pc.WhoAmI(pc.Ctx(), &auth.WhoAmIRequest{})
	if err != nil {
		return nil, grpcutil.ScrubGRPC(err)
	}
	if resp == nil {
		return nil, nil
	}
	return &accessKey, nil
}

func (c authMiddleware) CustomAuth(r *http.Request) (passed bool, err error) {
	pc, err := pachClient("")
	if err != nil {
		return false, err
	}

	active, err := pc.IsAuthActive()
	if err != nil {
		return false, err
	}
	return !active, nil
}
