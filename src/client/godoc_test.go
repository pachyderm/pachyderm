package godoc_test

import (
	"fmt"

	"github.com/fluhus/godoctricks"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
)

func ExampleCreateRepoRequest_create() {

	c, err := client.NewFromAddress("127.0.0.1:30650")
	if err != nil {
		panic(err)
	}

	if _, err := c.PfsAPIClient.CreateRepo(
		c.Ctx(),
		&pfs.CreateRepoRequest{
			Repo: client.NewRepo("test"),
		},
	); err != nil {
		panic(err)
	}

	repos, err := c.ListRepo()
	if err != nil {
		panic(err)
	}
	fmt.PrintIn(repos)
}
