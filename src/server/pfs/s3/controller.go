package s3

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/sirupsen/logrus"
)

type controller struct {
	pachdPort uint16

	logger *logrus.Entry

	// Name of the PFS repo holding multipart content
	repo string

	// the maximum number of allowed parts that can be associated with any
	// given file
	maxAllowedParts int

	driver Driver
}

func (c *controller) pachClient(authToken string) (*client.APIClient, error) {
	pc, err := client.NewFromAddress(fmt.Sprintf("localhost:%d", c.pachdPort))
	if err != nil {
		return nil, err
	}
	if authToken != "" {
		pc.SetAuthToken(authToken)
	}
	return pc, nil
}
