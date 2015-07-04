package storage

// json structures that shard will return in response to requests.

type branchMsg struct {
	Name   string `json:"name"`
	TStamp string `json:"tstamp"`
}

type commitMsg struct {
	Name   string `json:"name"`
	TStamp string `json:"tstamp"`
}
