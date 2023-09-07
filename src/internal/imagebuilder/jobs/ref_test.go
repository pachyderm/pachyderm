package jobs

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseRef(t *testing.T) {
	testData := []struct {
		input   string
		want    Reference
		wantErr bool
	}{
		{
			input:   "",
			wantErr: true,
		},
		{
			input: "foo#bar",
			want: NameAndPlatform{
				Name:     "foo",
				Platform: "bar",
			},
		},
		{
			input: "foo",
			want:  Name("foo"),
		},
		{
			input: "#bar",
			want:  Platform("bar"),
		},
	}

	for _, test := range testData {
		t.Run(test.input, func(t *testing.T) {
			got, err := ParseRef(test.input)
			if err != nil {
				if test.wantErr {
					return
				} else {
					t.Error(err)
				}
			} else if err == nil && test.wantErr {
				t.Errorf("expected error, but got success")
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("ref (-want +got):\n%s", diff)
			}
		})
	}
}
