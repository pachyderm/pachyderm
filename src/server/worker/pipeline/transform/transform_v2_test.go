package transform

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	pfstesting "github.com/pachyderm/pachyderm/src/server/pfs/server/testing"
	"github.com/pachyderm/pachyderm/src/server/pkg/tar"
	"github.com/pachyderm/pachyderm/src/server/pkg/tarutil"
)

func TestJobSuccessV2(t *testing.T) {
	V2Test(t, func(pi *pps.PipelineInfo, env *testEnv) error {
		ctx, etcdJobInfo := mockBasicJob(t, env, pi)
		tarFiles := []tarutil.File{
			tarutil.NewMemFile("/file", []byte("foobar")),
		}
		triggerJobV2(t, env, pi, tarFiles)
		ctx = withTimeout(ctx, 10*time.Second)
		<-ctx.Done()
		require.Equal(t, pps.JobState_JOB_SUCCESS, etcdJobInfo.State)

		// Ensure the output commit is successful
		outputCommitID := etcdJobInfo.OutputCommit.ID
		outputCommitInfo, err := env.PachClient.InspectCommit(pi.Pipeline.Name, outputCommitID)
		require.NoError(t, err)
		require.NotNil(t, outputCommitInfo.Finished)

		branchInfo, err := env.PachClient.InspectBranch(pi.Pipeline.Name, pi.OutputBranch)
		require.NoError(t, err)
		require.NotNil(t, branchInfo)

		r, err := env.PachClient.GetTarV2(pi.Pipeline.Name, outputCommitID, "/*")
		require.NoError(t, err)
		require.NoError(t, tarutil.Iterate(r, func(file tarutil.File) error {
			ok, err := tarutil.Equal(tarFiles[0], file)
			require.NoError(t, err)
			require.True(t, ok)
			tarFiles = tarFiles[1:]
			return nil
		}))
		return nil
	})
}

func V2Test(t *testing.T, cb func(*pps.PipelineInfo, *testEnv) error) {
	if os.Getenv("CI") == "true" {
		t.SkipNow()
	}
	pi := defaultPipelineInfo()
	pi.EnableStats = true
	require.NoError(t, withWorkerSpawnerPair(pi, func(env *testEnv) error {
		return cb(pi, env)
	}, pfstesting.NewPachdConfig()))
}

func triggerJobV2(t *testing.T, env *testEnv, pi *pps.PipelineInfo, files []tarutil.File) {
	commit, err := env.PachClient.StartCommit(pi.Input.Pfs.Repo, "master")
	require.NoError(t, err)
	buf := &bytes.Buffer{}
	require.NoError(t, tarutil.WithWriter(buf, func(tw *tar.Writer) error {
		for _, f := range files {
			if err := tarutil.WriteFile(tw, f); err != nil {
				return err
			}
		}
		return nil
	}))
	require.NoError(t, env.PachClient.PutTarV2(pi.Input.Pfs.Repo, commit.ID, buf, false))
	require.NoError(t, env.PachClient.FinishCommit(pi.Input.Pfs.Repo, commit.ID))
}

func TestJobFailedDatumV2(t *testing.T) {
	V2Test(t, func(pi *pps.PipelineInfo, env *testEnv) error {
		pi.Transform.Cmd = []string{"bash", "-c", "(exit 1)"}
		ctx, etcdJobInfo := mockBasicJob(t, env, pi)
		tarFiles := []tarutil.File{
			tarutil.NewMemFile("/file", []byte("foobar")),
		}
		triggerJobV2(t, env, pi, tarFiles)
		ctx = withTimeout(ctx, 10*time.Second)
		<-ctx.Done()
		require.Equal(t, pps.JobState_JOB_FAILURE, etcdJobInfo.State)
		// TODO: check job stats
		return nil
	})
}

func TestJobMultiDatumV2(t *testing.T) {
	V2Test(t, func(pi *pps.PipelineInfo, env *testEnv) error {
		ctx, etcdJobInfo := mockBasicJob(t, env, pi)
		tarFiles := []tarutil.File{
			tarutil.NewMemFile("/a", []byte("foobar")),
			tarutil.NewMemFile("/b", []byte("barfoo")),
		}
		triggerJobV2(t, env, pi, tarFiles)
		ctx = withTimeout(ctx, 10*time.Second)
		<-ctx.Done()
		require.Equal(t, pps.JobState_JOB_SUCCESS, etcdJobInfo.State)

		// Ensure the output commit is successful.
		outputCommitID := etcdJobInfo.OutputCommit.ID
		outputCommitInfo, err := env.PachClient.InspectCommit(pi.Pipeline.Name, outputCommitID)
		require.NoError(t, err)
		require.NotNil(t, outputCommitInfo.Finished)

		branchInfo, err := env.PachClient.InspectBranch(pi.Pipeline.Name, pi.OutputBranch)
		require.NoError(t, err)
		require.NotNil(t, branchInfo)

		// Get the output files.
		r, err := env.PachClient.GetTarV2(pi.Pipeline.Name, outputCommitID, "/*")
		require.NoError(t, err)
		require.NoError(t, tarutil.Iterate(r, func(file tarutil.File) error {
			ok, err := tarutil.Equal(tarFiles[0], file)
			require.NoError(t, err)
			require.True(t, ok)
			tarFiles = tarFiles[1:]
			return nil
		}))
		return nil
	})
}

