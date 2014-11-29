package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var once sync.Once

func randSeq(n int) string {
	once.Do(func() { rand.Seed(time.Now().UTC().UnixNano()) })
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type randReader struct{}

func (r randReader) Read(p []byte) (n int, err error) {
	once.Do(func() { rand.Seed(time.Now().UTC().UnixNano()) })
	for i := range p {
		p[i] = byte(rand.Int63() & 0xff)
	}
	return len(p), nil
}

var reader randReader

func TrafficHandler(w http.ResponseWriter, r *http.Request) {
	for i := 0; i < 100; i++ {
		http.Post("http://localhost/pfs/"+randSeq(10), "application/text", reader)
	}
}

// TesterMux creates a multiplexer for a Tester
func TesterMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/traffic", TrafficHandler)
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "pong\n") })

	return mux
}

// RunServer runs a master server listening on port 80
func RunServer() {
	http.ListenAndServe(":80", TesterMux())
}

func main() {
	log.SetFlags(log.Lshortfile)
	log.Print("Listening on port 80...")
	RunServer()
}
