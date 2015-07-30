package source

import (
	"fmt"
	"io/ioutil"

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
	if pipelineSource.GithubPipelineSource != nil {
		dirPath, err := githubClone(pipelineSource.GithubPipelineSource)
		if err != nil {
			return "", nil, err
		}
		pipeline, err := parse.NewParser().ParsePipeline(dirPath, pipelineSource.GithubPipelineSource.ContextDir)
		if err != nil {
			return "", nil, err
		}
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
		"",
		githubPipelineSource.AccessToken,
	); err != nil {
		return "", err
	}
	return dirPath, nil
}

func makeTempDir() (string, error) {
	return ioutil.TempDir("", "pachyderm")
}
