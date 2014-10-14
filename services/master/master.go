package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
    "fmt"
    "github.com/coreos/go-etcd/etcd"
    "pfs/lib/btrfs"
    "path"
    "strconv"
)

func headPath() string { return ".commits/HEAD" }

func IncrCommit(commit string) (string, error) {
	split := strings.Split(commit, "/")
	index, err := strconv.Atoi(split[len(split) - 1])
	if err != nil { return "", err }
	index++
    split[len(split) - 1] = strconv.Itoa(index)
	return strings.Join(split, "/"), nil
}

func DecrCommit(commit string) (string, error) {
	split := strings.Split(commit, "/")
	index, err := strconv.Atoi(split[len(split) - 1])
	if err != nil { return "", err }
	index--
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
	log.Printf("Getting slaves for %s.", os.Args[1])
	shard_prefix := path.Join("/pfsd", os.Args[1])
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

func BrowseHandler(w http.ResponseWriter, r *http.Request, fs *btrfs.FS) {
    client := etcd.NewClient([]string{"http://172.17.42.1:4001"})
	log.Printf("Getting slaves for %s.", os.Args[1])
	shard_prefix := path.Join("/pfsd", os.Args[1])
    slaves, err := client.Get(shard_prefix, false, false)
	log.Printf("Got slaves.")
	log.Print(slaves)
    if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }

    exists, err := fs.FileExists(".commits")
    if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
    if exists {
        commits, err := fs.ReadDir(".commits")
        if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }

        for _, c := range commits {
            fmt.Fprintf(w, "<html>")
            fmt.Fprintf(w, "<pre>")
            fmt.Fprintf(w, "<a href=\"/pfs/.commits/%s\">%s</a> - %s <a href=\"/del/%s\">Delete</a>\n", c.Name(), c.Name(), c.ModTime().Format("Jan 2, 2006 at 3:04pm (PST)"), c.Name());
            fmt.Fprintf(w, "</pre>")
            fmt.Fprintf(w, "</html>")
        }
    } else {
        fmt.Fprint(w, "Nothing here :(.");
    }
}

// http://host/del

func DelCommitHandler(w http.ResponseWriter, r *http.Request, fs *btrfs.FS) {
    url := strings.Split(r.URL.String(), "/")
    commit := path.Join(".commits", url[2])
    client := etcd.NewClient([]string{"http://172.17.42.1:4001"})
	log.Printf("Getting slaves for %s.", os.Args[1])
	shard_prefix := path.Join("/pfs", os.Args[1])
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
        err = fs.SubvolumeDelete(commit)
        if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
        if commit == last_commit {
            //TODO this is actually broken if commits have already been deleted
            //from the middle. Or if you try to delete 0 (which would just be
            //disallowed).
            err = fs.Remove(headPath())
            if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
            prev_head, err := DecrCommit(commit)
            err = fs.Symlink(prev_head, headPath())
            if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
        }

        for _, s := range slaves.Node.Nodes {
			log.Print("Key: ", s.Key)
			log.Print(path.Join(shard_prefix, "master"))
			if s.Key != path.Join(shard_prefix, "master") {
				log.Print(s.Value)
				log.Print("Sending update to:" + s.Value)
                _, err = http.Get("http://" + s.Value + "/del/" + url[2])
				if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
				//fmt.Fprintf(w, "Deleted commit from: %s.\n", s.Value)
			}
        }
    } else {
        http.Error(w, "Commit not found.", 500)
        log.Print(err)
    }

    err = btrfs.Sync()
    if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
    http.Redirect(w, r, "/browse", 200)
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

    browseHandler := func (w http.ResponseWriter, r *http.Request) {
        BrowseHandler(w, r, fs)
    }

    delCommitHandler := func (w http.ResponseWriter, r *http.Request) {
        DelCommitHandler(w, r, fs)
    }

	mux.HandleFunc("/add/", addHandler)
    mux.HandleFunc("/commit", commitHandler)
    mux.Handle("/pfs/", http.StripPrefix("/pfs/", http.FileServer(http.Dir("/mnt/pfs/master"))))
    mux.HandleFunc("/browse", browseHandler)
    mux.HandleFunc("/del/", delCommitHandler)
	mux.HandleFunc("/ping", func (w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "pong") })

    fmt.Printf("This has the /pfs/ route in it!!!")

    return mux;
}

func RunServer(fs *btrfs.FS) {
    http.ListenAndServe(":80", MasterMux(fs))
}

// usage: pfsd path role shard
func main() {
    log.SetFlags(log.Lshortfile)
	fs := btrfs.ExistingFS("pfs", "master-" + os.Args[1])
    fs.EnsureNamespace()
    log.Print("Listening on port 80...")
    RunServer(fs)
}
