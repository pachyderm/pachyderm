package server

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/coredb"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/anypb"
)

type negativeValidator struct{}

func (negativeValidator) ValidateRepoExists(r *pfs.Repo) error {
	return pfsdb.ErrRepoNotFound{
		Project: r.GetProject().GetName(),
		Type:    r.GetType(),
		Name:    r.GetName(),
	}
}

func (negativeValidator) ValidateProjectExists(p *pfs.Project) error {
	return coredb.ErrProjectNotFound{
		Name: p.GetName(),
	}
}

func fieldViolations(vs ...*errdetails.BadRequest_FieldViolation) []*anypb.Any {
	any, err := anypb.New(&errdetails.BadRequest{
		FieldViolations: vs,
	})
	if err != nil {
		panic(err)
	}
	return []*anypb.Any{any}
}

func TestValidateRequest(t *testing.T) {
	ctx := pctx.TestContext(t)
	tx := &pachsql.Tx{}

	testData := []struct {
		name string
		req  validatable
		want *spb.Status
	}{
		{
			name: "InspectProject",
			req: &pfs.InspectProjectRequest{
				Project: &pfs.Project{
					Name: "default",
				},
			},
			want: &spb.Status{
				Code:    int32(codes.FailedPrecondition),
				Message: `validate request: field project: project "default" not found`,
				Details: fieldViolations(&errdetails.BadRequest_FieldViolation{
					Field:       "project",
					Description: `project "default" not found`,
				}),
			},
		},
		{
			name: "InspectRepo",
			req: &pfs.InspectRepoRequest{
				Repo: &pfs.Repo{
					Project: &pfs.Project{
						Name: "default",
					},
					Name: "repo",
					Type: pfs.UserRepoType,
				},
			},
			want: &spb.Status{
				Code:    int32(codes.FailedPrecondition),
				Message: `validate request: field repo.project: project "default" not found`,
				Details: fieldViolations(&errdetails.BadRequest_FieldViolation{
					Field:       "repo.project",
					Description: `project "default" not found`,
				}),
			},
		},
		{
			name: "DeleteRepos",
			req: &pfs.DeleteReposRequest{
				Projects: []*pfs.Project{
					{
						Name: "foo",
					},
					{
						Name: "bar",
					},
				},
			},
			want: &spb.Status{
				Code:    int32(codes.FailedPrecondition),
				Message: "validate request: field projects[0]: project \"foo\" not found\nfield projects[1]: project \"bar\" not found",
				Details: fieldViolations(
					&errdetails.BadRequest_FieldViolation{
						Field:       "projects[0]",
						Description: `project "foo" not found`,
					},
					&errdetails.BadRequest_FieldViolation{
						Field:       "projects[1]",
						Description: `project "bar" not found`,
					},
				),
			},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			got := validateRequest(ctx, tx, test.req, negativeValidator{}).Proto()
			if diff := cmp.Diff(got, test.want, protocmp.Transform()); diff != "" {
				t.Errorf("diff (+got -want):\n%s", diff)
			}
		})
	}
}
