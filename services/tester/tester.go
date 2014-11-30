package main

import (
	"fmt"
	"io"
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

type ConstReader struct{}

func (r ConstReader) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = 'a'
	}
	return len(p), nil
}

var reader ConstReader

func timeParam(r *http.Request) string {
	if p := r.URL.Query().Get("time"); p != "" {
		return p
	}
	return "30"
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	workers := 3
	var wg sync.WaitGroup
	wg.Add(workers)
	startTime := time.Now()
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for time.Since(startTime) < (10 * time.Second) {
				url := "http://172.17.42.1/pfs/" + randSeq(10)
				_, err := http.Post(url, "application/text", io.LimitReader(reader, 1<<10))
				if err != nil {
					//TODO do something here?
					log.Print(err)
					return
				}
			}
		}()
	}
	wg.Wait()
	fmt.Fprint(w, "Sent files.\n")
}

// TesterMux creates a multiplexer for a Tester
func TesterMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/profile", ProfileHandler)
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
