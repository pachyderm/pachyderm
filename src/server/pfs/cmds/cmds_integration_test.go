//go:build k8s

package cmds

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"golang.org/x/sync/errgroup"
)

// TestMount creates two projects, each containing a repo with same name.  It
// mounts the repos and adds a single file to each, and verifies that the
// expected file appears in each.
func TestMount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	ctx := context.Background()
	// If the test has a deadline, cancel the context slightly before it in
	// order to allow time for clean subprocess teardown.  Without this it
	// is possible to leave filesystems mounted after test failure.
	if deadline, ok := t.Deadline(); ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(float64(time.Until(deadline))*.99))
		defer cancel()
	}
	eg, ctx := errgroup.WithContext(ctx)
	repoName := tu.UniqueString("TestMount-repo")
	configDir := t.TempDir()
	p, err := tu.NewPachctl(ctx, c, filepath.Join(configDir, "config.json"))
	require.NoError(t, err)
	defer p.Close()
	for _, projectName := range []string{tu.UniqueString("TestMount-project1"), tu.UniqueString("TestMount-project2")} {
		projectName := projectName
		mntDirPath := filepath.Join(t.TempDir())
		fileName := tu.UniqueString("filename")
		// TODO: Refactor tu.PachctlBashCmd to handle this a bit more
		// elegantly, perhaps based on a context or something like that
		// rather than on a name.  For now, though, this does work, even
		// if the indirection through subtests which always succeed but
		// spawn goroutines which may fail is a bit confusing.
		eg.Go(func() error {
			cmd, err := p.CommandTemplate(ctx, `
					pachctl create project {{.projectName}}
					pachctl create repo {{.repoName}} --project {{.projectName}}
					# this needs to be execed in order for process killing to cleanly unmount
					exec pachctl mount {{.mntDirPath}} -w --project {{.projectName}}
					`,
				map[string]string{
					"projectName": projectName,
					"repoName":    repoName,
					"mntDirPath":  mntDirPath,
				})
			if err != nil {
				return errors.Wrap(err, "could not create mount command")
			}
			if err := cmd.Run(); err != nil {
				t.Log("stdout:", cmd.Stdout())
				t.Log("stderr:", cmd.Stderr())
				return errors.Wrap(err, "could not mount")
			}
			if cmd, err = p.CommandTemplate(ctx, `
					pachctl list files {{.repoName}}@master --project {{.projectName}} | grep {{.fileName}} > /dev/null || exit "could not find {{.fileName}}"
					# check that only one file is present
					[[ $(pachctl list files {{.repoName}}@master --project {{.projectName}} | wc -l) -eq 2 ]] || exit "more than one file found in repo"
					`,
				map[string]string{
					"projectName": projectName,
					"repoName":    repoName,
					"fileName":    fileName,
				}); err != nil {
				return errors.Wrap(err, "could not create validation command")
			}
			if err := cmd.Run(); err != nil {
				t.Log("stdout:", cmd.Stdout())
				t.Log("stderr:", cmd.Stderr())
				return errors.Wrap(err, "could not validate")
			}
			return nil
		})
		eg.Go(func() error {
			if err := backoff.Retry(func() error {
				ff, err := os.ReadDir(mntDirPath)
				if err != nil {
					return errors.Wrapf(err, "could not read %s", mntDirPath)
				}
				if len(ff) == 0 {
					return errors.Errorf("%s not yet mounted", mntDirPath)
				}
				return nil
			}, backoff.NewExponentialBackOff()); err != nil {
				return errors.Wrapf(err, "%q never mounted", mntDirPath)
			}
			testFilePath := filepath.Join(mntDirPath, repoName, fileName)
			cmd, err := p.CommandTemplate(ctx, `
					echo "this is a test" > {{.testFilePath}}
					fusermount -u {{.mntDirPath}}
					`,
				map[string]string{
					"mntDirPath":   mntDirPath,
					"testFilePath": testFilePath,
				})
			if err != nil {
				return errors.Wrap(err, "could not create mutator")
			}
			if err := cmd.Run(); err != nil {
				t.Log("stdout:", cmd.Stdout())
				t.Log("stderr:", cmd.Stderr())
				return errors.Wrap(err, "could not run mutator")
			}
			return nil
		})
	}
	require.NoError(t, eg.Wait(), "goroutines failed")
}
