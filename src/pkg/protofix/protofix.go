package main

import(
	"os"
	"fmt"

	"github.com/pachyderm/pachyderm/src/pkg/protofix/protofix"
)

func usage() {
	fmt.Println("usage: protofix <rootdirectory>")
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Missing root directory")
		usage()
		os.Exit(1)
	}

	protofix.FixAllPBGOFilesInDirectory(os.Args[1])
}
