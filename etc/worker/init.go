//nolint:wrapcheck
package main

import (
	"fmt"
	"io"
	"os"
)

func cp(src, dst string) error {
	// init containers can be restarted spuriously, so make sure we actually need to copy
	info, err := os.Stat(dst)
	if err == nil {
		// if the file already exists and is executable, assume it's correct
		if info.Mode() == os.ModePerm {
			return nil
		}
	} else if !os.IsNotExist(err) {
		return err
	}

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
	for _, bin := range []string{"worker", "pachctl", "pachtf"} {
		src, dst := fmt.Sprintf("/app/%s", bin), fmt.Sprintf("/pach-bin/%s", bin)
		if err := cp(src, dst); err != nil {
			panic(err)
		}
	}
}
