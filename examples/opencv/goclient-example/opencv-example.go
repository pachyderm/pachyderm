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
	//	"os"
)

func main() {

	// Replace the IP address with your `pachd` address.
	// If running in minikube, this will be your minikube
	// IP.
	c, err := client.NewFromAddress("localhost:30650")
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

	if err := c.PutFileURL("images", "master", "liberty.png", "http://imgur.com/46Q8nDz.png", false, false); err != nil {
		panic(err)
	}

	if err := c.PutFileURL("images", "master", "AT-AT.png", "http://imgur.com/8MN9Kg0.png", false, false); err != nil {
		panic(err)
	}

	if err := c.PutFileURL("images", "master", "kitten.png", "http://imgur.com/g2QnNqa.png", false, false); err != nil {
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
