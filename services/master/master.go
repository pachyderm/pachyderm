package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
    "fmt"
    "github.com/coreos/go-etcd/etcd"
    "github.com/jdoliner/btrfs"
    "path"
    "strconv"
)

func headPath() string { return ".commits/HEAD" }

func IncrCommit(snap string) (string, error) {
	split := strings.Split(snap, "/")
	index, err := strconv.Atoi(split[len(split) - 1])
	if err != nil { return "", err }
	index++
    split[len(split) - 1] = strconv.Itoa(index)
	return strings.Join(split, "/"), nil
}

//  http://host/add/file
func AddHandler(w http.ResponseWriter, r *http.Request, fs *btrfs.FS) {
    url := strings.Split(r.URL.String(), "/")
    filename := strings.Join(url[2:], "/")
    file, err := fs.Create(filename)
    if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
    defer file.Close()
    size, err := io.Copy(file, r.Body)
    if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
    fmt.Fprintf(w, "Added %s, size: %d.\n", filename, size)
}

// http://host/commit
func CommitHandler(w http.ResponseWriter, r *http.Request, fs *btrfs.FS) {
    client := etcd.NewClient([]string{"http://172.17.42.1:4001"})
	log.Printf("Getting slaves for %s.", os.Args[3])
	shard_prefix := path.Join("/pfsd", os.Args[3])
    slaves, err := client.Get(shard_prefix, false, false)
	log.Printf("Got slaves.")
	log.Print(slaves)
    if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }

    exists, err := fs.FileExists(".commits")
    if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
    if exists {
		log.Print("exists")
        last_commit, err := fs.Readlink(headPath())
        if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
        new_commit, err := IncrCommit(last_commit)
        if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
        err = fs.Snapshot(".", new_commit, true)
        if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
        err = fs.Remove(headPath())
        if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
        err = fs.Symlink(new_commit, headPath())
        if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
        err = btrfs.Sync()
        if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
        fmt.Fprintf(w, "Created commit: %s.\n", new_commit)
        for _, s := range slaves.Node.Nodes {
			log.Print("Key: ", s.Key)
			log.Print(path.Join(shard_prefix, "master"))
			if s.Key != path.Join(shard_prefix, "master") {
				log.Print(s.Value)
				log.Print("Sending update to:" + s.Value)
				err = fs.Send(last_commit, new_commit,
						func (data io.ReadCloser) error {
							_, err = http.Post("http://" + s.Value + "/" + "recv",
									"text/plain", data)
							return err
						})
				if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
				fmt.Fprintf(w, "Sent commit to: %s.\n", s.Value)
			}
        }
    } else {
		log.Print("First commit.")
        first_commit := path.Join(".commits", "0")
        err = fs.MkdirAll(".commits", 0777)
        if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
        err = fs.Snapshot("." , first_commit, true)
        if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
        err = fs.Symlink(first_commit, headPath())
        if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
        err = btrfs.Sync()
        if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
        fmt.Fprintf(w, "Created commit: %s.\n", first_commit)
        for _, s := range slaves.Node.Nodes {
			log.Print("Key: ", s.Key)
			log.Print(path.Join(shard_prefix, "master"))
			if s.Key != path.Join(shard_prefix, "master") {
				log.Print(s.Value)
				if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
				log.Print("Sending base to: " + s.Value)
				err = fs.SendBase(first_commit,
						func (data io.ReadCloser) error {
							_, err = http.Post("http://" + s.Value + "/" + "recvbase",
									"text/plain", data)
							return err
						})
				if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
				fmt.Fprintf(w, "Sent commit to: %s.\n", s.Value)
			}
        }
    }
}

// http://host/recvbase

func RecvBaseHandler(w http.ResponseWriter, r *http.Request, fs *btrfs.FS) {
	log.Print("RecvBase.")
    err := fs.Recv(".", r.Body)
    if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
}

// http://host/recv

func RecvHandler(w http.ResponseWriter, r *http.Request, fs *btrfs.FS) {
	log.Print("Recv.")
	err := fs.Recv(".", r.Body)
    if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
}

func MasterMux(fs *btrfs.FS) *http.ServeMux {
    mux := http.NewServeMux()

    //  http://host/add/file
    addHandler := func (w http.ResponseWriter, r *http.Request) {
        AddHandler(w, r, fs)
    }

    // http://host/commit/fs
    commitHandler := func (w http.ResponseWriter, r *http.Request) {
        CommitHandler(w, r, fs)
    }

	mux.HandleFunc("/add/", addHandler)
    mux.HandleFunc("/commit", commitHandler)

    return mux;
}

func SlaveMux(fs *btrfs.FS) *http.ServeMux {
    mux := http.NewServeMux()

	// http://host/recvbase/fs
	recvBaseHandler := func (w http.ResponseWriter, r *http.Request) {
		RecvBaseHandler(w, r, fs)
	}

	// http://host/recv/fs

	recvHandler := func (w http.ResponseWriter, r *http.Request) {
        RecvHandler(w, r, fs)
    }

	mux.HandleFunc("/recvbase", recvBaseHandler)
	mux.HandleFunc("/recv", recvHandler)

    return mux;
}

func RunServer(fs *btrfs.FS) {
    if os.Args[1] == "master" {
        http.ListenAndServe(":5656", MasterMux(fs))
    } else if strings.HasPrefix(os.Args[1], "slave") {
        http.ListenAndServe(":5656", SlaveMux(fs))
    } else {
        log.Fatalf("Invalid command %s.", os.Args[1])
    }
}

// usage: pfsd path role shard
func main() {
    log.SetFlags(log.Lshortfile)
	fs := btrfs.ExistingFS(os.Args[2])
    log.Print("Listening on 5656...")
    RunServer(fs)
}
