package source

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/pachyderm/pachyderm/src/pkg/clone"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/parse"
)

type sourcer struct{}

func newSourcer() *sourcer {
	return &sourcer{}
}

func (s *sourcer) GetDirPathAndPipeline(pipelineSource *pps.PipelineSource) (string, *pps.Pipeline, error) {
	return getDirPathAndPipeline(pipelineSource)
}

func getDirPathAndPipeline(pipelineSource *pps.PipelineSource) (string, *pps.Pipeline, error) {
	if pipelineSource.GetGithubPipelineSource() != nil {
		dirPath, err := githubClone(pipelineSource.GetGithubPipelineSource())
		if err != nil {
			return "", nil, err
		}
		dirPath = filepath.Join(dirPath, pipelineSource.GetGithubPipelineSource().ContextDir)
		pipeline, err := parse.NewParser().ParsePipeline(dirPath)
		if err != nil {
			return "", nil, err
		}
		pipeline.PipelineSourceId = pipelineSource.Id
		return dirPath, pipeline, nil
	}
	return "", nil, fmt.Errorf("must specify pipeline source")
}

func githubClone(githubPipelineSource *pps.GithubPipelineSource) (string, error) {
	dirPath, err := makeTempDir()
	if err != nil {
		return "", err
	}
	if err := clone.GithubClone(
		dirPath,
		githubPipelineSource.User,
		githubPipelineSource.Repository,
		githubPipelineSource.Branch,
		githubPipelineSource.CommitId,
		githubPipelineSource.AccessToken,
	); err != nil {
		return "", err
	}
	return dirPath, nil
}

func makeTempDir() (string, error) {
	return ioutil.TempDir("", "pachyderm")
}
