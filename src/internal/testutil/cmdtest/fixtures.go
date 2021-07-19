package cmdtest

import (
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/pfs"

	"github.com/gogo/protobuf/types"
)

func RepoInfo() *pfs.RepoInfo {
	return &pfs.RepoInfo{
		Repo:        client.NewRepo("foo"),
		Created:     types.TimestampNow(),
		Description: "bar",
		Branches:    []*pfs.Branch{},
		AuthInfo: &pfs.RepoAuthInfo{
			Permissions: []auth.Permission{
				auth.Permission_REPO_READ,
				auth.Permission_REPO_WRITE,
			},
			Roles: []string{
				"roleA",
				"roleB",
			},
		},
		Details: &pfs.RepoInfo_Details{
			SizeBytes: 100,
		},
	}
}
