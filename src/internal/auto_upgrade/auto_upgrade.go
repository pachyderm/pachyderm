package auto_upgrade

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"syscall"
	"time"
)

const timeFormat = "15:04"

type Option func(*autoUpdater)

// Register starts a goroutine which is meant to wake up once a day after the specified check time,
// and query github to see if the current commit sha of the provided tag matches the provided commit sha.
// If not, it triggers the defined signal which is then the responsibility of the caller to handle gracefully.
func Register(org, repo, tag, commit string, opts ...Option) error {
	o := autoUpdater{
		org:       org,
		repo:      repo,
		tag:       tag,
		checkTime: "01:00",
		interupt:  syscall.SIGTERM,
	}
	for _, opt := range opts {
		opt(&o)
	}

	//c := make(chan os.Signal, 1)
	//signal.Notify(c, o.interupt)
	//errors := make(chan error, 1)
	log.Printf("input time: %v", o.checkTime)
	check, err := time.Parse(timeFormat, o.checkTime)
	if err != nil {
		return err
	}
	log.Printf("parsed time: %v", check)
	// TODO: Properly handle errors here
	go func(check time.Time, o *autoUpdater) {
		for {
			// avoid's restarting rapidly during the 1min of the checkTime.
			time.Sleep(time.Minute * 1)
			noDate, err := time.Parse(timeFormat, time.Now().Format(timeFormat))
			if err != nil {
				log.Print(err)
				//	errors <- err
				return
			}
			d := diff(check.Sub(noDate))
			log.Printf("auto upgrade check waking up in %v", d)
			time.Sleep(d)
			sha, err := QueryGithubTag(o.org, o.repo, o.tag)
			log.Printf("current  github sha: %v", sha)
			log.Printf("current version sha: %v", commit)
			if err != nil {
				log.Print(err)
				//	errors <- err
				return
			}
			if strings.TrimSpace(sha) != strings.TrimSpace(commit) {
				log.Print("triggering restart...")
				syscall.Kill(syscall.Getpid(), o.interupt)
				//c <- o.interupt
			}
		}
		return
	}(check, &o)
	//err = <-errors
	//if err != nil {
	//	return err
	//}
	return nil
}

func SetCheckTime(t string) Option {
	return func(o *autoUpdater) {
		o.checkTime = t
	}
}

func SetInterrupt(i syscall.Signal) Option {
	return func(o *autoUpdater) {
		o.interupt = i
	}
}

func diff(t time.Duration) time.Duration {
	if t >= 0 {
		return t
	}
	return (time.Hour * 24) + t
}

type autoUpdater struct {
	org           string
	repo          string
	tag           string
	checkTime     string
	checkDuration string
	interupt      syscall.Signal
}

type Reference struct {
	Ref    *string    `json:"ref"`
	URL    *string    `json:"url"`
	Object *GitObject `json:"object"`
	NodeID *string    `json:"node_id,omitempty"`
}

type GitObject struct {
	Type *string `json:"type"`
	SHA  *string `json:"sha"`
	URL  *string `json:"url"`
}

type Tag struct {
	Tag          *string                `json:"tag,omitempty"`
	SHA          *string                `json:"sha,omitempty"`
	URL          *string                `json:"url,omitempty"`
	Message      *string                `json:"message,omitempty"`
	Tagger       map[string]interface{} `json:"tagger,omitempty"`
	Object       *GitObject             `json:"object,omitempty"`
	Verification map[string]interface{} `json:"verification,omitempty"`
	NodeID       *string                `json:"node_id,omitempty"`
}

// This function takes an org, repo, and tag to query a public github repo to return the current commit sha based on that tag.
func QueryGithubTag(org, repo, tag string) (string, error) {
	tagUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/ref/tags/%s", org, repo, tag)
	req, err := http.NewRequest(http.MethodGet, tagUrl, nil)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	ref := Reference{}
	err = json.NewDecoder(resp.Body).Decode(&ref)
	if err != nil {
		return "", err
	}
	if &ref == nil {
		return "", fmt.Errorf("unable to decode response A")
	}
	if ref.Object == nil {
		return "", fmt.Errorf("unable to decode response B")
	}
	if ref.Object.URL == nil {
		return "", fmt.Errorf("unable to decode response C")
	}
	req2, err := http.NewRequest(http.MethodGet, *ref.Object.URL, nil)
	if err != nil {
		return "", err
	}
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
	}
	defer resp2.Body.Close()
	ref2 := Tag{}
	json.NewDecoder(resp2.Body).Decode(&ref2)
	if ref2.Object.SHA != nil {
		return *ref2.Object.SHA, nil
	}
	return "", fmt.Errorf("unable to decode response")
}
