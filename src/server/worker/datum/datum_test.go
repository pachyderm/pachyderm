package datum

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

func TestUploadOutput(t *testing.T) {
	testData := []struct {
		name         string
		files        map[string]string
		wantErrMatch string
	}{
		{
			name:  "valid filename",
			files: map[string]string{"example.txt": "hello world\n"},
		},
		{
			name:         "filename with glob",
			files:        map[string]string{"example*txt": "hello world\n"},
			wantErrMatch: "cannot upload.*globbing character",
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			d := &Datum{
				storageRoot: root,
				set: &Set{
					pfsOutputClient: client.NewNoOpModifyFileClient(),
				},
				meta: &Meta{
					Stats: &pps.ProcessStats{},
					Job:   client.NewJob(pfs.DefaultProjectName, "test", "id"),
				},
			}
			if err := os.MkdirAll(filepath.Join(root, "pfs", "out"), 0o755); err != nil {
				t.Fatal(err)
			}
			for path, content := range test.files {
				if err := os.WriteFile(filepath.Join(root, "pfs", "out", path), []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			err := d.uploadOutput()
			if err != nil {
				if test.wantErrMatch == "" {
					t.Errorf("unexpected error: %v", err)
				} else if !regexp.MustCompile(test.wantErrMatch).MatchString(err.Error()) {
					t.Errorf("regexp does not match error:\n  got: %v\n want: %v", err, test.wantErrMatch)
				}
			} else if test.wantErrMatch != "" {
				t.Errorf("expected error but got success")
			}
		})
	}
}
