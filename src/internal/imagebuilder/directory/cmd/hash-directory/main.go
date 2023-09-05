package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pachyderm/pachyderm/v2/src/internal/imagebuilder/directory"
)

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	out, err := directory.Hash(dir)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%x\n", out)
}
