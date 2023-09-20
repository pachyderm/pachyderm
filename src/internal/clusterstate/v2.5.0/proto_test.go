package v2_5_0

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestRoundtripCopiedProto(t *testing.T) {
	want := &CommitInfo{
		Commit: &pfs.Commit{
			Repo: &pfs.Repo{
				Name: "test",
				Type: "user",
				Project: &pfs.Project{
					Name: "default",
				},
			},
			Id: "2",
		},
		Origin: &pfs.CommitOrigin{
			Kind: pfs.OriginKind_USER,
		},
		Description: "description",
		ParentCommit: &pfs.Commit{
			Repo: &pfs.Repo{
				Name: "test",
				Type: "user",
				Project: &pfs.Project{
					Name: "default",
				},
			},
			Id: "1",
		},
		ChildCommits: []*pfs.Commit{
			{
				Repo: &pfs.Repo{
					Name: "output1",
					Type: "user",
					Project: &pfs.Project{
						Name: "default",
					},
				},
				Id: "2",
			},
			{
				Repo: &pfs.Repo{
					Name: "output2",
					Type: "user",
					Project: &pfs.Project{
						Name: "default",
					},
				},
				Id: "2",
			},
		},
		Started:   timestamppb.Now(),
		Finished:  timestamppb.Now(),
		Finishing: timestamppb.Now(),
		Error:     "none",
		DirectProvenance: []*pfs.Branch{
			{
				Repo: &pfs.Repo{
					Name: "test",
					Type: "user",
					Project: &pfs.Project{
						Name: "default",
					},
				},
			},
		},
		SizeBytesUpperBound: 1234,
		Details: &pfs.CommitInfo_Details{
			CompactingTime: durationpb.New(time.Second),
			SizeBytes:      1234,
			ValidatingTime: durationpb.New(time.Second),
		},
	}
	bs, err := proto.Marshal(want)
	if err != nil {
		t.Fatalf("marshal original: %v", err)
	}
	got := new(CommitInfo)
	if err := proto.Unmarshal(bs, got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if diff := cmp.Diff(got, want, protocmp.Transform()); diff != "" {
		t.Errorf("round-trip (-got +want):\n%s", diff)
	}
}
