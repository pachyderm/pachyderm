package s3

import (
	"fmt"
	"net/http"

	"github.com/pachyderm/pachyderm/src/client/auth"
)

func (c *controller) SecretKey(r *http.Request, accessKey string, region *string) (*string, error) {
	pc, err := c.pachClient(accessKey)
	if err != nil {
		return nil, fmt.Errorf("could not create a pach client for auth: %s", err)
	}

	// Some S3 clients (like minio) require the use of authenticated requests,
	// but there's no use for auth in clusters that don't have it enabled.
	// This allows the use of empty access and secret keys in the case where
	// auth is not enabled on the cluster.
	if accessKey == "" {
		active, err := pc.IsAuthActive()
		if err != nil {
			return nil, fmt.Errorf("could not check whether auth is active: %s", err)
		}
		if !active {
			return &accessKey, nil
		}
	}

	_, err = pc.WhoAmI(pc.Ctx(), &auth.WhoAmIRequest{})
	if err != nil {
		return nil, nil
	}
	return &accessKey, nil
}

func (c *controller) CustomAuth(r *http.Request) (bool, error) {
	pc, err := c.pachClient("")
	if err != nil {
		return false, fmt.Errorf("could not create a pach client for auth: %s", err)
	}

	active, err := pc.IsAuthActive()
	if err != nil {
		return false, fmt.Errorf("could not check whether auth is active: %s", err)
	}
	return !active, nil
}
