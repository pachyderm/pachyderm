package pps_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	pfs "github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestPipelinePicker_UnmarshalTest(t *testing.T) {
	testData := []struct {
		name  string
		input string
		want  *pps.PipelinePicker
	}{
		{
			name:  "user repo",
			input: "default/images",
			want: &pps.PipelinePicker{
				Picker: &pps.PipelinePicker_Name{
					Name: &pps.PipelinePicker_PipelineName{
						Project: &pfs.ProjectPicker{
							Picker: &pfs.ProjectPicker_Name{
								Name: "default",
							},
						},
						Name: "images",
					},
				},
			},
		},
		{
			name:  "user repo without project",
			input: "images",
			want: &pps.PipelinePicker{
				Picker: &pps.PipelinePicker_Name{
					Name: &pps.PipelinePicker_PipelineName{
						Name:    "images",
						Project: &pfs.ProjectPicker{},
					},
				},
			},
		},
	}
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			var p pps.PipelinePicker
			if err := p.UnmarshalText([]byte(test.input)); err != nil {
				t.Fatal(err)
			}
			require.NoDiff(t, &p, test.want, []cmp.Option{protocmp.Transform()})
		})
	}
}
