package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	t "github.com/n0madic/twitter-scraper"
)

func main() {
	queries, err := ioutil.ReadDir("/pfs/queries/")
	if err != nil {
		panic(err)
	}
	for _, q := range queries {
		f, err := os.Open(filepath.Join("/pfs/queries", q.Name()))
		if err != nil {
			panic(err)
		}
		defer f.Close()
		out, err := os.Create(filepath.Join("/pfs/out/", q.Name()))
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := out.Close(); err != nil {
				panic(err)
			}
		}()
		s := bufio.NewScanner(f)
		for s.Scan() {
			for tweet := range t.GetTweets(context.Background(),
				strings.TrimSpace(s.Text()), 999999999) {
				if tweet.Error != nil {
					panic(tweet.Error)
				}
				fmt.Fprintf(out, "<|startoftext|> %s <|endoftext|>\n", tweet.Text)
			}
		}
	}
}
