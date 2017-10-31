package githook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/src/client"

	"gopkg.in/go-playground/webhooks.v3"
	"gopkg.in/go-playground/webhooks.v3/github"
)

// GitHookPort specifies the port the server will listen on
const GitHookPort = 999
const apiVersion = "v1"

// gitHookServer serves GetFile requests over HTTP
type gitHookServer struct {
	hook   *github.Webhook
	client *client.APIClient
}

func RunGitHookServer(address string) error {
	fmt.Printf("new in cluster\n")
	c, err := client.NewFromAddress(address)
	fmt.Printf("new in cluster err %v\n", err)
	if err != nil {
		return err
	}
	fmt.Printf("new github hook\n")
	hook := github.New(&github.Config{})
	s := &gitHookServer{
		hook,
		c,
	}
	hook.RegisterEvents(s.HandlePush, github.PushEvent)
	fmt.Printf("running github webhook\n")
	return webhooks.Run(hook, ":"+strconv.Itoa(GitHookPort), fmt.Sprintf("/%v/handle/push", apiVersion))
}

func (s *gitHookServer) findRepoByGitURL(gitURL string) string {
	repos, err := s.client.ListRepo(nil)
	if err != nil {
		fmt.Printf("error listing repo %v\n", err)
		return ""
	}
	fmt.Printf("Repos:%v\n", repos)
	fmt.Printf("looking for: %v\n", gitURL)
	for _, repo := range repos {
		fmt.Printf("comparing repo info: %v\n", repo)
		fmt.Printf("current git url: %v\n", repo.Repo.GithubURL)
		if repo.Repo.GithubURL == gitURL {
			return repo.Repo.Name
		}
	}
	return ""
}

func (s *gitHookServer) HandlePush(payload interface{}, header webhooks.Header) {

	pl := payload.(github.PushPayload)
	fmt.Printf("push payload: %v\n", pl)

	raw, err := json.Marshal(pl)
	if err != nil {
		fmt.Printf("error marshalling payload (%v): %v", pl, err)
		return
	}

	t := time.Now()
	path := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	repo := s.findRepoByGitURL(pl.Repository.SSHURL)
	fmt.Printf("got repo: %v\n", repo)
	commit, err := s.client.StartCommit(repo, getBranch(pl.Ref))
	if err != nil {
		fmt.Printf("error starting commit on repo %v: %v\n", repo, err)
		return
	}
	_, err = s.client.PutFile(repo, commit.ID, path, bytes.NewReader(raw))
	if err != nil {
		fmt.Printf("error putting file: %v\n", err)
		return
	}
	//TODO: what if the putfile fails ... what happens if we start multiple commits on head and some are open?
	// should we destroy these open commits?
	err = s.client.FinishCommit(repo, commit.ID)
	if err != nil {
		fmt.Printf("error finishing commit %v on repo %v: %v\n", commit.ID, repo, err)
	}

}

func getBranch(fullRef string) string {
	// e.g. 'refs/heads/master'
	tokens := strings.Split(fullRef, "/")
	return tokens[len(tokens)-1]
}
