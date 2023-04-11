package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	inputFolder := os.Args[1]
	outputFolder := os.Args[2]
	inputCommit := os.Getenv("code-coverage-source_COMMIT") // DNJ TODO - parameterize
	formatArgs := []string{"tool", "covdata", "txtfmt", "-i"}

	extractDir, err := os.MkdirTemp(filepath.Join("/tmpfs", inputCommit), "")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(extractDir)
	// check for 'consumed' file intially?
	files := []string{}
	filepath.WalkDir(inputFolder, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	formatArgs = append(formatArgs, strings.Join(files, ","), "-o", outputFolder)
	// DNJ TODO - do we need explicit merge call
	cmd := exec.Command("go", formatArgs...)
	fmt.Println("Running command: ", cmd.String())
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	// check for consumed file right before save
	// put cov file in pfsout - deferred pipeline pushes to codecov
}
