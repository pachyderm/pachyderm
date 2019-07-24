package s3

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/sirupsen/logrus"
)

type authMiddleware struct {
	logger *logrus.Entry
}

func (c authMiddleware) SecretKey(r *http.Request, accessKey string, region *string) (secretKey *string, err error) {
	// accessKey might have a colon in it, so they're b64 encoded. Decode it
	// here.
	accessKeyDecodedBytes, err := base64.StdEncoding.DecodeString(accessKey)
	if err != nil {
		return nil, fmt.Errorf("could not decode access key: %s", err)
	}
	accessKeyDecoded := string(accessKeyDecodedBytes)

	pc, err := pachClient(accessKeyDecoded)
	if err != nil {
		return nil, fmt.Errorf("could not create a pach client for auth: %s", err)
	}

	// Some S3 clients (like minio) require the use of authenticated requests,
	// but there's no use for auth in clusters that don't have it enabled.
	// This allows the use of empty access and secret keys in the case where
	// auth is not enabled on the cluster.
	if accessKeyDecoded == "" {
		active, err := pc.IsAuthActive()
		if err != nil {
			return nil, fmt.Errorf("could not check whether auth is active: %s", err)
		}
		if !active {
			return &accessKeyDecoded, nil
		}
	}

	// TODO(ys): what happens if auth fails? does this return nil, or an
	// error? currently assuming nil is returned.
	resp, err := pc.WhoAmI(pc.Ctx(), &auth.WhoAmIRequest{})
	if err != nil {
		return nil, fmt.Errorf("could not check whoami: %s", grpcutil.ScrubGRPC(err))
	}
	if resp == nil {
		return nil, nil
	}
	return &accessKeyDecoded, nil
}

func (c authMiddleware) CustomAuth(r *http.Request) (passed bool, err error) {
	pc, err := pachClient("")
	if err != nil {
		return false, fmt.Errorf("could not create a pach client for auth: %s", err)
	}

	active, err := pc.IsAuthActive()
	if err != nil {
		return false, fmt.Errorf("could not check whether auth is active: %s", err)
	}
	return !active, nil
}
