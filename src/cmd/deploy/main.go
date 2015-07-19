package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"text/template"
	"time"

	"github.com/pachyderm/pachyderm/src/log"
)

var port = 49153

type service struct {
	Container, Name      string
	Shard, Nshards, Port int
	Disk                 string
}

var outPath string = "/host/home/core/pfs"

func printShardedService(name string) error {
	sTemplate, err := template.New("sharded").Parse(shardedTemplateString)
	if err != nil {
		return err
	}

	for s := 0; s < *shards; s++ {
		for r := 0; r < *replicas; r++ {
			config := new(service)
			config.Name = name
			config.Container = *container
			config.Shard = s
			config.Nshards = *shards
			config.Port = port
			port++
			server, err := os.Create(fmt.Sprintf("%s/%s-%d-%d:%d.service", outPath, config.Name, config.Shard, config.Nshards, r))
			if err != nil {
				return err
			}

			err = sTemplate.Execute(server, config)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func printGlobalService(name string) error {
	template, err := template.New("global").Parse(globalTemplateString)
	if err != nil {
		return err
	}

	config := new(service)
	config.Name = name
	config.Container = *container
	config.Nshards = *shards

	server, err := os.Create(fmt.Sprintf("%s/%s.service", outPath, config.Name))
	if err != nil {
		return err
	}

	return template.Execute(server, config)
}

func printRegistryService(name string, port int) error {
	sTemplate, err := template.New("registry").Parse(registryTemplateString)
	if err != nil {
		return err
	}

	aTemplate, err := template.New("announce").Parse(announceTemplateString)
	if err != nil {
		return err
	}

	config := new(service)
	config.Name = name
	config.Port = port

	server, err := os.Create(fmt.Sprintf("%s/%s.service", outPath, config.Name))
	if err != nil {
		return err
	}

	announce, err := os.Create(fmt.Sprintf("%s/announce-%s.service", outPath, config.Name))
	if err != nil {
		return err
	}

	err = sTemplate.Execute(server, config)
	if err != nil {
		return err
	}

	return aTemplate.Execute(announce, config)
}

func printGitDaemonService(name string) error {
	template, err := template.New("gitdaemon").Parse(gitDaemonTemplateString)
	if err != nil {
		return err
	}

	config := new(service)
	config.Name = name

	server, err := os.Create(fmt.Sprintf("%s/%s.service", outPath, config.Name))
	if err != nil {
		return err
	}

	return template.Execute(server, config)
}

func printStorageService(name string) error {
	template, err := template.New("storage").Parse(storageTemplateString)
	if err != nil {
		return err
	}

	config := new(service)
	config.Name = name
	config.Disk = *disk

	server, err := os.Create(fmt.Sprintf("%s/%s.service", outPath, config.Name))
	if err != nil {
		return err
	}

	return template.Execute(server, config)
}

var shards, replicas *int
var container *string
var disk *string

func main() {
	if err := do(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func do() error {
	rand.Seed(time.Now().UTC().UnixNano())
	shards = flag.Int("shards", 3, "The number of shards in the deploy.")
	replicas = flag.Int("replicas", 3, "The number of replicas of each shard.")
	container = flag.String("container", "pachyderm/pfs", "The container to use for the deploy.")
	disk = flag.String("disk", "/var/lib/pfs/data.img", "The disk to use for pfs' storage.")
	flag.Parse()

	if err := printShardedService("shard"); err != nil {
		return err
	}
	if err := printGlobalService("router"); err != nil {
		return err
	}
	if err := printStorageService("storage"); err != nil {
		return err
	}
	return nil
}

const (
	announceTemplateString = `
[Unit]
Description = pfs announce: {{.Name}}
PartOf = {{.Name}}.service
BindsTo = {{.Name}}.service

[Service]
ExecStart = /bin/sh -c "while true; do etcdctl set /pfs/{{.Name}} '%H:{{.Port}}' --ttl 60;sleep 45;done"
ExecStop = /usr/bin/etcdctl rm /pfs/{{.Name}}

[X-Fleet]
MachineOf = {{.Name}}.service
`

	announceShardedTemplateString = `
[Unit]
Description = pfs announce: {{.Name}}
PartOf = {{.Name}}-{{.Shard}}-{{.Nshards}}.service
BindsTo = {{.Name}}-{{.Shard}}-{{.Nshards}}.service

[Service]
ExecStart = /bin/sh -c "while true; do etcdctl set /pfs/{{.Name}}/{{.Shard}}-{{.Nshards}} '%H:{{.Port}}' --ttl 60;sleep 45;done"
ExecStop = /usr/bin/etcdctl rm /pfs/{{.Name}}/{{.Shard}}-{{.Nshards}}

[X-Fleet]
MachineOf = {{.Name}}-{{.Shard}}-{{.Nshards}}.service
`

	gitDaemonTemplateString = `
[Unit]
Description = pfs service: {{.Name}}

[Service]
ExecStart = /usr/bin/git daemon --base-path=. --export-all --enable=receive-pack --reuseaddr --informative-errors --verbose

[X-Fleet]
Global=true
`

	globalTemplateString = `
[Unit]
Description= pfs service: {{.Name}}
After = docker.service
Requires = docker.service

[Service]
TimeoutStartSec = 300
ExecStartPre = /bin/sh -c "echo $(-docker kill {{.Name}})"
ExecStartPre = /bin/sh -c "echo $(-docker rm {{.Name}})"
ExecStartPre = /bin/sh -c "echo $(docker pull {{.Container}})"
ExecStart = /bin/sh -c "echo $(docker run --name {{.Name}} -p 80:80 -i {{.Container}} /go/src/github.com/pachyderm/pachyderm/etc/bin/launch-wrapper /go/bin/{{.Name}} {{.Nshards}})"
ExecStop = /bin/sh -c "echo $(docker rm -f {{.Name}})"

[X-Fleet]
Global=true
`

	registryTemplateString = `
[Unit]
Description = pfs service: {{.Name}}
After = docker.service
Requires = docker.service

[Service]
ExecStartPre = -/bin/sh -c "echo $(docker kill {{.Name}})"
ExecStartPre = -/bin/sh -c "echo $(docker rm {{.Name}})"
ExecStart = /bin/sh -c "echo $(docker run \
            --name registry \
            -e SETTINGS_FLAVOR=s3 \
            -e AWS_BUCKET=$(etcdctl get /pfs/creds/IMAGE_BUCKET) \
            -e STORAGE_PATH=/registry \
            -e AWS_KEY=$(etcdctl get /pfs/creds/AWS_ACCESS_KEY_ID) \
            -e AWS_SECRET=$(etcdctl get /pfs/creds/AWS_SECRET_ACCESS_KEY) \
            -e SEARCH_BACKEND=sqlalchemy \
            -p {{.Port}}:5000 \
            registry)"
ExecStop = /bin/sh -c "echo $(docker rm -f {{.Name}})"
`

	shardedTemplateString = `
[Unit]
Description = pfs service: {{.Name}}
After = docker.service storage.service
Requires = docker.service

[Service]
TimeoutStartSec = 300
ExecStartPre = -/bin/sh -c "echo $(docker kill {{.Name}}-{{.Shard}}-{{.Nshards}})"
ExecStartPre = -/bin/sh -c "echo $(docker rm {{.Name}}-{{.Shard}}-{{.Nshards}})"
ExecStart = /bin/sh -c "echo $(docker run \
            --privileged=true \
            --name {{.Name}}-{{.Shard}}-{{.Nshards}} \
            -v /:/host:ro \
            -v /var/lib/pfs/vol:/host/var/lib/pfs/vol \
            -v /var/lib/pfs:/var/lib/pfs \
            -v /var/run/docker.sock:/var/run/docker.sock \
            -e AWS_ACCESS_KEY_ID=$(etcdctl get /pfs/creds/AWS_ACCESS_KEY_ID) \
            -e AWS_SECRET_ACCESS_KEY=$(etcdctl get /pfs/creds/AWS_SECRET_ACCESS_KEY) \
            -p {{.Port}}:80 \
            -i {{.Container}} \
            /go/src/github.com/pachyderm/pachyderm/etc/bin/launch-wrapper /go/bin/{{.Name}} -shard {{.Shard}} -modulos {{.Nshards}} -address %H:{{.Port}})"
ExecStop = /bin/sh -c "echo $(docker rm -f {{.Name}}-{{.Shard}}-{{.Nshards}})"
Restart = always
StartLimitInterval = 0
StartLimitBurst = 0


[X-Fleet]
Conflicts={{.Name}}-{{.Shard}}-{{.Nshards}}*
`

	storageTemplateString = `
[Unit]
Description = pfs storage

[Service]
Type = oneshot
RemainAfterExit = yes
ExecStart = /bin/sh -c "echo $(mkdir /var/lib/pfs)"
ExecStart = /bin/sh -c "echo $(truncate /var/lib/pfs/data.img -s 10G)"
ExecStart = /bin/sh -c "echo $(while [ ! -e {{.Disk}} ] ; do sleep 2; done)"
ExecStart = /bin/sh -c "echo $(mkfs.btrfs {{.Disk}})"
ExecStart = /bin/sh -c "echo $(mkdir -p /var/lib/pfs/vol)"
ExecStart = /bin/sh -c "echo $(mount {{.Disk}} /var/lib/pfs/vol)"

[X-Fleet]
Global=true
`
)
