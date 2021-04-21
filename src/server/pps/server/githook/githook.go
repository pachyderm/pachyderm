// Package githook adds support for git-based sources in pipeline specs. It
// does so by exposing an HTTP server that listens for webhook requests. This
// works with github's webhook API, and anything else API-compatible with
// their push events.
package githook

// TODO(ys): remove githook server in pachyderm 2.0

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/pps"

	logrus "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"gopkg.in/go-playground/webhooks.v5/github"
)

// GitHookPort specifies the port the server will listen on
const GitHookPort = 655
const apiVersion = "v1"

// gitHookServer serves GetFile requests over HTTP
type gitHookServer struct {
	env       serviceenv.ServiceEnv
	hook      *github.Webhook
	pipelines col.PostgresCollection
}

func hookPath() string {
	return fmt.Sprintf("/%v/handle/push", apiVersion)
}

// ExternalPort provides the port used to access the service via a load
// balancer
func ExternalPort() int32 {
	return int32(31000 + GitHookPort)
}

// NodePort provides the port used to access the service via a node
func NodePort() int32 {
	return int32(30000 + GitHookPort)
}

// URLFromDomain provides the webhook URL given an input domain
func URLFromDomain(domain string) string {
	return fmt.Sprintf("http://%v:%v%v", domain, ExternalPort(), hookPath())
}

// RunGitHookServer starts the webhook server
func RunGitHookServer(env serviceenv.ServiceEnv) error {
	hook, err := github.New()
	if err != nil {
		return err
	}
	pipelines := ppsdb.Pipelines(env.GetDBClient(), env.GetPostgresListener())
	s := &gitHookServer{
		env,
		hook,
		pipelines,
	}
	return http.ListenAndServe(fmt.Sprintf(":%d", GitHookPort), s)
}

func matchingBranch(inputBranch string, payloadBranch string) bool {
	if inputBranch == payloadBranch {
		return true
	}
	if inputBranch == "" && payloadBranch == "master" {
		return true
	}
	return false
}

func (s *gitHookServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload, err := s.hook.Parse(r, github.PushEvent)
	if err != nil {
		// `ErrEventNotFound` implies github sent an event we didn't ask for
		if !errors.Is(err, github.ErrEventNotFound) {
			logrus.Errorf("error parsing github hook: %v", err)
		}
		return
	}

	pl, ok := payload.(github.PushPayload)
	if !ok {
		logrus.Errorf("github webhook failed to cast payload, this is likely a bug")
	}

	if err = s.handlePush(pl); err != nil {
		logrus.Errorf("github webhook failed to handle push for repo (%v) on branch (%v) with error %v", pl.Repository.Name, path.Base(pl.Ref), err)
	}
}

func (s *gitHookServer) findMatchingPipelineInputs(payload github.PushPayload) (pipelines []*pps.PipelineInfo, inputs []*pps.GitInput, err error) {
	payloadBranch := path.Base(payload.Ref)
	pipelines, err = s.env.GetPachClient(context.Background()).ListPipeline()
	if err != nil {
		return nil, nil, err
	}
	for _, pipelineInfo := range pipelines {
		pps.VisitInput(pipelineInfo.Input, func(input *pps.Input) {
			if input.Git != nil {
				if input.Git.URL == payload.Repository.CloneURL && matchingBranch(input.Git.Branch, payloadBranch) {
					inputs = append(inputs, input.Git)
				}
			}
		})
	}
	if len(inputs) == 0 {
		return nil, nil, errors.Errorf("no pipeline inputs corresponding to git URL (%v) on branch (%v) found, perhaps the git input is not set yet on a pipeline", payload.Repository.CloneURL, payloadBranch)
	}
	return pipelines, inputs, nil
}

func (s *gitHookServer) handlePush(pl github.PushPayload) (retErr error) {
	logrus.Infof("received github push payload for repo (%v) on branch (%v)", pl.Repository.Name, path.Base(pl.Ref))

	raw, err := json.Marshal(pl)
	if err != nil {
		return errors.Wrapf(err, "error marshalling payload (%v)", pl)
	}
	pipelines, gitInputs, err := s.findMatchingPipelineInputs(pl)
	if err != nil {
		return err
	}
	if pl.Repository.Private {
		for _, pipelineInfo := range pipelines {
			if err := ppsutil.FailPipeline(context.Background(), s.env.GetDBClient(), s.pipelines, pipelineInfo.Pipeline.Name, fmt.Sprintf("unable to clone private github repo (%v)", pl.Repository.CloneURL)); err != nil {
				// err will be handled but first we want to
				// try and fail all relevant pipelines
				logrus.Errorf("error marking pipeline %v as failed %v", pipelineInfo.Pipeline.Name, err)
				retErr = err
			}
		}
		return retErr
	}
	triggeredRepos := make(map[string]bool)
	for _, input := range gitInputs {
		if triggeredRepos[input.Name] {
			// This input is used on multiple pipelines, and we've already
			// committed to this input repo
			continue
		}
		if err := s.commitPayload(input.Name, input.Branch, raw); err != nil {
			logrus.Errorf("github webhook failed to commit payload to repo (%v) push with error: %v\n", input.Name, err)
			retErr = err
			continue
		}
		triggeredRepos[input.Name] = true
	}
	return nil
}

func (s *gitHookServer) commitPayload(repoName string, branchName string, rawPayload []byte) (retErr error) {
	client := s.env.GetPachClient(context.Background())
	commit, err := client.StartCommit(repoName, branchName)
	if err != nil {
		return err
	}
	defer func() {
		if retErr != nil {
			if err := client.SquashCommit(repoName, commit.ID); err != nil {
				logrus.Errorf("git webhook failed to delete partial commit (%v) on repo (%v) with error %v", commit.ID, repoName, err)
			}
			return
		}
		retErr = client.FinishCommit(repoName, commit.ID)
	}()
	if err = client.DeleteFile(repoName, commit.ID, "commit.json"); err != nil {
		return err
	}
	if err = client.PutFile(repoName, commit.ID, "commit.json", bytes.NewReader(rawPayload)); err != nil {
		return err
	}
	return nil
}
