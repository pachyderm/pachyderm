package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func check(err error, t *testing.T) {
	if err != nil {
		t.Fatal(err)
	}
}

func checkBody(res *http.Response, expected string, t *testing.T) {
	value, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(value) != expected {
		t.Fatalf("Got: %s\nExpected: %s\n", value, expected)
	}
}

func TestPing(t *testing.T) {
	shard := NewShard("TestPingData", "TestPingComp", 0, 1)
	check(shard.EnsureRepos(), t)
	s := httptest.NewServer(shard.ShardMux())
	defer s.Close()

	res, err := http.Get(s.URL + "/ping")
	check(err, t)
	checkBody(res, "pong\n", t)
	res.Body.Close()
}
