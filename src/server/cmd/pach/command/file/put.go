package file

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/src/client"

	"github.com/urfave/cli"
)

func newPutCommand() cli.Command {
	return cli.Command{
		Name:        "put",
		Aliases:     []string{"p"},
		Usage:       "Put a file into the filesystem.",
		ArgsUsage:   "repo-name commit-id path/to/file/in/pfs",
		Description: descPut,
		Action:      actPut,
	}
}

var descPut = `put supports a number of ways to insert data into pfs:

   Put data from stdin as repo/commit/path:
   $ echo "data" | pachctl file put repo commit path

   Start a new commmit on branch, put data from stdin as repo/branch/path and
   finish the commit:
	 $ echo "data" | pachctl file put -c repo branch path

   Put a file from the local filesystem as repo/commit/path:
	 $ pachctl file put repo commit path -f file

   Put a file from the local filesystem as repo/commit/file:
	 $ pachctl file put repo commit -f file

   Put the contents of a directory as repo/commit/path/dir/file:
	 $ pachctl file put -r repo commit path -f dir

   Put the contents of a directory as repo/commit/dir/file:
	 $ pachctl file put -r repo commit -f dir

   Put the data from a URL as repo/commit/path:
	 $ pachctl file put repo commit path -f http://host/path

   Put the data from a URL as repo/commit/path:
	 $ pachctl file put repo commit -f http://host/path

   Put several files or URLs that are listed in file.
   Files and URLs should be newline delimited.
	 $ pachctl file put repo commit -i file

   Put several files or URLs that are listed at URL.
   NOTE this URL can reference local files, so it could cause you to put sensitive
   files into your Pachyderm cluster.
	 $ pachctl file put repo commit -i http://host/path`

func actPut(c *cli.Context) error {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	return clnt.CreateRepo(c.Args().First())
}

func putFileHelper(client *client.APIClient, repo, commit, path, source string, recursive bool) (retErr error) {
	if source == "-" {
		_, err := client.PutFile(repo, commit, path, os.Stdin)
		return err
	}
	// try parsing the filename as a url, if it is one do a PutFileURL
	if url, err := url.Parse(source); err == nil && url.Scheme != "" {
		return client.PutFileURL(repo, commit, path, url.String(), recursive)
	}
	if recursive {
		var eg errgroup.Group
		filepath.Walk(source, func(filePath string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			eg.Go(func() error {
				return putFileHelper(client, repo, commit, filepath.Join(path, strings.TrimPrefix(filePath, source)), filePath, false)
			})
			return nil
		})
		return eg.Wait()
	}
	f, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() {
		if err = f.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	_, err = client.PutFile(repo, commit, path, f)
	return err
}

func joinPaths(prefix, filePath string) string {
	if url, err := url.Parse(filePath); err == nil && url.Scheme != "" {
		return filepath.Join(prefix, strings.TrimPrefix(url.Path, "/"))
	}
	return filepath.Join(prefix, filePath)
}
