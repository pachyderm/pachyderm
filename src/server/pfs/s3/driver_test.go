package s3

import (
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func TestBucketNameToCommit(t *testing.T) {
	// pattern: [commitID.][branch.]repoName[.projectName]
	var cases = map[string]*pfs.Commit{
		"testLOPed701519c9c4": {
			Id: "",
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
		"master.testLOPed701519c9c4": {
			Id: "",
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
		"notmaster.testLOPed701519c9c4": {
			Id: "",
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
		"master.testLOPed701519c9c4.default": {
			Id: "",
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
		"notmaster.testLOPed701519c9c4.default": {
			Id: "",
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
		"master.testLOPed701519c9c4.proj": {
			Id: "",
			Branch: &pfs.Branch{
				Name: "master",
				Repo: &pfs.Repo{
					Project: &pfs.Project{
						Name: "proj",
					},
					Name: "testLOPed701519c9c4",
					Type: pfs.UserRepoType,
				},
			},
		},
		"notmaster.testLOPed701519c9c4.proj": {
			Id: "",
			Branch: &pfs.Branch{
				Name: "notmaster",
				Repo: &pfs.Repo{
					Project: &pfs.Project{
						Name: "proj",
					},
					Name: "testLOPed701519c9c4",
					Type: pfs.UserRepoType,
				},
			},
		},
		"8b234298216044d4accaf3f472175a54.testLOPed701519c9c4": {
			Id: "8b234298216044d4accaf3f472175a54",
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
		"8b234298216044d4accaf3f472175a54.master.testLOPed701519c9c4": {
			Id: "8b234298216044d4accaf3f472175a54",
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
		"8b234298216044d4accaf3f472175a54.notmaster.testLOPed701519c9c4": {
			Id: "8b234298216044d4accaf3f472175a54",
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
		"8b234298216044d4accaf3f472175a54.master.testLOPed701519c9c4.default": {
			Id: "8b234298216044d4accaf3f472175a54",
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
		"8b234298216044d4accaf3f472175a54.notmaster.testLOPed701519c9c4.default": {
			Id: "8b234298216044d4accaf3f472175a54",
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
		"8b234298216044d4accaf3f472175a54.master.testLOPed701519c9c4.proj": {
			Id: "8b234298216044d4accaf3f472175a54",
			Branch: &pfs.Branch{
				Name: "master",
				Repo: &pfs.Repo{
					Project: &pfs.Project{
						Name: "proj",
					},
					Name: "testLOPed701519c9c4",
					Type: pfs.UserRepoType,
				},
			},
		},
		"8b234298216044d4accaf3f472175a54.notmaster.testLOPed701519c9c4.proj": {
			Id: "8b234298216044d4accaf3f472175a54",
			Branch: &pfs.Branch{
				Name: "notmaster",
				Repo: &pfs.Repo{
					Project: &pfs.Project{
						Name: "proj",
					},
					Name: "testLOPed701519c9c4",
					Type: pfs.UserRepoType,
				},
			},
		},
	}
	for b, c := range cases {
		cc, err := bucketNameToCommit(b)
		if err != nil {
			t.Error(err)
		}
		if c.Id != cc.Id {
			t.Errorf("%s: mismatched commit IDs: %s ≠ %s", b, c.Id, cc.Id)
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

func TestBucketNameError(t *testing.T) {
	c, err := bucketNameToCommit("7b234298216044d4accaf3f472175a54.too.many.components.bucket")
	if err == nil {
		t.Fatalf("expected error but valid result: %+v", c)
	}
	if !strings.Contains(err.Error(), "invalid bucket name") {
		t.Fatalf("expected 'invalid bucket name' error, but got: %v", err)
	}
}
