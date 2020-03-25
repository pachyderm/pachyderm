package s3

import (
	"net/http"

	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
)

func (c *controller) SecretKey(r *http.Request, accessKey string, region *string) (*string, error) {
	pc, err := c.clientFactory.Client(accessKey)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create a pach client for auth")
	}

	_, err = pc.WhoAmI(pc.Ctx(), &auth.WhoAmIRequest{})
	if err != nil {
		// Some S3 clients (like minio) require the use of authenticated
		// requests, so in the case that auth is not enabled on pachyderm,
		// just allow any access credentials.
		if auth.IsErrNotActivated(err) {
			blankAccessKey := ""
			return &blankAccessKey, nil
		}
		return nil, nil
	}
	return &accessKey, nil
}

func (c *controller) CustomAuth(r *http.Request) (bool, error) {
	pc, err := c.clientFactory.Client("")
	if err != nil {
		return false, errors.Wrapf(err, "could not create a pach client for auth")
	}

	active, err := pc.IsAuthActive()
	if err != nil {
		return false, errors.Wrapf(err, "could not check whether auth is active")
	}
	return !active, nil
}
