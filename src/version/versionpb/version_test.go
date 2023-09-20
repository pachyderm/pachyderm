package versionpb

import (
	"errors"
	"testing"
)

func TestCanonical(t *testing.T) {
	testData := []struct {
		input *Version
		want  string
	}{
		{
			input: &Version{},
			want:  "v0.0.0",
		},
		{
			input: &Version{Major: 0, Minor: 0, Micro: 0, Additional: "+12345"}, // local build,
			want:  "v0.0.0-12345",
		},
		{
			input: &Version{Major: 0, Minor: 0, Micro: 0},
			want:  "v0.0.0",
		},
		{
			input: &Version{Major: 2, Minor: 4, Micro: 3},
			want:  "v2.4.3",
		},
		{
			input: &Version{Major: 2, Minor: 4, Micro: 3, Additional: "-nightly.20221221"},
			want:  "v2.4.3-nightly.20221221",
		},
	}

	for _, test := range testData {
		got := test.input.Canonical()
		if want := test.want; got != want {
			t.Errorf("Canonical(%s):\n  got: %v\n want: %v", test.input.String(), got, want)
		}
	}
}

func TestCompare(t *testing.T) {
	testData := []struct {
		client, server *Version
		wantErr        error
	}{
		{
			wantErr: ErrClientIsNil,
		},
		{
			// This mirrors the development environment.
			client: new(Version),
			server: new(Version),
		},
		{
			client: &Version{Major: 2, Minor: 5},
			server: &Version{Major: 0, Minor: 0, Micro: 0},
		},
		{
			client: &Version{Major: 2, Minor: 5},
			server: &Version{Major: 0, Minor: 0, Micro: 0, Additional: "+12345"}, // local build
		},
		{
			client: &Version{Major: 2, Minor: 5, Micro: 1},
			server: &Version{Major: 2, Minor: 5, Micro: 2},
		},
		{
			client: &Version{Major: 2, Minor: 5, Micro: 2},
			server: &Version{Major: 2, Minor: 5, Micro: 1},
		},
		{
			client: &Version{Major: 2, Minor: 5, Micro: 3},
			server: &Version{Major: 2, Minor: 5, Micro: 3},
		},
		{
			client:  &Version{Major: 2, Minor: 4, Micro: 3},
			server:  &Version{Major: 2, Minor: 5, Micro: 0},
			wantErr: ErrClientTooOld,
		},
		{
			client:  &Version{Major: 2, Minor: 5, Micro: 3},
			server:  &Version{Major: 2, Minor: 4, Micro: 0},
			wantErr: ErrServerTooOld,
		},
		{
			client:  &Version{Major: 2, Minor: 5, Micro: 0, Additional: "-nightly.20230101"},
			server:  &Version{Major: 2, Minor: 5, Micro: 0},
			wantErr: ErrIncompatiblePreview,
		},
		{
			client:  &Version{Major: 2, Minor: 5, Micro: 0},
			server:  &Version{Major: 2, Minor: 5, Micro: 0, Additional: "-nightly.20230101"},
			wantErr: ErrIncompatiblePreview,
		},
		{
			client:  &Version{Major: 2, Minor: 5, Micro: 0, Additional: "-nightly.20230101"},
			server:  &Version{Major: 2, Minor: 5, Micro: 0, Additional: "-nightly.20230102"},
			wantErr: ErrIncompatiblePreview,
		},
		{
			client: &Version{Major: 2, Minor: 5, Micro: 0, Additional: "-nightly.20230102"},
			server: &Version{Major: 2, Minor: 5, Micro: 0, Additional: "-nightly.20230102"},
		},
	}

	for _, test := range testData {
		err := IsCompatible(test.client, test.server)
		if !errors.Is(err, test.wantErr) {
			t.Errorf("IsCompatible(client=%s, server=%s):\n  got: %v\n want: %v", test.client.Canonical(), test.server.Canonical(), err, test.wantErr)
		}
	}
}
