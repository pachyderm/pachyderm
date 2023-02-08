package main

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/blevesearch/bleve/v2"
)

func main() {
	// open a new index
	mapping := bleve.NewIndexMapping()
	index, err := bleve.New("/pfs/out/index", mapping)

	// index some data
	if err = filepath.Walk("/pfs/books", func(path string, info fs.FileInfo, err error) error {
		if err := index.Index(identifier); err != nil {
			fmt.Errorf("Walk: %v\n", err)
		}
	}); err != nil {
		fmt.Errorf("Walk: %v\n", err)
	}

}
