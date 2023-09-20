//go:build k8s

package client

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

func ExampleAPIClient_CreateRepo() {
	// Create a repo called "test" and print the list of
	// repositories.
	c, err := NewFromURI(context.Background(), "127.0.0.1:30650")
	if err != nil {
		panic(err)
	}

	if _, err := c.PfsAPIClient.CreateRepo(
		c.Ctx(),
		&pfs.CreateRepoRequest{
			Repo:        NewRepo(pfs.DefaultProjectName, "test"),
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
	// Delete a repository called "test".
	c, err := NewFromURI(context.Background(), "127.0.0.1:30650")
	if err != nil {
		panic(err)
	}

	if _, err := c.PfsAPIClient.CreateRepo(
		c.Ctx(),
		&pfs.CreateRepoRequest{
			Repo:        NewRepo(pfs.DefaultProjectName, "test"),
			Description: "A test repo",
			Update:      true,
		},
	); err != nil {
		panic(err)
	}

	if err := c.DeleteRepo("default", "test", false); err != nil {
		panic(err)
	}
}

func ExampleAPIClient_ListRepo() {
	// View the list of existing repositories.
	c, err := NewFromURI(context.Background(), "127.0.0.1:30650")
	if err != nil {
		panic(err)
	}

	if _, err := c.PfsAPIClient.CreateRepo(
		c.Ctx(),
		&pfs.CreateRepoRequest{
			Repo:        NewRepo(pfs.DefaultProjectName, "test"),
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

func ExampleAPIClient_NewModifyFileClient() {
	// This method enables you to group multiple "put file" operations into one
	// request.

	c, err := NewFromURI(context.Background(), "127.0.0.1:30650")
	if err != nil {
		panic(err)
	}

	testRepo := NewRepo(pfs.DefaultProjectName, "test")
	testCommit := testRepo.NewCommit("master", "")

	if _, err := c.PfsAPIClient.CreateRepo(
		c.Ctx(),
		&pfs.CreateRepoRequest{
			Repo:        testRepo,
			Description: "A test repo",
			Update:      true,
		},
	); err != nil {
		panic(err)
	}
	mfc, err := c.NewModifyFileClient(testCommit)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := mfc.Close(); err != nil {
			panic(err)
		}
	}()
	if err := mfc.PutFile("file", strings.NewReader("foo\n")); err != nil {
		panic(err)
	}
	files, err := c.ListFileAll(testCommit, "/")
	if err != nil {
		panic(err)
	}
	fmt.Println(files)
}

func ExampleAPIClient_PutFile_string() {
	// This method provides a simple way to to put file into a Pachyderm repo
	// by using "io.Reader". In this example, Pachyderm client reads the specified
	// string in the scripts and adds it to a file named "file" in the "test"
	// repository.

	c, err := NewFromURI(context.Background(), "127.0.0.1:30650")
	if err != nil {
		panic(err)
	}

	testRepo := NewRepo(pfs.DefaultProjectName, "test")
	testCommit := testRepo.NewCommit("master", "")

	if _, err := c.PfsAPIClient.CreateRepo(
		c.Ctx(),
		&pfs.CreateRepoRequest{
			Repo:        testRepo,
			Description: "A test repo",
			Update:      true,
		},
	); err != nil {
		panic(err)
	}

	if err := c.PutFile(testCommit, "file", strings.NewReader("foo\n")); err != nil {
		panic(err)
	}
	files, err := c.ListFileAll(testCommit, "/")
	if err != nil {
		panic(err)
	}
	fmt.Println(files)

	// Output:
	// [file:<commit:<repo:<name:"test" > id:"3534c68a694747b6a78a6334ceae437e" > path:"/file" > file_type:FILE size_bytes:4 committed:<seconds:1584999407 nanos:25292638 > hash:"\031\375\365{\337\236\265\251`+\372\234\016m\327\35585\370\375C\035\221P\003\352\202tw\007\276f" ]
}

func ExampleAPIClient_PutFile_file() {
	// This method provides a simple way to to put contents of a file into a
	// Pachyderm repo by using "io.Reader". In this example the Pachyderm client
	// opens the file named "text.md" and then uses the PutFile method to add it
	// to the repo "test@master".

	c, err := NewFromURI(context.Background(), "127.0.0.1:30650")
	if err != nil {
		panic(err)
	}

	testRepo := NewRepo(pfs.DefaultProjectName, "test")
	testCommit := testRepo.NewCommit("master", "")

	if _, err := c.PfsAPIClient.CreateRepo(
		c.Ctx(),
		&pfs.CreateRepoRequest{
			Repo:        testRepo,
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
	if err := c.PutFile(testCommit, "text", f); err != nil {
		panic(err)
	}
	files, err := c.ListFileAll(testCommit, "/")
	if err != nil {
		panic(err)
	}
	fmt.Println(files)
}

func ExampleAPIClient_CreateBranch() {
	// When you create a branch, you need to specify four parameters:
	// a repository name, a name of the new branch, a commit ID or a
	// branch that will be used as head, and provenance. Provenance
	// is an optional paramater and get be set to "nil" for nothing,
	// to a commit ID or to a branch name.
	c, err := NewFromURI(context.Background(), "127.0.0.1:30650")
	if err != nil {
		panic(err)
	}

	if _, err := c.PfsAPIClient.CreateRepo(
		c.Ctx(),
		&pfs.CreateRepoRequest{
			Repo:        NewRepo(pfs.DefaultProjectName, "test"),
			Description: "A test repo",
			Update:      true,
		},
	); err != nil {
		panic(err)
	}

	if err := c.CreateBranch("default", "test", "newbranch", "master", "", nil); err != nil {
		panic(err)
	}

	bis, err := c.ListBranch("default", "test")
	if err != nil {
		panic(err)
	}
	fmt.Print(bis)

	// Output:
	// [branch:<repo:<name:"test" > name:"newbranch" > head:<repo:<name:"test" > id:"1eb10a08e29f4abcaaa647dbaa113514" > name:"newbranch" ]
}

func ExampleAPIClient_ListCommit() {
	// This method enables you to list commits in a repository.
	// To list all commits, pass only the name of the repository. If you want
	// to list specific commits, you can specify "to" and "from" parameters.
	// These both parameters accept a commit ID.
	// In the example below, the "to" parameter is left blank, and the "from"
	// parameter is set to "0", which means all commits.

	c, err := NewFromURI(context.Background(), "127.0.0.1:30650")
	if err != nil {
		panic(err)
	}

	testRepo := NewRepo(pfs.DefaultProjectName, "test")
	testCommit := testRepo.NewCommit("master", "")

	if _, err := c.PfsAPIClient.CreateRepo(
		c.Ctx(),
		&pfs.CreateRepoRequest{
			Repo:        testRepo,
			Description: "A test repo",
			Update:      true,
		},
	); err != nil {
		panic(err)
	}

	if err := c.PutFile(testCommit, "file", strings.NewReader("foo\n")); err != nil {
		panic(err)
	}
	if err := c.PutFile(testCommit, "file", strings.NewReader("bar\n")); err != nil {
		panic(err)
	}
	if err := c.PutFile(testCommit, "file", strings.NewReader("buzz\n")); err != nil {
		panic(err)
	}

	cis, err := c.ListCommit(testRepo, testCommit, nil, 0)
	if err != nil {
		panic(err)
	}
	for _, ci := range cis {
		fmt.Println(ci)
	}

	// Output:
	// commit:<repo:<name:"test" > id:"1eb10a08e29f4abcaaa647dbaa113514" > branch:<repo:<name:"test" > name:"master" > origin:<> parent_commit:<repo:<name:"test" > id:"f222d4d469454541b373d6c0f66f3f4a" > child_commits:<repo:<name:"test" > id:"1a9f7f3607c54edcaa3f716605bfd2e2" > started:<seconds:1585775813 nanos:529811155 > finished:<seconds:1585775813 nanos:533868208 > size_bytes:12 tree:<hash:"135347e1392ffa244b030d855f3335562970cbbc73125f4bf01ca82e8a9a6eda2a1591573eda30c86c2644c67d989506d2dfc14219865e75f3b383f32dcb550d" >
	// commit:<repo:<name:"test" > id:"f222d4d469454541b373d6c0f66f3f4a" > branch:<repo:<name:"test" > name:"master" > origin:<> parent_commit:<repo:<name:"test" > id:"3e6bdffda01b47ecbbdf4833343317ad" > child_commits:<repo:<name:"test" > id:"1eb10a08e29f4abcaaa647dbaa113514" > started:<seconds:1585774635 nanos:376380297 > finished:<seconds:1585774635 nanos:382910322 > size_bytes:8 tree:<hash:"2e65e9751f240226046aafad86aa07526fea6c7343c93bc3bd859c1812ebcad86af5be74d9a5a06e4978791852864a07cc3c820802b7eb304ff7792e6e23458e" >
	// commit:<repo:<name:"test" > id:"3e6bdffda01b47ecbbdf4833343317ad" > branch:<repo:<name:"test" > name:"master" > origin:<> child_commits:<repo:<name:"test" > id:"f222d4d469454541b373d6c0f66f3f4a" > started:<seconds:1585772570 nanos:475140438 > finished:<seconds:1585772570 nanos:481163804 > size_bytes:4 tree:<hash:"9b7798e84aa2885ac7fd1b86fadf556e3330a36802979a4349e44743b26f5e63584506299ecae5bcd5e72c4f42e236618ef8600f6b8da5fd17d8f4ba8d03d0d1" >
}

func ExampleAPIClient_CreateBranch_fromcommit() {
	// This example demonstrates how you can create a branch from a specific
	// commit. In this example, we create a branch from a commit by referring
	// to it with the commit index ("cis[1].Commit.ID"). The latest commit
	// always has the index of "0" and the earliest commit will have a
	// corresponding index depending on the number commits in the repo.
	// In the example below, we will create a branch from the commit with
	// index [1] that adds "file2" in the repository "test" and will create
	// a branch "new-branch" in which we will add "file4". "newbranch" will
	// have "file1", "file2", and "file4", but will not have "file3".
	c, err := NewFromURI(context.Background(), "127.0.0.1:30650")
	if err != nil {
		panic(err)
	}

	testRepo := NewRepo(pfs.DefaultProjectName, "test")
	testCommit := testRepo.NewCommit("master", "")

	if _, err := c.PfsAPIClient.CreateRepo(
		c.Ctx(),
		&pfs.CreateRepoRequest{
			Repo:        testRepo,
			Description: "A test repo",
			Update:      true,
		},
	); err != nil {
		panic(err)
	}

	if err := c.PutFile(testCommit, "file1", strings.NewReader("foo\n")); err != nil {
		panic(err)
	}
	if err := c.PutFile(testCommit, "file2", strings.NewReader("bar\n")); err != nil {
		panic(err)
	}
	if err := c.PutFile(testCommit, "file3", strings.NewReader("buzz\n")); err != nil {
		panic(err)
	}

	cis, err := c.ListCommit(testRepo, testCommit, nil, 0)
	if err != nil {
		panic(err)
	}

	if err := c.CreateBranch("default", "test", "new-branch", "", cis[1].Commit.Id, nil); err != nil {
		panic(err)
	}

	newCommit := testRepo.NewCommit("new-branch", "")
	if err := c.PutFile(newCommit, "file4", strings.NewReader("fizz\n")); err != nil {
		panic(err)
	}
	files, err := c.ListFileAll(newCommit, "/")
	if err != nil {
		panic(err)
	}
	fmt.Println(files)

	// Output:
	// [file:<commit:<repo:<name:"test" > id:"c9f65a41818947b29d99c8a3140a3aa7" > path:"/file1" > file_type:FILE size_bytes:4 committed:<seconds:1585780965 nanos:118756372 > hash:"\031\375\365{\337\236\265\251`+\372\234\016m\327\35585\370\375C\035\221P\003\352\202tw\007\276f"  file:<commit:<repo:<name:"test" > id:"c9f65a41818947b29d99c8a3140a3aa7" > path:"/file2" > file_type:FILE size_bytes:4 committed:<seconds:1585780965 nanos:118756372 > hash:"\252t\367r\331\3442S\306\n\013f#\1770\350\3758\345\275\030\374\241\351Z\221\002\273\347sS@"  file:<commit:<repo:<name:"test" > id:"c9f65a41818947b29d99c8a3140a3aa7" > path:"/file4" > file_type:FILE size_bytes:5 committed:<seconds:1585780965 nanos:118756372 > hash:"\002hQ!\276\242\2710\362Y4\215SY\275v\037)T\205m\177JK\340\322\177\247\"\"A\367" ]
}

func ExampleAPIClient_ListCommitF() {
	// ListCommitF streams information about each commit one at a time
	// instead of returning all commits at once. Most of Pachyderm's
	// "List" methods have a similar function.

	c, err := NewFromURI(context.Background(), "127.0.0.1:30650")
	if err != nil {
		panic(err)
	}

	testRepo := NewRepo(pfs.DefaultProjectName, "test")
	testCommit := testRepo.NewCommit("master", "")

	if _, err := c.PfsAPIClient.CreateRepo(
		c.Ctx(),
		&pfs.CreateRepoRequest{
			Repo:        testRepo,
			Description: "A test repo",
			Update:      true,
		},
	); err != nil {
		panic(err)
	}

	if err := c.PutFile(testCommit, "file1", strings.NewReader("foo\n")); err != nil {
		panic(err)
	}
	if err := c.PutFile(testCommit, "file2", strings.NewReader("bar\n")); err != nil {
		panic(err)
	}
	if err := c.PutFile(testCommit, "file3", strings.NewReader("buzz\n")); err != nil {
		panic(err)
	}

	var nCommits int
	if err := c.ListCommitF(testRepo, testCommit, nil, 0, false, func(ci *pfs.CommitInfo) error {
		fmt.Println(ci)
		return nil
	}); err != nil {
		panic(err)
	}
	fmt.Println(nCommits)

	// Output:
	// commit:<repo:<name:"test" > id:"3ccb4b6876c64b7eb4d9ad4c493b9358" > branch:<repo:<name:"test" > name:"master" > origin:<> parent_commit:<repo:<name:"test" > id:"18308cb493e047f9b0719b1ccd1c03d3" > started:<seconds:1585782258 nanos:38033497 > finished:<seconds:1585782258 nanos:41459369 > size_bytes:13 tree:<hash:"f429d8646b4e920267758c31966489f64820be928339303ae5c944432e364fcff0e4e2c008758a1a0693931cb526eb233ce51824ae20d2f9ff9f2fb5ff95f38b" >
	// commit:<repo:<name:"test" > id:"18308cb493e047f9b0719b1ccd1c03d3" > branch:<repo:<name:"test" > name:"master" > origin:<> parent_commit:<repo:<name:"test" > id:"de3f5657da294c7a899273693353eb4b" > child_commits:<repo:<name:"test" > id:"3ccb4b6876c64b7eb4d9ad4c493b9358" > started:<seconds:1585782258 nanos:30833622 > finished:<seconds:1585782258 nanos:34399188 > size_bytes:8 tree:<hash:"e12d984752f65223ca2374f11ac93f18641411c2909fab56070140c80c1e7547c7dfdea855b0423a411f894320d6d8b1cde126a5b00c3ccce6d1a8d02bbac4ab" >
	// commit:<repo:<name:"test" > id:"de3f5657da294c7a899273693353eb4b" > branch:<repo:<name:"test" > name:"master" > origin:<> child_commits:<repo:<name:"test" > id:"18308cb493e047f9b0719b1ccd1c03d3" > started:<seconds:1585782258 nanos:22342806 > finished:<seconds:1585782258 nanos:27145336 > size_bytes:4 tree:<hash:"85d69e603965f7a15867e24a150783d44b778818582aec8311dc90cd6094be61ba5ef7e3a98a950e10317f784de031941de48a4be524af0fa191558da41cffb5" >
	// 0

}

func ExampleAPIClient_CreatePipeline() {
	// This example shows how to create a simple pipeline
	// called "test-pipeline" that uses a basic image and runs a
	// bash command that copies the "/pfs/test" directory
	// to "/pfs/out" directory. Because no image is specified,
	// Pachyderm will use the basic image. The input is defined as
	// the repo "test" with the "/*" glob pattern.
	c, err := NewFromURI(context.Background(), "192.168.64.2:30650")
	if err != nil {
		panic(err)
	}

	testRepo := NewRepo(pfs.DefaultProjectName, "test")
	testCommit := testRepo.NewCommit("master", "")

	if _, err := c.PfsAPIClient.CreateRepo(
		c.Ctx(),
		&pfs.CreateRepoRequest{
			Repo:        testRepo,
			Description: "A test repo",
			Update:      true,
		},
	); err != nil {
		panic(err)
	}

	if err := c.PutFile(testCommit, "file1", strings.NewReader("foo\n")); err != nil {
		panic(err)
	}

	if err := c.CreatePipeline(
		"default",
		"test-pipeline",
		"",
		[]string{"bash"},
		[]string{
			"cp /pfs/test/* /pfs/out/",
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		NewPFSInput(pfs.DefaultProjectName, "test", "/*"),
		"",
		false,
	); err != nil {
		panic(err)
	}

	pipelines, err := c.ListPipeline(true)
	if err != nil {
		panic(err)
	}
	fmt.Println(pipelines)

	// Output:
	// [pipeline:<name:"test-pipeline" > version:1 transform:<image:"ubuntu:20.04" cmd:"bash" stdin:"cp /pfs/test/* /pfs/out/" > parallelism_spec:<constant:1 > created_at:<seconds:1585783817 nanos:990814317 > output_branch:"master" input:<pfs:<name:"test" repo:"test" branch:"master" glob:"/*" > > cache_size:"64M" salt:"95e86369074f472c87d77f1116656813" max_queue_size:1 spec_commit:<repo:<name:"__spec__" > id:"7ddade34799b450db6b10231bbb287b7" > datum_tries:3 ]
}
