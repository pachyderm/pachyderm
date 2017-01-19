package file

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"

	"github.com/urfave/cli"
)

func newGetCommand() cli.Command {
	return cli.Command{
		Name:        "get",
		Aliases:     []string{"g"},
		ArgsUsage:   "repo-name commit-id path/to/file",
		Usage:       "Return the contents of a file.",
		Description: "Return the contents of a file.",
		Action:      actGet,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "file, f",
				Usage: "The file to be put, it can be a local file or a URL.",
				Value: &cli.StringSlice{"-"},
			},
			cli.StringFlag{
				Name:  "input-file, i",
				Usage: "Read filepaths or URLs from a file.  If - is used, paths are read from the standard input.",
			},
			cli.BoolFlag{
				Name:  "recursive, r",
				Usage: "Recursively put the files in a directory.",
			},
			cli.BoolFlag{
				Name:  "commit, c",
				Usage: "Start and finish the commit in addition to putting data.",
			},
		},
	}
}

func actGet(c *cli.Context) (retErr error) {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	args := c.Args()
	repoName := args[0]
	commitID := args[1]
	var path string
	if len(args) == 3 {
		path = args[2]
	}
	if c.Bool("commit") {
		// We start a commit on a UUID branch and merge the commit
		// back to the main branch if PutFile was successful.
		//
		// The reason we do that is that we don't want to create a cancelled
		// commit on the main branch, which can cause future commits to be
		// cancelled as well.
		tmpCommit, err := clnt.StartCommit(repoName, uuid.NewWithoutDashes())
		if err != nil {
			return err
		}
		// Archiving the commit because we don't want this temporary commit
		// to trigger downstream pipelines.
		if err := clnt.ArchiveCommit(tmpCommit.Repo.Name, tmpCommit.ID); err != nil {
			return err
		}

		// PutFile should be operating on this temporary commit
		commitID = tmpCommit.ID
		defer func() {
			if retErr != nil {
				// something errored so we try to cancel the commit
				if err := clnt.CancelCommit(tmpCommit.Repo.Name, tmpCommit.ID); err != nil {
					fmt.Printf("Error cancelling commit: %s", err.Error())
				}
			} else {
				if err := clnt.FinishCommit(tmpCommit.Repo.Name, tmpCommit.ID); err != nil {
					retErr = err
					return
				}
				// replay the temp commit onto the main branch
				if _, err := clnt.ReplayCommit(repoName, []string{tmpCommit.ID}, args[1]); err != nil {
					retErr = err
					return
				}
			}
		}()
	}
	var sources []string
	inputFile := c.String("input-file")
	if inputFile != "" {
		var r io.Reader
		if inputFile == "-" {
			r = os.Stdin
		} else if url, err := url.Parse(inputFile); err == nil && url.Scheme != "" {
			resp, err := http.Get(url.String())
			if err != nil {
				return err
			}
			defer func() {
				if err := resp.Body.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			r = resp.Body
		} else {
			inputFile, err := os.Open(inputFile)
			if err != nil {
				return err
			}
			defer func() {
				if err := inputFile.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			r = inputFile
		}
		// scan line by line
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			if filePath := scanner.Text(); filePath != "" {
				sources = append(sources, filePath)
			}
		}
	} else {
		sources = c.StringSlice("file")
	}
	recursive := c.Bool("recursive")
	var eg errgroup.Group
	for _, source := range sources {
		source := source
		if len(args) == 2 {
			// The user has not specific a path so we use source as path.
			if source == "-" {
				return fmt.Errorf("no filename specified")
			}
			eg.Go(func() error {
				return putFileHelper(clnt, repoName, commitID, joinPaths("", source), source, recursive)
			})
		} else if len(sources) == 1 && len(args) == 3 {
			// We have a single source and the user has specified a path,
			// we use the path and ignore source (in terms of naming the file).
			eg.Go(func() error { return putFileHelper(clnt, repoName, commitID, path, source, recursive) })
		} else if len(sources) > 1 && len(args) == 3 {
			// We have multiple sources and the user has specified a path,
			// we use that path as a prefix for the filepaths.
			eg.Go(func() error {
				return putFileHelper(clnt, repoName, commitID, joinPaths(path, source), source, recursive)
			})
		}
	}
	return eg.Wait()
}
