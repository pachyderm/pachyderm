package pfs

import "fmt"

const (
	// TagPrefixLength is the length of the pipeline-specific prefix
	// before each tag.
	TagPrefixLength = 4
)

// FullID prints repoName/CommitID
func (c *Commit) FullID() string {
	return fmt.Sprintf("%s/%s", c.Repo.Name, c.ID)
}
