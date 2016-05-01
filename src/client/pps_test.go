package client_test

import (
	"bytes"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
)

func Example_pps() {
	var client client.APIClient

	// we assume there's already a repo called "repo"
	// and that it already has some data in it
	// take a look at src/client/pfs_test.go for an example of how to get there.

	// Create a map pipeline
	if err := pps.CreatePipeline(
		client,
		"map", // the name of the pipeline
		"pachyderm/test_image", // your docker image
		[]string{"map"},        // the command run in your docker image
		nil,                    // no stdin
		0,                      // let pachyderm decide the parallelism
		[]*pps.PipelineInput{
			// map over "repo"
			pps.NewPipelineInput("repo", pps.MAP),
		},
	); err != nil {
		return // handle error
	}
	if err := pps.CreatePipeline(
		client,
		"reduce",               // the name of the pipeline
		"pachyderm/test_image", // your docker image
		[]string{"reduce"},     // the command run in your docker image
		nil,                    // no stdin
		0,                      // let pachyderm decide the parallelism
		[]*pps.PipelineInput{
			// reduce over "map"
			pps.NewPipelineInput("map", pps.REDUCE),
		},
	); err != nil {
		return // handle error
	}

	commits, err := pfs.ListCommit( // List commits that are...
		client,
		[]string{"reduce"}, // from the "reduce" repo (which the "reduce" pipeline outputs)
		nil,                // starting at the beginning of time
		pfs.READ,           // are readable
		true,               // block until commits are available
		false,              // ignore cancelled commits
	)
	if err != nil {
		return // handle error
	}
	for _, commitInfo := range commits {
		// Read output from the pipeline
		var buffer bytes.Buffer
		if err := pfs.GetFile(client, "reduce", commitInfo.Commit.ID, "file", 0, 0, "", nil, &buffer); err != nil {
			return //handle error
		}
	}
}
