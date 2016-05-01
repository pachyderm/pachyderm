package client_test

import (
	"bytes"
	"strings"

	"github.com/pachyderm/pachyderm/src/client/pfs"
)

func Example_pfs() {
	var client APIClient
	// Create a repo called "repo"
	if err := pfs.CreateRepo(client, "repo"); err != nil {
		return // handle error
	}
	// Start a commit in our new repo on the "master" branch
	commit1, err := pfs.StartCommit(client, "repo", "", "master")
	if err != nil {
		return // handle error
	}
	// Put a file called "file" in the newly created commit with the content "foo\n".
	if _, err := pfs.PutFile(client, "repo", "master", "file", strings.NewReader("foo\n")); err != nil {
		return // handle error
	}
	// Finish the commit.
	if err := pfs.FinishCommit(client, "repo", "master"); err != nil {
		return //handle error
	}
	// Read what we wrote.
	var buffer bytes.Buffer
	if err := pfs.GetFile(client, "repo", "master", "file", 0, 0, "", nil, &buffer); err != nil {
		return //handle error
	}
	// buffer now contains "foo\n"

	// Start another commit with the previous commit as the parent.
	if _, err := pfs.StartCommit(client, "repo", "", "master"); err != nil {
		return //handle error
	}
	// Extend "file" in the newly created commit with the content "bar\n".
	if _, err := pfs.PutFile(client, "repo", "master", "file", strings.NewReader("bar\n")); err != nil {
		return // handle error
	}
	// Finish the commit.
	if err := pfs.FinishCommit(client, "repo", "master"); err != nil {
		return //handle error
	}
	// Read what we wrote.
	buffer.Reset()
	if err := pfs.GetFile(client, "repo", "master", "file", 0, 0, "", nil, &buffer); err != nil {
		return //handle error
	}
	// buffer now contains "foo\nbar\n"

	// We can still read the old version of the file though:
	buffer.Reset()
	if err := pfs.GetFile(client, "repo", commit1.ID, "file", 0, 0, "", nil, &buffer); err != nil {
		return //handle error
	}
	// buffer now contains "foo\n"

	// We can also see the Diff between the most recent commit and the first one:
	buffer.Reset()
	if err := pfs.GetFile(client, "repo", "master", "file", 0, 0, commit1.ID, nil, &buffer); err != nil {
		return //handle error
	}
}
