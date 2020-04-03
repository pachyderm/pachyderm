package main

// This is an OpenCV example written in Go. Run this file
// from the root of the repo. You must have a working
// Pachyderm cluster running on your machine to run this
// example.

import (
	"fmt"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"os"
)

func main() {

	// Replace the IP address with your `pachd` address.
	// If running in minikube, this will be your minikube
	// IP.
	c, err := client.NewFromAddress("192.168.64.2:30650")
	if err != nil {
		panic(err)
	}

	if _, err := c.PfsAPIClient.CreateRepo(
		c.Ctx(),
		&pfs.CreateRepoRequest{
			Repo:        client.NewRepo("images"),
			Description: "An images repo",
			Update:      true,
		},
	); err != nil {
		panic(err)
	}

	file1, err := os.Open("examples/opencv/liberty.png")
	if err != nil {
		panic(err)
	}

	if _, err := c.PutFile("images", "master", "liberty.png", file1); err != nil {
		panic(err)
	}

	file2, err := os.Open("examples/opencv/AT-AT.png")
	if err != nil {
		panic(err)
	}

	if _, err := c.PutFile("images", "master", "AT-AT.png", file2); err != nil {
		panic(err)
	}

	file3, err := os.Open("examples/opencv/kitten.png")
	if err != nil {
		panic(err)
	}

	if _, err := c.PutFile("images", "master", "kitten.png", file3); err != nil {
		panic(err)
	}

	defer func() {
		if err := c.Close(); err != nil {
			panic(err)
		}
	}()

	files, err := c.ListFile("images", "master", "/")
	if err != nil {
		panic(err)
	}
	fmt.Println(files)

	if err := c.CreatePipeline(
		"edges",
		"pachyderm/opencv",
		[]string{"python3", "/edges.py"},
		[]string{},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput("images", "/*"),
		"",
		false,
	); err != nil {
		panic(err)
	}

	if err := c.CreatePipeline(
		"montage",
		"v4tech/imagemagick",
		[]string{"sh"},
		[]string{
			"montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs -type f | sort) /pfs/out/montage.png",
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewCrossInput(
			client.NewPFSInput("images", "/"),
			client.NewPFSInput("edges", "/"),
		),
		"",
		false,
	); err != nil {
		panic(err)
	}

	pipelines, err := c.ListPipeline()
	if err != nil {
		panic(err)
	}
	fmt.Println(pipelines)
}
