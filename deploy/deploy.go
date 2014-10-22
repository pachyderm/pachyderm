package main

import (
    "fmt"
    "log"
    "math/rand"
    "os"
    "strconv"
    "text/template"
    "time"
)

var minPort, maxPort int = 49153, 65535 // Docker uses this range we're just replicating.

type service struct {
    Container, Name string
    Shard, Nshards, Port int
}

func usage() {
    log.Print("Usage:")
    log.Print("$ deploy shard_count")
}

func printShardedService(name string, shards int) {
    sTemplate, err := template.New("server").ParseFiles("templates/server")
    if err != nil { log.Fatal(err) }
    aTemplate, err := template.New("announce").ParseFiles("templates/announce")
    if err != nil { log.Fatal(err) }

    for s := 0; s < shards; s++ {
        config := new(service)
        config.Name = name
        config.Container = "pachyderm/pfs"
        config.Shard = s
        config.Nshards = shards
        config.Port = minPort + rand.Intn(maxPort - minPort)
        server, err := os.Create(fmt.Sprintf("%s-%d-%d.service", config.Name, config.Shard, config.Nshards))
        if err != nil { log.Fatal(err) }
        announce, err := os.Create(fmt.Sprintf("announce-%s-%d-%d.service", config.Name, config.Shard, config.Nshards))
        if err != nil { log.Fatal(err) }

        err = sTemplate.Execute(server, config)
        if err != nil { log.Fatal(err) }
        err = aTemplate.Execute(announce, config)
        if err != nil { log.Fatal(err) }
    }
}

func printGlobalService(name string, shards int) {
    template, err := template.New("global").ParseFiles("templates/global")
    if err != nil { log.Fatal(err) }

    config := new(service)
    config.Name = name
    config.Container = "pachyderm/pfs"
    config.Nshards = shards

    server, err := os.Create(fmt.Sprintf("%s.service", config.Name))
    if err != nil { log.Fatal(err) }

    err = template.Execute(server, config)
    if err != nil { log.Fatal(err) }
}

func printStorage() {
    template, err := template.New("storage").ParseFiles("templates/storage")
    if err != nil { log.Fatal(err) }

    config := new(service)

    storage, err := os.Create("storage.service")
    if err != nil { log.Fatal(err) }

    err = template.Execute(storage, config)
    if err != nil { log.Fatal(err) }
}

func printPull() {
    template, err := template.New("pull").ParseFiles("templates/pull")
    if err != nil { log.Fatal(err) }

    config := new(service)
    config.Container = "pachyderm/pfs"

    storage, err := os.Create("pull.service")
    if err != nil { log.Fatal(err) }

    err = template.Execute(storage, config)
    if err != nil { log.Fatal(err) }
}

func main() {
    log.SetFlags(log.Lshortfile)
    rand.Seed( time.Now().UTC().UnixNano())
    nShards, err := strconv.Atoi(os.Args[1])
    if err != nil { log.Fatal(err) }

    printShardedService("master", nShards)
    printShardedService("replica", nShards)
    printGlobalService("router", nShards)
    printStorage()
    printPull()
}
