package githook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
	/*
		_, err := client.NewInCluster()
		fmt.Printf("new in cluster err %v\n", err)
		if err != nil {
			return err
		}*/
	fmt.Printf("new github hook\n")
	hook := github.New(&github.Config{})
	s := &gitHookServer{
		hook,
		nil,
	}
	hook.RegisterEvents(s.HandlePush, github.PushEvent)
	fmt.Printf("running github webhook\n")
	return webhooks.Run(hook, ":"+strconv.Itoa(GitHookPort), fmt.Sprintf("/%v/handle/push", apiVersion))
}

// HandleRelease handles GitHub release events
func HandleRelease(payload interface{}, header webhooks.Header) {

	fmt.Println("Handling Release")

	pl := payload.(github.ReleasePayload)

	// only want to compile on full releases
	if pl.Release.Draft || pl.Release.Prerelease || pl.Release.TargetCommitish != "master" {
		return
	}

	// Do whatever you want from here...
	fmt.Printf("%+v", pl)
}

func (s *gitHookServer) HandlePush(payload interface{}, header webhooks.Header) {

	pl := payload.(github.PushPayload)
	fmt.Printf("push payload: %v\n", pl)

	repos, err := s.client.ListRepo(nil)
	if err != nil {
		fmt.Printf("error listing repo %v\n", err)
		return
	}
	fmt.Printf("Repos:%v\n", repos)

	raw, err := json.Marshal(pl)
	if err != nil {
		fmt.Printf("error marshalling payload (%v): %v", pl, err)
		return
	}

	t := time.Now()
	path := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d-00:00\n",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	_, err = s.client.PutFile("hook1", pl.HeadCommit.ID, path, bytes.NewReader(raw))
	if err != nil {
		fmt.Printf("error putting file: %v\n", err)
	}

	fmt.Printf("received push payload:\n%v\n", string(raw))
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	http.Error(w, "route not found", http.StatusNotFound)
}
