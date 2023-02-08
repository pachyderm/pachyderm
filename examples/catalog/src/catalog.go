package main

import (
	"fmt"

	"github.com/blevesearch/bleve/v2"
)

func main() {
	// open a new index
	mapping := bleve.NewIndexMapping()
	index, err := bleve.New("example.bleve", mapping)
	if err != nil {
		fmt.Errorf("bleve.New: %s\n", err)
	}

	// index some data
	err = index.Index(identifier, "foo")
	if err != nil {
		fmt.Errorf("index.Index: %s\n", err)
	}

	// search for some text
	query := bleve.NewMatchQuery("foo")
	search := bleve.NewSearchRequest(query)
	searchResults, err := index.Search(search)
}
