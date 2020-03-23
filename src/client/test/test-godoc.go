package main

import (

  "fmt"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
)

func main() {

	c, err := client.NewFromAddress("192.168.64.14:30650")
	if err != nil {
         panic(err)
  }

  if _, err := c.PfsAPIClient.CreateRepo(
         c.Ctx(),
         &pfs.CreateRepoRequest{
             Repo:        client.NewRepo("test"),
             Description: "A test repo",
             Update:      true,
         },
  ); err != nil {
      panic(err)
  }
  w, err := c.PutFileWriter("test", "master", "file")
  if err != nil {
      panic(err)
  }
	// The defer closes the writing request.
  defer func() {
      if err := w.Close(); err != nil {
          panic(err)
      }
  }()
  if _, err := w.Write([]byte("foo\n")); err != nil {
      panic(err)
  }

	files, err := c.ListFile("test", "master", "/")
	     if err != nil {
	         panic(err)
	     }
	     fmt.Println(files)
}
