// main implements a the user logic run by the "split" loadtest
// (in loadtest/loadtest.go)
package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"path/filepath"
)

var keySz int

func init() {
	flag.IntVar(&keySz, "key-size", -1, "size of each record's key, in bytes")
}

func main() {
	flag.Parse()
	if keySz < 0 {
		log.Fatalf("must set --key-size")
	}

	if err := filepath.Walk("/pfs/input/", func(path string, info os.FileInfo, _ error) error {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		s := bufio.NewScanner(f)
		for s.Scan() {
			outPath := "/pfs/out/" + s.Text()[:keySz] + ".psv"
			outF, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
			if err != nil {
				return err
			}
			defer outF.Close()
			if _, err := outF.Write(s.Bytes()); err != nil {
				return err
			}
			if _, err := outF.Write([]byte{'\n'}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
}
