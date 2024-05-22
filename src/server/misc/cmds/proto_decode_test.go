package cmds

import (
	"encoding/hex"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestDecodeBinaryProto(t *testing.T) {
	testData := []struct {
		name    string
		kind    string
		input   string
		want    proto.Message
		wantErr bool
	}{
		{
			name:    "invalid data",
			kind:    "admin_v2.InspectClusterRequest",
			input:   "41414141",
			wantErr: true,
		},
		{
			name:  "empty",
			kind:  "google.protobuf.Empty",
			input: "",
			want:  &emptypb.Empty{},
		},
		{
			name: "BranchInfo",
			kind: "pfs_v2.BranchInfo",
			input: "0a1f0a170a04746573741204757365721a09" +
				"0a0764656661756c7412046d61696e125c0a" +
				"1f0a170a04746573741204757365721a090a" +
				"0764656661756c7412046d61696e12203434" +
				"343434343434343434343434343434343434" +
				"3434343434343434343434341a170a047465" +
				"73741204757365721a090a0764656661756c74",
			want: &pfs.BranchInfo{
				Branch: &pfs.Branch{
					Repo: &pfs.Repo{
						Name: "test",
						Type: pfs.UserRepoType,
						Project: &pfs.Project{
							Name: "default",
						},
					},
					Name: "main",
				},
				Head: &pfs.Commit{
					Repo: &pfs.Repo{
						Name: "test",
						Type: pfs.UserRepoType,
						Project: &pfs.Project{
							Name: "default",
						},
					},
					Id: "44444444444444444444444444444444",
					Branch: &pfs.Branch{
						Repo: &pfs.Repo{
							Name: "test",
							Type: pfs.UserRepoType,
							Project: &pfs.Project{
								Name: "default",
							},
						},
						Name: "main",
					},
				},
			},
		},
	}
	all := allProtoMessages()
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			b, err := hex.DecodeString(test.input)
			if err != nil {
				t.Fatalf("decode test data: %v", err)
			}
			md, ok := all[test.kind]
			if !ok {
				t.Fatalf("unknown message kind: %v", test.kind)
			}
			got, err := decodeBinaryProto(md, b)
			if err != nil && !test.wantErr {
				t.Fatalf("decode: %v", err)
			} else if err == nil && test.wantErr {
				t.Fatalf("expected decoding error, but got success")
			}
			if diff := cmp.Diff(test.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("decoded message (-want +got):\n%s", diff)
			}
		})
	}
}
