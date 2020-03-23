package client

import (
	"fmt"
	"strings"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
)

func ExampleAPIClient_CreateRepo() {

	// Create a repo called "test"  and print the list of
	// repositories.
	c, err := client.NewFromAddress("127.0.0.1:30650")
	if err != nil {
		panic(err)
	}

	if _, err := c.PfsAPIClient.CreateRepo(
		c.Ctx(),
		&pfs.CreateRepoRequest{
			Repo:        client.NewRepo("test"),
			Description: "A test repo",
		},
	); err != nil {
		panic(err)
	}

	repos, err := c.ListRepo()
	if err != nil {
		panic(err)
	}
	fmt.Println(repos)

	// Output:
	// // [repo:<name:"test" > created:<seconds:1584994990 nanos:139176186 > description:"A test repo"
}

func ExampleAPIClient_DeleteRepo() {

	// Delete a repository called test.
	c, err := client.NewFromAddress("127.0.0.1:30650")
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

	if _, err := c.DeleteRepo("test"); err != nil {
		panic(err)
	}
}

func ExampleAPIClient_ListRepo() {

	// View the list of existing repositories.
	c, err := client.NewFromAddress("127.0.0.1:30650")
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

	repos, err := c.ListRepo()
	if err != nil {
		panic(err)
	}
	fmt.Println(repos)
}

func ExampleAPIClient_PutFileWriter() {

	// This method enables you to put data into a
	// Pachyderm repo by using an io.Writer API.

	c, err := client.NewFromAddress("127.0.0.1:30650")
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

func ExampleAPIClient_NewPutFileClient() {

	// This method enables you to group multiple put file operations into one
	// request.

	c, err := client.NewFromAddress("127.0.0.1:30650")
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
	pfc, err := c.NewPutFileClient()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := pfc.Close(); err != nil {
			panic(err)
		}
	}()
	if _, err := pfc.PutFile("test", "master", "file", strings.NewReader("foo\n")); err != nil {
		panic(err)
	}
	files, err := c.ListFile("test", "master", "/")
	if err != nil {
		panic(err)
	}
	fmt.Println(files)
}

func ExampleAPIClient_PutFile_string() {

	// This method provides a simple way to to put file into a Pachyderm repo
	// by using io.Reader. In this example, Pachyderm client reads the specified
	// string in the scripts and adds it to a file named "file" in the "test"
	// repository.

	c, err := client.NewFromAddress("127.0.0.1:30650")
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

	if _, err := c.PutFile("test", "master", "file", strings.NewReader("foo\n")); err != nil {
		panic(err)
	}
	files, err := c.ListFile("test", "master", "/")
	if err != nil {
		panic(err)
	}
	fmt.Println(files)

	// Output:
	// [file:<commit:<repo:<name:"test" > id:"3534c68a694747b6a78a6334ceae437e" > path:"/file" > file_type:FILE size_bytes:4 committed:<seconds:1584999407 nanos:25292638 > hash:"\031\375\365{\337\236\265\251`+\372\234\016m\327\35585\370\375C\035\221P\003\352\202tw\007\276f" ]
}

func ExampleAPIClient_PutFile_file() {

	// This method provides a simple way to to put contents of a file into a
	// Pachyderm repo by using io.Reader. In this example the Pachyderm client
	// opens the file named "text.md" and then uses the PutFile method to add it
	// to the repo "test@master".

	c, err := client.NewFromAddress("127.0.0.1:30650")
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

	f, err := os.Open("text.md")
	if err != nil {
		panic(err)
	}
	if _, err := c.PutFile("test", "master", "text", f); err != nil {
		panic(err)
	}
	files, err := c.ListFile("test", "master", "/")
	if err != nil {
		panic(err)
	}
	fmt.Println(files)
}
