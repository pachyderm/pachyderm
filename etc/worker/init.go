package main

import (
	"io"
	"os"
)

func cp(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	// make the file executable
	if err = out.Chmod(os.ModePerm); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := cp("/app/worker", "/pach-bin/worker"); err != nil {
		panic(err)
	}
}
