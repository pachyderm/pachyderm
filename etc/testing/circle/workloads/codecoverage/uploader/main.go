package main

import (
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	logger := log.Default()
	logger.Println("Running uploader code")
	inputFolder := os.Args[1]
	// outputFolder := os.Args[2]
	codecovToken := os.Getenv("CODECOV_TOKEN")
	logger.Printf("token: %s", codecovToken) // DNJ TODO - delete
	cmd := exec.Command("ls", "/pfs")
	if out, err := cmd.CombinedOutput(); err != nil {
		logger.Printf("Command failed: %v \n", string(out))
		log.Fatal(err)
	} else {
		logger.Printf("Command succeeded: %v \n", string(out))
	}
	sym, err := filepath.EvalSymlinks(inputFolder)
	if err != nil {
		logger.Fatal(err)
	}
	files := []string{}
	filepath.WalkDir(sym, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() { // DNJ TODO - check for non-code coverage files
			files = append(files, strings.Replace(path, sym, inputFolder, 1))
		}
		return nil
	})
	if len(files) > 1 {
		logger.Fatalf("Only one file expected but found %v \n", files)
	}
	logger.Printf("file list: %v \n", files)
	// check for consumed file right before save
	// put cov file in codecov - deferred pipeline pushes to codecov
}
