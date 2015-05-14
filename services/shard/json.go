package main

// json.go contains json structures that shard will return in response to
// requests.

import ()

type BranchMsg struct {
	Name   string `json:"name"`
	TStamp string `json:"tstamp"`
}

type CommitMsg struct {
	Name   string `json:"name"`
	TStamp string `json:"tstamp"`
}
