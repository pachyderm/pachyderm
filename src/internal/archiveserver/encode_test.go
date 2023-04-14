package archiveserver

import "testing"

func TestEncodeV1(t *testing.T) {
	testData := []struct {
		name  string
		paths []string
		want  string
	}{
		{
			name: "empty",
			want: "AQ",
		},
		{
			name:  "one file",
			paths: []string{"default/images@master:/"},
			want:  "ASi1L_0EAMEAAGRlZmF1bHQvaW1hZ2VzQG1hc3RlcjovAGmFDFc",
		},
		{
			name: "doc example",
			paths: []string{
				"default/montage@master:/montage.png",
				"default/images@master:/",
			},
			want: "ASi1L_0EAK0BALQCZGVmYXVsdC9pbWFnZXNAbWFzdGVyOi8AbW9udGFnZW1vbnRhZ2UucG5nAAIQBFwMS4wBy2xQ2w",
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			got, err := EncodeV1(test.paths)
			if err != nil {
				t.Fatal(err)
			}
			if want := test.want; got != want {
				t.Errorf("EncodeV1():\n  got: %v\n want: %v", got, want)
			}
		})
	}
}
