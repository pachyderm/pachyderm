package githook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pps"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsdb"
	"github.com/pachyderm/pachyderm/src/server/pkg/util"

	etcd "github.com/coreos/etcd/clientv3"
	"golang.org/x/net/context"
	"gopkg.in/go-playground/webhooks.v3"
	"gopkg.in/go-playground/webhooks.v3/github"
)

// GitHookPort specifies the port the server will listen on
const GitHookPort = 999
const apiVersion = "v1"

// gitHookServer serves GetFile requests over HTTP
type gitHookServer struct {
	hook       *github.Webhook
	client     *client.APIClient
	etcdClient *etcd.Client
	pipelines  col.Collection
}

func RunGitHookServer(address string, etcdAddress string, etcdPrefix string) error {
	fmt.Printf("new in cluster\n")
	c, err := client.NewFromAddress(address)
	fmt.Printf("new in cluster err %v\n", err)
	if err != nil {
		return err
	}
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{etcdAddress},
		DialOptions: client.EtcdDialOptions(),
	})
	if err != nil {
		return err
	}
	fmt.Printf("new github hook\n")
	hook := github.New(&github.Config{})
	s := &gitHookServer{
		hook,
		c,
		etcdClient,
		ppsdb.Pipelines(etcdClient, etcdPrefix),
	}
	hook.RegisterEvents(s.HandlePush, github.PushEvent)
	fmt.Printf("running github webhook\n")
	return webhooks.Run(hook, ":"+strconv.Itoa(GitHookPort), fmt.Sprintf("/%v/handle/push", apiVersion))
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

func (s *gitHookServer) findMatchingPipelineInputs(payload github.PushPayload) (pipelines []*pps.PipelineInfo, inputs []*pps.GithubInput, err error) {
	urls := []string{
		payload.Repository.SSHURL,
		payload.Repository.GitURL,
		payload.Repository.CloneURL,
		payload.Repository.SvnURL,
	}
	var walkInput func(*pps.Input)
	payloadBranch := getBranch(payload.Ref)
	// TODO use this instead:
	//pps.VisitInput(pipelineInfo.Input, func(input *pps.Input) {
	walkInput = func(input *pps.Input) {
		if input.Github != nil {
			for _, url := range urls {
				fmt.Printf("compariny input url (%v) to push event (%v)\n", input.Github.URL, url)
				fmt.Printf("pipeline branch (%v), payload branch (%v), match? (%v)\n", input.Github.Branch, payloadBranch, matchingBranch(input.Github.Branch, payloadBranch))
				if input.Github.URL == url && matchingBranch(input.Github.Branch, payloadBranch) {
					//			repoName = pps.RepoNameFromGithubInfo(input.Github.URL, input.Github.Name)
					inputs = append(inputs, input.Github)
				}
			}
		}
		var inputs []*pps.Input
		if input.Cross != nil {
			inputs = input.Cross
		}
		if input.Union != nil {
			inputs = input.Union
		}
		for _, input := range inputs {
			walkInput(input)
		}
	}
	pipelines, err = s.client.ListPipeline()
	if err != nil {
		return nil, nil, err
	}
	for _, pipelineInfo := range pipelines {
		walkInput(pipelineInfo.Input)
	}
	if len(inputs) == 0 {
		return nil, nil, fmt.Errorf("no pipeline inputs corresponding to github URL (%v) on branch (%v) found, perhaps the github input is not set yet on a pipeline", urls[0], payloadBranch)
	}
	return pipelines, inputs, nil
}

func (s *gitHookServer) HandlePush(payload interface{}, header webhooks.Header) {
	var err error
	defer func() {
		// Handle any return error
		if err != nil {
			//TODO: This should probably be a logger of some sort
			// so that we emit the error in the right format for a log parser
			fmt.Printf("Github Webhook failed to handle push with error: %v\n", err)
		}
	}()

	pl := payload.(github.PushPayload)
	fmt.Printf("push payload: %v\n", pl)

	raw, err := json.Marshal(pl)
	if err != nil {
		err = fmt.Errorf("error marshalling payload (%v): %v", pl, err)
		return
	}
	pipelines, githubInputs, err := s.findMatchingPipelineInputs(pl)
	if err != nil {
		return
	}
	if pl.Repository.Private {
		var loopErr error
		for _, pipelineInfo := range pipelines {
			someErr := util.FailPipeline(context.Background(), s.etcdClient, s.pipelines, pipelineInfo.Pipeline.Name, fmt.Sprintf("unable to clone private github repo (%v)", pl.Repository.CloneURL))
			// Err will be handled by final defer function, but first we want to
			// try and fail all relevant pipelines
			if someErr != nil {
				// Only set loopErr if someErr isn't nil
				loopErr = someErr
			}
		}
		err = loopErr
		return
	}
	triggeredRepos := make(map[string]bool)
	for _, input := range githubInputs {
		repoName := pps.RepoNameFromGithubInfo(input.URL, input.Name)
		if alreadyTriggered := triggeredRepos[repoName]; alreadyTriggered {
			// This input is used on multiple pipelines, and we've already
			// committed to this input repo
			continue
		}
		fmt.Printf("got repo: %v\n", repoName)
		branchName := "master"
		if input.Branch != "" {
			branchName = input.Branch
		}
		commit, err := s.client.StartCommit(repoName, branchName)
		if err != nil {
			fmt.Printf("error starting commit on repo %v: %v\n", repoName, err)
			return
		}
		func() {
			defer func() {
				// Handle any return error
				if err != nil {
					//TODO: This should probably be a logger of some sort
					// so that we emit the error in the right format for a log parser
					fmt.Printf("Github Webhook failed to handle push with error: %v\n", err)
				}
				triggeredRepos[repoName] = true
			}()
			defer func() {
				if err != nil {
					err = s.client.DeleteCommit(repoName, commit.ID)
					return
				}
				err = s.client.FinishCommit(repoName, commit.ID)
				//Final deferred function deals w non nil error
			}()
			_, err = s.client.PutFile(repoName, commit.ID, "commit.json", bytes.NewReader(raw))
			if err != nil {
				return
			}
		}()
	}
}

func getBranch(fullRef string) string {
	// e.g. 'refs/heads/master'
	tokens := strings.Split(fullRef, "/")
	return tokens[len(tokens)-1]
}
