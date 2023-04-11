package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	logger := log.Default()
	logger.Println("Running extractor code")

	inputFolder := os.Args[1]
	outputFolder := os.Args[2]
	// inputCommit := os.Getenv("code-coverage-source_COMMIT") // DNJ TODO - parameterize

	// check for 'consumed' file intially?
	files := []string{}
	logger.Printf("About to walk Input: %s \nFiles: %v \nOutput: %s \n", inputFolder, files, outputFolder)
	out, _ := exec.Command("ls", "..").CombinedOutput()
	logger.Printf("Home Dir: %s \n", string(out))
	out, _ = exec.Command("ls", filepath.Join(inputFolder)).CombinedOutput()
	logger.Printf("Input folder ls: %s \n", string(out))
	sym, err := filepath.EvalSymlinks(inputFolder)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Printf("Symlink ls: %s \n", sym)
	err = filepath.WalkDir(sym, func(path string, d fs.DirEntry, err error) error {
		logger.Printf("Walking Dir: %s + %s is dir? %v \n", path, d.Name(), d.IsDir())
		if !d.IsDir() {
			files = append(files, strings.Replace(path, sym, "", 1))
		}
		return nil
	})
	if err != nil {
		logger.Fatal(err)
	}
	if len(files) < 1 {
		logger.Println("No files found to extract, ending early.")
		return
	}
	branch := strings.Split(
		strings.TrimLeft(files[0], string(filepath.Separator)),
		string(filepath.Separator),
	)[0]
	outputPath := filepath.Join(outputFolder, branch)
	logger.Printf("Branch: %s \nFiles: %v \nOutput: %s \n", branch, files, outputPath)
	err = os.MkdirAll(outputPath, fs.ModePerm)
	if err != nil {
		logger.Fatal(err)
	}
	for _, f := range files {
		file, err := os.Open(filepath.Join(inputFolder, f))
		gzr, err := gzip.NewReader(file)
		if err != nil {
			log.Fatal(err)
		}
		defer gzr.Close()

		tarReader := tar.NewReader(gzr)
		for {
			header, err := tarReader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			if header.Typeflag == tar.TypeReg {
				fmt.Printf("Extracting %v ; %v\n", header.Name, header.Typeflag)
				splitPath := strings.Split(header.Name, string(filepath.Separator))
				outputFile := filepath.Join(outputFolder, branch, splitPath[len(splitPath)-1])
				newFile, err := os.OpenFile(outputFile, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
				if err != nil {
					log.Fatal(err)
				}
				_, err = io.Copy(newFile, tarReader)
				newFile.Close()
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
	// check for consumed file right before save
	// put cov file in pfsout - deferred pipeline pushes to codecov
}
