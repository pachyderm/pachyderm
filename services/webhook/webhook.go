package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/bitly/go-simplejson"
	"github.com/pachyderm/pfs/lib/mapreduce"
	"github.com/pachyderm/pfs/lib/pfsclient"
)

func repoUrlToBranch(url string) string {
	return strings.Replace(strings.TrimPrefix(url, "https://github.com/"), "/", "-", 1)
}

/* pingHandler gets called when a new repo subscribes. (it doesn't actually get
 * routed to but is called by pushHandler) */
func pingHandler(w http.ResponseWriter, r *http.Request) {
	json, err := simplejson.NewFromReader(r.Body)
	if err != nil {
		log.Print("Failed to parse json: \n", err)
	}
	name, err := json.Get("repository").Get("name").String()
	if err != nil {
		log.Print("repository/name not string:\n", err)
		return
	}
	clone_url, err := json.Get("repository").Get("clone_url").String()
	if err != nil {
		log.Print("clone_url not string:\n", err)
		return
	}
	username, err := json.Get("repository").Get("owner").Get("login").String()
	if err != nil {
		log.Print("owner/login not string:\n", err)
		return
	}

	pfs := pfsclient.NewClient(os.Args[1])
	if err := pfs.Branch(name, username); err != nil {
		log.Print("Branch failed: ", err)
		return
	}

	err = pfs.MakeJob(username, "gh", mapreduce.Job{Type: "map", Input: name, Repo: clone_url, Command: []string{"/bin/map"}})
	if err != nil {
		log.Print("MakeJob failed: ", err)
		return
	}

	if err := pfs.Commit(username, "", false); err != nil {
		log.Print("Commit failed: ", err)
	}
}

func pushHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("Request to push handler.")
	event := r.Header.Get("X-GitHub-Event")
	if event == "ping" {
		pingHandler(w, r)
		return
	}
	json, err := simplejson.NewFromReader(r.Body)
	if err != nil {
		log.Print("Failed to parse json:\n", err)
		return
	}
	commit, err := json.Get("after").String()
	if err != nil {
		log.Print("Failed to get \"head\" from json:", err)
		return
	}
	// `branch` is the name of the branch for this fork of the repo
	repo, err := json.Get("repository").Get("full_name").String()
	if err != nil {
		log.Print("Failed to get \"url\" from json:", err)
	}
	branch := strings.Replace(repo, "/", "-", 1)

	pfs := pfsclient.NewClient(os.Args[1])
	if err := pfs.Commit(branch, commit, true); err != nil {
		log.Print("Commit failed: ", err)
	}
}

func WebHookMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/push", pushHandler)

	return mux
}

func main() {
	log.SetFlags(log.Lshortfile)
	log.Print("Listening on port 80...")
	log.Fatal(http.ListenAndServe(":80", WebHookMux()))
}
