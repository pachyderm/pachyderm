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

	_, err = pc.WhoAmI(pc.Ctx(), &auth.WhoAmIRequest{})
	if err != nil {
		// Some S3 clients (like minio) require the use of authenticated
		// requests, so in the case that auth is not enabled on pachyderm,
		// just allow any access credentials.

		// allow any access key when auth is disabled
		if auth.IsErrNotActivated(err) {
			blankAccessKey := ""
			return &blankAccessKey, nil
		}
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
