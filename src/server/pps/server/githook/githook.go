package githook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsconsts"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsdb"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"

	"github.com/gogo/protobuf/types"
	logrus "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"gopkg.in/go-playground/webhooks.v3"
	"gopkg.in/go-playground/webhooks.v3/github"
)

var (
	// superUserToken is the cached auth token used by PPS to write to the spec
	// repo, create pipeline subjects, and
	superUserToken string

	// superUserTokenOnce ensures that ppsToken is only read from etcd once. These are
	// read/written by apiServer#sudo()
	superUserTokenOnce sync.Once
)

// sudo is a helper function that copies 'pachClient', grants it PPS's superuser
// token, and calls 'f' with the superuser client. This helps isolate the
// githook server's use of PPS's superuser token, so that it's not widely copied
// and is unlikely to leak authority to parts of the code that aren't supposed
// to have it.
//
// Note: because the argument to 'f' is a superuser client, it should not be
// used to make any calls with unvalidated user input. Any such use could be
// exploited to make PPS a confused deputy
//
// This is a copy of an equivalent function in pps/server/api_server.go
func (s *gitHookServer) sudo(f func(*client.APIClient) error) error {
	// Get PPS auth token
	superUserTokenOnce.Do(func() {
		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = 60 * time.Second
		b.MaxInterval = 5 * time.Second
		if err := backoff.Retry(func() error {
			superUserTokenCol := col.NewCollection(s.env.GetEtcdClient(), ppsconsts.PPSTokenKey, nil, &types.StringValue{}, nil, nil).ReadOnly(context.Background())
			var result types.StringValue
			if err := superUserTokenCol.Get("", &result); err != nil {
				return fmt.Errorf("couldn't get PPS superuser token on startup")
			}
			superUserToken = result.Value
			return nil
		}, b); err != nil {
			panic("couldn't get PPS superuser token within 60s of starting up")
		}
	})

	// Copy pach client, but replace token with superUserToken
	superUserClient := s.env.GetPachClient(context.Background())
	superUserClient.SetAuthToken(superUserToken)
	return f(superUserClient)
}

// GitHookPort specifies the port the server will listen on
const GitHookPort = 999
const apiVersion = "v1"

// gitHookServer serves GetFile requests over HTTP
type gitHookServer struct {
	env       *serviceenv.ServiceEnv
	hook      *github.Webhook
	pipelines col.Collection
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
func RunGitHookServer(env *serviceenv.ServiceEnv, etcdPrefix string) error {
	hook := github.New(&github.Config{})
	s := &gitHookServer{env, hook, ppsdb.Pipelines(env.GetEtcdClient(), etcdPrefix)}
	hook.RegisterEvents(
		func(payload interface{}, header webhooks.Header) {
			if err := s.HandlePush(payload, header); err != nil {
				pl, ok := payload.(github.PushPayload)
				if !ok {
					logrus.Infof("github webhook failed to cast payload, this is likely a bug")
					logrus.Infof("github webhook failed to handle push with error %v", err)
					return
				}
				logrus.Infof("github webhook failed to handle push for repo (%v) on branch (%v) with error %v", pl.Repository.Name, path.Base(pl.Ref), err)
			}
		},
		github.PushEvent,
	)
	return webhooks.Run(hook, ":"+strconv.Itoa(GitHookPort), hookPath())
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

func (s *gitHookServer) findMatchingPipelineInputs(payload github.PushPayload) (pipelines []*pps.PipelineInfo, inputs []*pps.GitInput, err error) {
	payloadBranch := path.Base(payload.Ref)
	if err := s.sudo(func(superUserClient *client.APIClient) (err error) {
		pipelines, err = superUserClient.ListPipeline()
		return err
	}); err != nil {
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
		return nil, nil, fmt.Errorf("no pipeline inputs corresponding to git URL (%v) on branch (%v) found, perhaps the git input is not set yet on a pipeline", payload.Repository.CloneURL, payloadBranch)
	}
	return pipelines, inputs, nil
}

func (s *gitHookServer) HandlePush(payload interface{}, _ webhooks.Header) (retErr error) {
	pl, ok := payload.(github.PushPayload)
	if !ok {
		return fmt.Errorf("received invalid github.PushPayload")
	}
	logrus.Infof("received github push payload for repo (%v) on branch (%v)", pl.Repository.Name, path.Base(pl.Ref))

	raw, err := json.Marshal(pl)
	if err != nil {
		return fmt.Errorf("error marshalling payload (%v): %v", pl, err)
	}
	pipelines, gitInputs, err := s.findMatchingPipelineInputs(pl)
	if err != nil {
		return err
	}
	if pl.Repository.Private {
		for _, pipelineInfo := range pipelines {
			if err := ppsutil.FailPipeline(context.Background(), s.env.GetEtcdClient(), s.pipelines, pipelineInfo.Pipeline.Name, fmt.Sprintf("unable to clone private github repo (%v)", pl.Repository.CloneURL)); err != nil {
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
	return s.sudo(func(superUserClient *client.APIClient) (err error) {
		commit, err := superUserClient.StartCommit(repoName, branchName)
		if err != nil {
			return err
		}
		defer func() {
			if retErr != nil {
				if err := superUserClient.DeleteCommit(repoName, commit.ID); err != nil {
					logrus.Errorf("git webhook failed to delete partial commit (%v) on repo (%v) with error %v", commit.ID, repoName, err)
				}
				return
			}
			retErr = superUserClient.FinishCommit(repoName, commit.ID)
		}()
		if err := superUserClient.DeleteFile(repoName, commit.ID, "commit.json"); err != nil {
			return err
		}
		_, err = superUserClient.PutFile(repoName, commit.ID, "commit.json", bytes.NewReader(rawPayload))
		return err
	})
}
