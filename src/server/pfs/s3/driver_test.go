package s3

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func TestBucketNameToProjectCommit(t *testing.T) {
	// pattern: [commitID.][branch. | branch.type.]repoName.projectName
	var cases = map[string]*pfs.Commit{
		"8b234298216044d4accaf3f472175a54.testLOPed701519c9c4.default": &pfs.Commit{
			ID: "8b234298216044d4accaf3f472175a54",
			Branch: &pfs.Branch{
				Name: "master",
				Repo: &pfs.Repo{
					Project: &pfs.Project{
						Name: "default",
					},
					Name: "testLOPed701519c9c4",
					Type: pfs.UserRepoType,
				},
			},
		},
		"8b234298216044d4accaf3f472175a54.notmaster.testLOPed701519c9c4.default": &pfs.Commit{
			ID: "8b234298216044d4accaf3f472175a54",
			Branch: &pfs.Branch{
				Name: "notmaster",
				Repo: &pfs.Repo{
					Project: &pfs.Project{
						Name: "default",
					},
					Name: "testLOPed701519c9c4",
					Type: pfs.UserRepoType,
				},
			},
		},
		"8b234298216044d4accaf3f472175a54.notmaster.spec.testLOPed701519c9c4.default": &pfs.Commit{
			ID: "8b234298216044d4accaf3f472175a54",
			Branch: &pfs.Branch{
				Name: "notmaster",
				Repo: &pfs.Repo{
					Project: &pfs.Project{
						Name: "default",
					},
					Name: "testLOPed701519c9c4",
					Type: pfs.SpecRepoType,
				},
			},
		},
		"notmaster.testLOPed701519c9c4.default": &pfs.Commit{
			ID: "",
			Branch: &pfs.Branch{
				Name: "notmaster",
				Repo: &pfs.Repo{
					Project: &pfs.Project{
						Name: "default",
					},
					Name: "testLOPed701519c9c4",
					Type: pfs.UserRepoType,
				},
			},
		},
		"notmaster.spec.testLOPed701519c9c4.default": &pfs.Commit{
			ID: "",
			Branch: &pfs.Branch{
				Name: "notmaster",
				Repo: &pfs.Repo{
					Project: &pfs.Project{
						Name: "default",
					},
					Name: "testLOPed701519c9c4",
					Type: pfs.SpecRepoType,
				},
			},
		},
		"testLOPed701519c9c4.default": &pfs.Commit{
			ID: "",
			Branch: &pfs.Branch{
				Name: "master",
				Repo: &pfs.Repo{
					Project: &pfs.Project{
						Name: "default",
					},
					Name: "testLOPed701519c9c4",
					Type: pfs.UserRepoType,
				},
			},
		},
	}
	for b, c := range cases {
		cc, err := bucketNameToProjectCommit(b)
		if err != nil {
			t.Error(err)
		}
		if c.ID != cc.ID {
			t.Errorf("%s: mismatched commit IDs: %s ≠ %s", b, c.ID, cc.ID)
		}
		if c.Branch.Name != cc.Branch.Name {
			t.Errorf("%s: mismatched branch names: %s ≠ %s", b, c.Branch.Name, cc.Branch.Name)
		}
		if c.Branch.Repo.Name != cc.Branch.Repo.Name {
			t.Errorf("%s: mismatched repo names: %s ≠ %s", b, c.Branch.Repo.Name, cc.Branch.Repo.Name)
		}
		if c.Branch.Repo.Type != cc.Branch.Repo.Type {
			t.Errorf("%s: mismatched repo types: %s ≠ %s", b, c.Branch.Repo.Type, cc.Branch.Repo.Type)
		}
		if c.Branch.Repo.Project.Name != cc.Branch.Repo.Project.Name {
			t.Errorf("%s: mismatched project names: %s ≠ %s", b, c.Branch.Repo.Project.Name, cc.Branch.Repo.Project.Name)
		}
	}
}
