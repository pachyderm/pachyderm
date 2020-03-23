package client

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
)

func ExampleAPIClient_CreateRepo() {

	c, err := client.NewFromAddress("127.0.0.1:30650")
	if err != nil {
		panic(err)
	}

	if _, err := c.PfsAPIClient.CreateRepo(
		c.Ctx(),
		&pfs.CreateRepoRequest{
			Repo:        client.NewRepo("test"),
			Description: "A test repo",
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

func ExampleAPIClient_DeleteRepo() {

	c, err := client.NewFromAddress("127.0.0.1:30650")
	if err != nil {
		panic(err)
	}

	if _, err := c.PfsAPIClient.CreateRepo(
		c.Ctx(),
		&pfs.CreateRepoRequest{
			Repo:        client.NewRepo("test"),
			Description: "A test repo",
		},
	); err != nil {
		panic(err)
	}

	if _, err := c.DeleteRepo("test"); err != nil {
		panic(err)
	}
}

func ExampleAPIClient_ListRepo() {

	c, err := client.NewFromAddress("127.0.0.1:30650")
	if err != nil {
		panic(err)
	}

	if _, err := c.PfsAPIClient.CreateRepo(
		c.Ctx(),
		&pfs.CreateRepoRequest{
			Repo:        client.NewRepo("test"),
			Description: "A test repo",
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

func ExampleAPIClient_PutFile() {

	c, err := client.NewFromAddress("127.0.0.1:30650")
	if err != nil {
		panic(err)
	}

	if _, err := c.PfsAPIClient.CreateRepo(
		c.Ctx(),
		&pfs.CreateRepoRequest{
			Repo:        client.NewRepo("test"),
			Description: "A test repo",
		},
	); err != nil {
		panic(err)
	}
	w, err := c.PutFileWriter("test", "master", "file")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := w.Close(); err != nil {
			panic(err)
		}
	}()
	if _, err := w.Writer([]byte("foo\n")); err != nil {
		panic(err)
	}
}
