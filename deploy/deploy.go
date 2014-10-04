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
    Name string
    Shard, Nshards, Port int
}

func usage() {
    log.Print("Usage:")
    log.Print("$ deploy shard_count")
}

func printShardedService(name string, shards int) {
    sTemplate, err := template.New("service").ParseFiles("templates/service")
    if err != nil { log.Fatal(err) }
    aTemplate, err := template.New("announce").ParseFiles("templates/announce")
    if err != nil { log.Fatal(err) }

    nShards, err := strconv.Atoi(os.Args[1])
    if err != nil { log.Fatal(err) }

    for s := 1; s <= nShards; s++ {
        config := new(service)
        config.Name = name
        config.Shard = s
        config.Nshards = nShards
        config.Port = minPort + rand.Intn(maxPort - minPort)
        server, err := os.Create(fmt.Sprintf("%s-%d-%d.service", config.Name, config.Shard, config.Nshards))
        if err != nil { log.Fatal(err) }
        announce, err := os.Create(fmt.Sprintf("announce-%s.%d.%d.service", config.Name, config.Shard, config.Nshards))
        if err != nil { log.Fatal(err) }

        err = sTemplate.Execute(server, config)
        if err != nil { log.Fatal(err) }
        err = aTemplate.Execute(announce, config)
        if err != nil { log.Fatal(err) }
    }
}

func main() {
    log.SetFlags(log.Lshortfile)
    rand.Seed( time.Now().UTC().UnixNano())
    nShards, err := strconv.Atoi(os.Args[1])
    if err != nil { log.Fatal(err) }

    printShardedService("master", nShards)
    printShardedService("slave", nShards)
}
