package main

import (
	"log"
	"testing"
    "crypto/rand"
    "fmt"
    "pfs/filesystem"
    "io"
    "io/ioutil"
    "net/http"
    "net/http/httptest"
    "path"
    "strings"
    "sync"
)

func Test(t *testing.T) {
    log.SetFlags(log.Lshortfile)
	fs := filesystem.NewFS("pfs_test", "loop2", "pfs", 10)
    fs.EnsureExists()
    defer fs.Destroy()
    server := httptest.NewServer(MasterMux(fs))
    defer server.Close()

    res, err := http.Get(server.URL + "/" + path.Join("init", "test"))
    if err != nil { t.Fatal(err) }
    if res.StatusCode != 200 { t.Fatal("Error status code.", res.StatusCode) }

    res, err = http.Post(server.URL + "/" + path.Join("add", "test", "foo"), "text/plain", strings.NewReader("foo"))
    if err != nil { t.Fatal(err) }
    if res.StatusCode != 200 { t.Fatal("Error status code.", res.StatusCode) }

    res, err = http.Get(server.URL + "/" + path.Join("commit", "test"))
    if err != nil { t.Fatal(err) }
    if res.StatusCode != 200 { t.Fatal("Error status code.", res.StatusCode) }

    res, err = http.Get(server.URL + "/" + path.Join("cat", "test", "foo"))
    if err != nil { t.Fatal(err) }
    if res.StatusCode != 200 { t.Fatal("Error status code.", res.StatusCode) }

    val, err := ioutil.ReadAll(res.Body)
    res.Body.Close()
    if err != nil { t.Fatal(err) }
    if string(val) != "foo" { t.Fatal("Received an incorrect value: ", val) }
}

func TestStress(t *testing.T) {
    log.SetFlags(log.Lshortfile)
	fs1 := filesystem.NewFS("pfs_test2", "loop3", "pfs2", 10)
    fs2 := filesystem.NewFS("pfs_test3", "loop4", "pfs3", 10)
    fs1.EnsureExists()
    fs2.EnsureExists()
    defer fs1.Destroy()
    defer fs2.Destroy()
    server1 := httptest.NewServer(MasterMux(fs1))
    server2 := httptest.NewServer(ReplicaMux(fs2))
    defer server1.Close()
    defer server2.Close()

    /* Initialize a filesystem on both servers. */
    res, err := http.Get(server1.URL + "/" + path.Join("init", "test"))
    if err != nil { t.Fatal(err) }
    if res.StatusCode != 200 { t.Fatal("Error status code.", res.StatusCode) }

    send_res, err := http.Get(server1.URL + "/" + path.Join("sendbase", "test@0"))
    if err != nil { t.Fatal(err) }
    if send_res.StatusCode != 200 { t.Fatal("Error status code.", send_res.StatusCode) }

    res, err = http.Post(server2.URL + "/" + path.Join("recvbase", "test"),
            "text/plain", send_res.Body)
    if err != nil { t.Fatal(err) }
    if res.StatusCode != 200 { t.Fatal("Error status code.", res.StatusCode) }

    for i := 0; i < 5; i++ {
        var wg sync.WaitGroup
        for j := 0; j < 10; j++ {
            filename := fmt.Sprintf("%d%d", i, j)
            wg.Add(1)
            go func () {
                defer wg.Done()
                reader := io.LimitReader(rand.Reader, 1024 * 1024)
                res, err = http.Post(server1.URL + "/" + path.Join("add", "test", filename), "text/plain", reader)
                if err != nil { t.Fatal(err) }
                if res.StatusCode != 200 { t.Fatal("Error status code.", res.StatusCode) }
            }()
        }
        wg.Wait()

        res, err = http.Get(server1.URL + "/" + path.Join("commit", "test"))
        if err != nil { t.Fatal(err) }
        if res.StatusCode != 200 { t.Fatal("Error status code.", res.StatusCode) }

        send_res, err := http.Get(server1.URL + "/" + path.Join("send", fmt.Sprintf("test@%d", i), fmt.Sprintf("test@%d", i + 1)))
        if err != nil { t.Fatal(err) }
        if send_res.StatusCode != 200 { t.Fatal("Error status code.", send_res.StatusCode) }

        res, err = http.Post(server2.URL + "/" + path.Join("recv", "test"),
                "text/plain", send_res.Body)
        if err != nil { t.Fatal(err) }
        if res.StatusCode != 200 { t.Fatal("Error status code.", res.StatusCode) }
    }

    for i := 0; i < 5; i++ {
        for j := 0; j < 10; j++ {
            res, err = http.Get(server1.URL + "/" + path.Join("cat", "test", fmt.Sprintf("%d%d", i, j)))
            if err != nil { t.Fatal(err) }
            if res.StatusCode != 200 { t.Fatal("Error status code.", res.StatusCode) }

            // res, err = http.Get(server2.URL + "/" + path.Join("cat", "test", fmt.Sprintf("%d%d", i, j)))
            // if err != nil { t.Fatal(err) }
            // if res.StatusCode != 200 { t.Fatal("Error status code.", res.StatusCode) }
        }
    }
}
