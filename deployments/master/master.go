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

type templateData struct {
    Shard, Nshards, Port int
}

func usage() {
    log.Print("Usage:")
    log.Print("$ master shard_count")
}

func main() {
    log.SetFlags(log.Lshortfile)
    rand.Seed(time.Now().UTC().UnixNano())

    pfsdService, err := template.New("master.service").ParseFiles("templates/master.service")
    if err != nil { log.Fatal(err) }
    pfsdAnnounce, err := template.New("announce-master.service").ParseFiles("templates/announce-master.service")
    if err != nil { log.Fatal(err) }

    nShards, err := strconv.Atoi(os.Args[1])
    if err != nil { log.Fatal(err) }
    log.Printf("Creating sharded masters with %d shards.", nShards);

    for s := 1; s <= nShards; s++ {
        config := new(templateData)
        config.Shard = s
        config.Nshards = nShards
        config.Port = minPort + rand.Intn(maxPort - minPort)
        master, err := os.Create(fmt.Sprintf("tmp/master-%d-%d.service", config.Shard, config.Nshards))
        if err != nil { log.Fatal(err) }
        announce, err := os.Create(fmt.Sprintf("tmp/announce-master.%d.%d.service", config.Shard, config.Nshards))
        if err != nil { log.Fatal(err) }

        err = pfsdService.Execute(master, config)
        if err != nil { log.Fatal(err) }
        err = pfsdAnnounce.Execute(announce, config)
        if err != nil { log.Fatal(err) }
    }
}
