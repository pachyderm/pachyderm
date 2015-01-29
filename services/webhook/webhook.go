package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/bitly/go-simplejson"
)

func repoUrlToBranch(url string) string {
	return strings.Replace(strings.TrimPrefix(url, "https://github.com/"), "/", "-", 1)
}

func pushHandler(w http.ResponseWriter, r *http.Request) {
	event := r.Header.Get("X-GitHub-Event")
	if event == "ping" {
		log.Print("got ping")
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
}

func WebHookMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/push", pushHandler)

	return mux
}

func main() {
	log.SetFlags(log.Lshortfile)
	log.Print("Listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", WebHookMux()))
}
