package main

// This is an OpenCV example written in Go. Run this file
// from the root of the repo. You must have a working
// Pachyderm cluster running on your machine to run this
// example.

import (
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	//	"os"
)

func main() {

	c, err := client.NewOnUserMachine("")
	if err != nil {
		panic(err)
	}

	imageRepo := client.NewProjectRepo(pfs.DefaultProjectName, "images")
	imageCommit := imageRepo.NewCommit("master", "")

	if _, err := c.PfsAPIClient.CreateRepo(
		c.Ctx(),
		&pfs.CreateRepoRequest{
			Repo:        imageRepo,
			Description: "An images repo",
			Update:      true,
		},
	); err != nil {
		panic(err)
	}

	if err := c.PutFileURL(imageCommit, "liberty.png", "https://docs.pachyderm.com/images/opencv/liberty.jpg", false); err != nil {
		panic(err)
	}

	if err := c.PutFileURL(imageCommit, "robot.png", "https://docs.pachyderm.com/images/opencv/robot.jpg", false); err != nil {
		panic(err)
	}

	if err := c.PutFileURL(imageCommit, "kitten.png", "https://docs.pachyderm.com/images/opencv/kitten.jpg", false); err != nil {
		panic(err)
	}

	defer func() {
		if err := c.Close(); err != nil {
			panic(err)
		}
	}()

	files, err := c.ListFileAll(imageCommit, "/")
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
		client.NewProjectPFSInput(pfs.DefaultProjectName, "images", "/*"),
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
			client.NewProjectPFSInput(pfs.DefaultProjectName, "images", "/"),
			client.NewProjectPFSInput(pfs.DefaultProjectName, "edges", "/"),
		),
		"",
		false,
	); err != nil {
		panic(err)
	}

	pipelines, err := c.ListPipeline(true)
	if err != nil {
		panic(err)
	}
	fmt.Println(pipelines)
}
