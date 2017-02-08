package pfs

import "fmt"

func (c *Commit) FullID() string {
	return fmt.Sprintf("%s/%s", c.Repo.Name, c.ID)
}