func TestJobSerialV2(t *testing.T) {
	V2Test(t, func(pi *pps.PipelineInfo, env *testEnv) error {
		ctx, etcdJobInfo := mockBasicJob(t, env, pi)
		tarFiles := []tarutil.File{
			tarutil.NewMemFile("/a", []byte("foobar")),
			tarutil.NewMemFile("/b", []byte("barfoo")),
		}
		triggerJobV2(t, env, pi, tarFiles[:1])
		ctx = withTimeout(ctx, 10*time.Second)
		<-ctx.Done()
		require.Equal(t, pps.JobState_JOB_SUCCESS, etcdJobInfo.State)

		ctx, etcdJobInfo = mockBasicJob(t, env, pi)
		triggerJobV2(t, env, pi, tarFiles[1:])
		ctx = withTimeout(ctx, 10*time.Second)
		<-ctx.Done()
		require.Equal(t, pps.JobState_JOB_SUCCESS, etcdJobInfo.State)

		// Ensure the output commit is successful
		outputCommitID := etcdJobInfo.OutputCommit.ID
		outputCommitInfo, err := env.PachClient.InspectCommit(pi.Pipeline.Name, outputCommitID)
		require.NoError(t, err)
		require.NotNil(t, outputCommitInfo.Finished)

		branchInfo, err := env.PachClient.InspectBranch(pi.Pipeline.Name, pi.OutputBranch)
		require.NoError(t, err)
		require.NotNil(t, branchInfo)

		// Get the output files.
		r, err := env.PachClient.GetTarV2(pi.Pipeline.Name, outputCommitID, "/*")
		require.NoError(t, err)
		require.NoError(t, tarutil.Iterate(r, func(file tarutil.File) error {
			ok, err := tarutil.Equal(tarFiles[0], file)
			require.NoError(t, err)
			require.True(t, ok)
			tarFiles = tarFiles[1:]
			return nil
		}))
		return nil
	})
}

func TestJobSerialDeleteV2(t *testing.T) {
	V2Test(t, func(pi *pps.PipelineInfo, env *testEnv) error {
		ctx, etcdJobInfo := mockBasicJob(t, env, pi)
		tarFiles := []tarutil.File{
			tarutil.NewMemFile("/a", []byte("foobar")),
			tarutil.NewMemFile("/b", []byte("barfoo")),
		}
		triggerJobV2(t, env, pi, tarFiles[:1])
		ctx = withTimeout(ctx, 10*time.Second)
		<-ctx.Done()
		require.Equal(t, pps.JobState_JOB_SUCCESS, etcdJobInfo.State)

		ctx, etcdJobInfo = mockBasicJob(t, env, pi)
		triggerJobV2(t, env, pi, tarFiles[1:])
		ctx = withTimeout(ctx, 10*time.Second)
		<-ctx.Done()
		require.Equal(t, pps.JobState_JOB_SUCCESS, etcdJobInfo.State)

		ctx, etcdJobInfo = mockBasicJob(t, env, pi)
		deleteFilesV2(t, env, pi, []string{"/a"})
		ctx = withTimeout(ctx, 10*time.Second)
		<-ctx.Done()
		require.Equal(t, pps.JobState_JOB_SUCCESS, etcdJobInfo.State)

		// Ensure the output commit is successful
		outputCommitID := etcdJobInfo.OutputCommit.ID
		outputCommitInfo, err := env.PachClient.InspectCommit(pi.Pipeline.Name, outputCommitID)
		require.NoError(t, err)
		require.NotNil(t, outputCommitInfo.Finished)

		branchInfo, err := env.PachClient.InspectBranch(pi.Pipeline.Name, pi.OutputBranch)
		require.NoError(t, err)
		require.NotNil(t, branchInfo)

		// Get the output files.
		r, err := env.PachClient.GetTarV2(pi.Pipeline.Name, outputCommitID, "/*")
		require.NoError(t, err)
		require.NoError(t, tarutil.Iterate(r, func(file tarutil.File) error {
			ok, err := tarutil.Equal(tarFiles[1], file)
			require.NoError(t, err)
			require.True(t, ok)
			return nil
		}))
		return nil
	})
}

func deleteFilesV2(t *testing.T, env *testEnv, pi *pps.PipelineInfo, files []string) {
	commit, err := env.PachClient.StartCommit(pi.Input.Pfs.Repo, "master")
	require.NoError(t, err)
	require.NoError(t, env.PachClient.DeleteFilesV2(pi.Input.Pfs.Repo, commit.ID, files))
	require.NoError(t, env.PachClient.FinishCommit(pi.Input.Pfs.Repo, commit.ID))
}
