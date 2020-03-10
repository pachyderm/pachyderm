package main

import (
	"archive/tar"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	// test is just to rapidly open and close the pipe to make sure the spout handles this situation well
	for x := 0; x < 1000; x++ {
		if err := func() error {

			// Open the /pfs/out pipe with write only permissons (the pachyderm spout will be reading at the other end of this)
			// Note: it won't work if you try to open this with read, or read/write permissions
			out, err := os.OpenFile("/pfs/out", os.O_WRONLY, 0644)
			if err != nil {
				panic(err)
			}
			defer out.Close()

			tw := tar.NewWriter(out)
			defer tw.Close()

			name := fmt.Sprintf("test%v", x)
			// write the header
			for err = tw.WriteHeader(&tar.Header{
				Name: name,
				Mode: 0600,
				Size: int64(len(name)),
			}); err != nil; {
				if !strings.Contains(err.Error(), "broken pipe") {
					return err
				}
				// if there's a broken pipe, just give it some time to get ready for the next message
				time.Sleep(5 * time.Millisecond)
			}
			// and the message
			for _, err = tw.Write([]byte(name)); err != nil; {
				if !strings.Contains(err.Error(), "broken pipe") {
					return err
				}
				// if there's a broken pipe, just give it some time to get ready for the next message
				time.Sleep(5 * time.Millisecond)
			}
			return nil
		}(); err != nil {
			panic(err)
		}
	}
}
