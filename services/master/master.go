package main

import (
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"github.com/pachyderm-io/pfs/lib/btrfs"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

//TODO commits should be content based

// headPath is the location of a symlink to the latest commit.
var headPath string = ".commits/HEAD"

// IncrCommit generates the next commit path.
func IncrCommit(commit string) (string, error) {
	split := strings.Split(commit, "/")
	index, err := strconv.Atoi(split[len(split)-1])
	if err != nil {
		return "", err
	}
	index++
	split[len(split)-1] = strconv.Itoa(index)
	return strings.Join(split, "/"), nil
}

// DecrCommit generates the previous commit path.
func DecrCommit(commit string) (string, error) {
	split := strings.Split(commit, "/")
	index, err := strconv.Atoi(split[len(split)-1])
	if err != nil {
		return "", err
	}
	index--
	split[len(split)-1] = strconv.Itoa(index)
	return strings.Join(split, "/"), nil
}

// PfsHandler is the core route for modifying the contents of the fileystem.
// Changes are not replicated until a call to CommitHandler.
func PfsHandler(w http.ResponseWriter, r *http.Request, fs *btrfs.FS) {
	if r.Method == "GET" {
		params := r.URL.Query()
		if params.Get("commit") == "" {
			http.StripPrefix("/pfs/", http.FileServer(http.Dir(fs.FilePath("")))).ServeHTTP(w, r)
		} else {
			servePath := fs.FilePath(path.Join(".commits", params.Get("commit")))
			log.Print(servePath)
			http.StripPrefix("/pfs/", http.FileServer(http.Dir(servePath))).ServeHTTP(w, r)
		}
	} else if r.Method == "POST" {
		url := strings.Split(r.URL.Path, "/")
		filename := strings.Join(url[2:], "/")
		size, err := fs.CreateFile(filename, r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		fmt.Fprintf(w, "Added %s, size: %d.\n", filename, size)
	} else if r.Method == "PUT" {
		url := strings.Split(r.URL.Path, "/")
		filename := strings.Join(url[2:], "/")
		size, err := fs.WriteFile(filename, r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		fmt.Fprintf(w, "Wrote %s, size: %d.\n", filename, size)
	} else if r.Method == "DELETE" {
		url := strings.Split(r.URL.Path, "/")
		filename := strings.Join(url[2:], "/")
		err := fs.Remove(filename)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		fmt.Fprintf(w, "Deleted %s.\n", filename)
	}
}

// CommitHandler creates a snapshot of outstanding changes and pushes it to
// replicas.
func CommitHandler(w http.ResponseWriter, r *http.Request, fs *btrfs.FS) {
	client := etcd.NewClient([]string{"http://172.17.42.1:4001", "http://10.1.42.1:4001"})
	log.Printf("Getting replica for %s.", os.Args[1])
	shard_prefix := path.Join("/pfs", "replica", os.Args[1])
	_replica, err := client.Get(shard_prefix, false, false)
	replica := _replica.Node.Value
	log.Print(replica)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
		return
	}

	exists, err := fs.FileExists(".commits")
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
		return
	}
	if exists {
		log.Print("exists")
		last_commit, err := fs.Readlink(headPath)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		new_commit, err := IncrCommit(last_commit)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		err = fs.Snapshot(".", new_commit, true)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		err = fs.Remove(headPath)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		err = fs.Symlink(new_commit, headPath)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		err = btrfs.Sync()
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		fmt.Fprintf(w, "Created commit: %s.\n", new_commit)

		err = fs.Send(last_commit, new_commit,
			func(data io.ReadCloser) error {
				_, err = http.Post("http://"+replica+"/"+"recv",
					"text/plain", data)
				return err
			})
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		fmt.Fprintf(w, "Sent commit to: %s.\n", replica)
	} else {
		log.Print("First commit.")
		first_commit := path.Join(".commits", "0")
		err = fs.MkdirAll(".commits")
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		err = fs.Snapshot(".", first_commit, true)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		err = fs.Symlink(first_commit, headPath)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		err = btrfs.Sync()
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		fmt.Fprintf(w, "Created commit: %s.\n", first_commit)
		err = fs.SendBase(first_commit,
			func(data io.ReadCloser) error {
				_, err = http.Post("http://"+replica+"/"+"recv",
					"text/plain", data)
				return err
			})
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		fmt.Fprintf(w, "Sent commit to: %s.\n", replica)
	}
}

//func BranchHandler(w http.ResponseWriter, r *http.Request, fs *brtfs.FS) {
//}

//BrowseHandler exposes existing snapshots.
func BrowseHandler(w http.ResponseWriter, r *http.Request, fs *btrfs.FS) {
	exists, err := fs.FileExists(".commits")
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
		return
	}
	if exists {
		commits, err := fs.ReadDir(".commits")
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}

		for _, c := range commits {
			fmt.Fprintf(w, "<html>")
			fmt.Fprintf(w, "<pre>")
			fmt.Fprintf(w, "<a href=\"/pfs/.commits/%s\">%s</a> - %s <a href=\"/del/%s\">Delete</a>\n", c.Name(), c.Name(), c.ModTime().Format("Jan 2, 2006 at 3:04pm (PST)"), c.Name())
			fmt.Fprintf(w, "</pre>")
			fmt.Fprintf(w, "</html>")
		}
	} else {
		fmt.Fprint(w, "Nothing here :(.")
	}
}

// MasterMux creates a multiplexer for a Master writing to the passed in FS.
func MasterMux(fs *btrfs.FS) *http.ServeMux {
	mux := http.NewServeMux()

	commitHandler := func(w http.ResponseWriter, r *http.Request) {
		CommitHandler(w, r, fs)
	}

	pfsHandler := func(w http.ResponseWriter, r *http.Request) {
		PfsHandler(w, r, fs)
	}

	browseHandler := func(w http.ResponseWriter, r *http.Request) {
		BrowseHandler(w, r, fs)
	}

	mux.HandleFunc("/commit", commitHandler)
	mux.HandleFunc("/pfs/", pfsHandler)
	mux.HandleFunc("/browse", browseHandler)
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "pong\n") })

	return mux
}

// RunServer runs a master server listening on port 80
func RunServer(fs *btrfs.FS) {
	http.ListenAndServe(":80", MasterMux(fs))
}

func main() {
	log.SetFlags(log.Lshortfile)
	fs := btrfs.NewFS("master-" + os.Args[1])
	fs.EnsureNamespace()
	log.Print("Listening on port 80...")
	RunServer(fs)
}
