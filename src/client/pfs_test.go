package client_test

import (
	"bytes"
	"strings"

	"github.com/pachyderm/pachyderm/src/client"
)

func Example_pfs() {
	c, err := client.NewFromAddress("0.0.0.0:30650")
	if err != nil {
		return // handle error
	}
	// Create a repo called "repo"
	if err := c.CreateRepo("repo"); err != nil {
		return // handle error
	}
	// Start a commit in our new repo on the "master" branch
	commit1, err := c.StartCommit("repo", "master")
	if err != nil {
		return // handle error
	}
	// Put a file called "file" in the newly created commit with the content "foo\n".
	if _, err := c.PutFile("repo", commit1.ID, "file", strings.NewReader("foo\n")); err != nil {
		return // handle error
	}
	// Finish the commit.
	if err := c.FinishCommit("repo", commit1.ID); err != nil {
		return //handle error
	}
	// Read what we wrote.
	var buffer bytes.Buffer
	if err := c.GetFile("repo", commit1.ID, "file", 0, 0, "", false, nil, &buffer); err != nil {
		return //handle error
	}
	// buffer now contains "foo\n"

	// Start another commit with the previous commit as the parent.
	commit2, err := c.StartCommit("repo", commit1.ID)
	if err != nil {
		return //handle error
	}
	// Extend "file" in the newly created commit with the content "bar\n".
	if _, err := c.PutFile("repo", commit2.ID, "file", strings.NewReader("bar\n")); err != nil {
		return // handle error
	}
	// Finish the commit.
	if err := c.FinishCommit("repo", commit2.ID); err != nil {
		return //handle error
	}
	// Read what we wrote.
	buffer.Reset()
	if err := c.GetFile("repo", commit2.ID, "file", 0, 0, "", false, nil, &buffer); err != nil {
		return //handle error
	}
	// buffer now contains "foo\nbar\n"

	// We can still read the old version of the file though:
	buffer.Reset()
	if err := c.GetFile("repo", commit1.ID, "file", 0, 0, "", false, nil, &buffer); err != nil {
		return //handle error
	}
	// buffer now contains "foo\n"

	// We can also see the Diff between the most recent commit and the first one:
	buffer.Reset()
	if err := c.GetFile("repo", commit2.ID, "file", 0, 0, commit1.ID, false, nil, &buffer); err != nil {
		return //handle error
	}
}
