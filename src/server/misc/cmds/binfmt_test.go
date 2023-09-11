package cmds

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDecodeBinfmt(t *testing.T) {
	testData := []struct {
		name        string
		fmt         string
		input, want []byte
		wantErr     bool
	}{
		{
			name:  "raw",
			fmt:   "raw",
			input: []byte("this is a test"),
			want:  []byte("this is a test"),
		},
		{
			name:  "base64",
			fmt:   "base64",
			input: []byte("dGhpcyBpcyBhIHRlc3QK"),
			want:  []byte("this is a test\n"),
		},
		{
			name:  "base64 with spaces",
			fmt:   "base64",
			input: []byte("dGhp cyBp cyBh IHRl c3QK\n"),
			want:  []byte("this is a test\n"),
		},
		{
			name:  "hex",
			fmt:   "hex",
			input: []byte("74686973206973206120746573740a"),
			want:  []byte("this is a test\n"),
		},
		{
			name:  "hex with spaces",
			fmt:   "hex",
			input: []byte("74 68 69 73 20 69 73 20 61 20 74 65 73 74 0a\n"),
			want:  []byte("this is a test\n"),
		},
		{
			name:  "hex with 0x",
			fmt:   "hex",
			input: []byte("0x74686973206973206120746573740a"),
			want:  []byte("this is a test\n"),
		},
		{
			name:  "hex with 0x and spaces",
			fmt:   "hex",
			input: []byte("0x74 68 69 73 20 69 73 20 61 20 74 65 73 74 0a\n"),
			want:  []byte("this is a test\n"),
		},
		{
			name:  "go byte slice",
			fmt:   "go",
			input: []byte(`[]byte("\x01\x02")`),
			want:  []byte{1, 2},
		},
		{
			name:  "go string",
			fmt:   "go",
			input: []byte(`"\x01\x02"`),
			want:  []byte{1, 2},
		},
		{
			name:    "invalid go (number)",
			fmt:     "go",
			input:   []byte(`123`),
			wantErr: true,
		},
		{
			name:    "invalid base64",
			fmt:     "base64",
			input:   []byte("this is not base64"),
			wantErr: true,
		},
		{
			name:    "invalid hex",
			fmt:     "hex",
			input:   []byte("this is not hex"),
			wantErr: true,
		},
		{
			name:    "invalid hex with 0x",
			fmt:     "hex",
			input:   []byte("0xthis is not hex"),
			wantErr: true,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			got, err := decodeBinfmt([]byte(test.input), test.fmt)
			if err != nil && !test.wantErr {
				t.Fatal(err)
			} else if err == nil && test.wantErr {
				t.Fatal("expected error, but got success")
			}
			if diff := cmp.Diff([]byte(test.want), got); diff != "" {
				t.Errorf("decoded binary (-want +got):\n%s", diff)
			}
		})
	}
}
