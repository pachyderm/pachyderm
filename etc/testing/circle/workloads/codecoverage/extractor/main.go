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

var logger = log.Default()

func main() {

	logger.Println("Running extractor code")

	inputFolder := os.Args[1]
	outputFolder := os.Args[2]

	tgzFiles := []string{}
	txtFiles := []string{}
	logger.Printf("About to walk Input: %s \nFiles: %v \nOutput: %s \n", inputFolder, tgzFiles, outputFolder)
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
			canonPath := strings.Replace(path, sym, "", 1)
			if strings.HasSuffix(d.Name(), ".tgz") { // extractable files integration test coverage
				tgzFiles = append(tgzFiles, canonPath)
			} else if strings.HasSuffix(d.Name(), ".txt") { // .txt covergage profile from normal go test
				txtFiles = append(txtFiles, canonPath)
			} else {
				logger.Printf("Discovered unexpected file extracting coverage. Ignoring %v \n", canonPath)
			}
		}
		return nil
	})
	if err != nil {
		logger.Fatal(err)
	}
	if len(tgzFiles) < 1 {
		logger.Println("No files found to extract, ending early.")
		return
	}
	branch := strings.Split( // DNJ TODO - encapsulate?
		strings.TrimLeft(tgzFiles[0], string(filepath.Separator)),
		string(filepath.Separator),
	)[0]
	outputPath := filepath.Join(outputFolder, branch)
	logger.Printf("Branch: %s \nFiles: %v \nOutput: %s \n", branch, tgzFiles, outputPath)
	err = os.MkdirAll(outputPath, fs.ModePerm)
	if err != nil {
		logger.Fatal(err)
	}
	// extract files
	for _, fileName := range tgzFiles { // DNJ TODO - refactor this for readability probably
		file, err := os.Open(filepath.Join(inputFolder, fileName))
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
			if header.Typeflag == tar.TypeReg && !strings.HasSuffix(header.Name, "error.txt") {
				fmt.Printf("Extracting %v ; %v\n", header.Name, header.Typeflag)
				splitPath := strings.Split(header.Name, string(filepath.Separator)) // DNJ TODO - this is confusing and repeated below
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
	for _, fileName := range txtFiles {
		file, err := os.Open(filepath.Join(inputFolder, fileName))
		logger.Printf("Saving normal coverage file: %v", fileName)
		splitPath := strings.Split(fileName, string(filepath.Separator))
		outputFile := filepath.Join(outputFolder, branch, splitPath[len(splitPath)-1])
		newFile, err := os.OpenFile(outputFile, os.O_CREATE|os.O_RDWR, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		_, err = io.Copy(newFile, file)
		newFile.Close()
		if err != nil {
			log.Fatal(err)
		}
	}
}
