package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

//go:embed load.yaml
var spec string

type containerEditor func(c *v1.Container)

func cpuAndRAM(cpu, ram string) containerEditor {
	return func(c *v1.Container) {
		r := v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(cpu),
			v1.ResourceMemory: resource.MustParse(ram),
		}
		c.Resources.Requests = r
		c.Resources.Limits = r
	}
}

func debugEnvironment() containerEditor {
	return func(c *v1.Container) {
		var godebug, diskCache, memCache bool
		for i, e := range c.Env {
			if e.Name == "GODEBUG" {
				godebug = true
				e.Value = "gctrace=1"
				c.Env[i] = e
				continue
			}
			if e.Name == "STORAGE_DISK_CACHE_SIZE" {
				diskCache = true
				e.Value = "0"
				c.Env[i] = e
				continue
			}
			if e.Name == "STORAGE_MEMORY_CACHE_SIZE" {
				memCache = true
				e.Value = "200"
				c.Env[i] = e
				continue
			}
		}
		if !godebug {
			c.Env = append(c.Env, v1.EnvVar{
				Name:  "GODEBUG",
				Value: "gctrace=1",
			})
		}
		if !diskCache {
			c.Env = append(c.Env, v1.EnvVar{
				Name:  "STORAGE_DISK_CACHE_SIZE",
				Value: "0",
			})
		}
		if !memCache {
			c.Env = append(c.Env, v1.EnvVar{
				Name:  "STORAGE_MEMORY_CACHE_SIZE",
				Value: "400",
			})
		}
	}
}

func combine(fs ...containerEditor) containerEditor {
	return func(c *v1.Container) {
		for _, f := range fs {
			f(c)
		}
	}
}

var scenarios = map[string]func(*v1.Container){
	"1 CPU, 32G RAM": combine(debugEnvironment(), cpuAndRAM("1", "32G")),
	"8 CPU, 32G RAM": combine(debugEnvironment(), cpuAndRAM("8", "32G")),
	"1 CPU, 8G RAM":  combine(debugEnvironment(), cpuAndRAM("1", "8G")),
	"8 CPU, 8G RAM":  combine(debugEnvironment(), cpuAndRAM("8", "8G")),
	"1 CPU, 1G RAM":  combine(debugEnvironment(), cpuAndRAM("1", "1G")),
	"8 CPU, 1G RAM":  combine(debugEnvironment(), cpuAndRAM("8", "1G")),
}

func main() {
	cfg, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		log.Fatal(err)
	}
	k8s, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	results := map[string][]time.Duration{}

	for name, f := range scenarios {
		log.Printf("running scenario %q", name)
		if err := editPachd(context.TODO(), k8s, f); err != nil {
			log.Printf("error editing pachd deployment: %v", err)
			continue
		}

		for i := 0; i < 3; i++ {
			var pachyderm *client.APIClient
			for j := 0; j < 10; j++ {
				var err error
				if pachyderm != nil {
					pachyderm.Close()
				}
				pachyderm, err = client.NewOnUserMachine("")
				if err != nil {
					log.Printf("retry %v: error creating client: %v", j, err)
					time.Sleep(5 * time.Second)
					continue
				}
				log.Printf("connected to pachyderm ok")
				time.Sleep(20 * time.Second)
				break
			}

			log.Printf("starting run %v of scenario %q", i, name)
			res, err := pachyderm.PfsAPIClient.RunLoadTest(pachyderm.Ctx(), &pfs.RunLoadTestRequest{Spec: spec})
			if err != nil {
				log.Printf("error starting load test: %v", err)
				continue
			}
			if err := res.GetError(); err != "" {
				log.Printf("error running load test: %v", err)
				continue
			}
			d, err := types.DurationFromProto(res.GetDuration())
			if err != nil {
				log.Printf("duration from proto: %v", err)
			}
			results[name] = append(results[name], d)
			log.Printf("RESULT: scenario %q: run %d: %s\n", name, i, d)

			if err := pachyderm.DeleteRepo("load_test", false); err != nil {
				log.Printf("delete repo load_test: %v", err)
			}
		}
	}
	log.Printf("RESULTS:")
	for name, ds := range results {
		var avg float64
		var entries []string
		for _, d := range ds {
			avg += d.Seconds()
			entries = append(entries, d.String())
		}
		avg /= float64(len(ds))
		log.Printf("%v: %v [%v]", name, avg, entries)
	}
}

func editPachd(ctx context.Context, k8s kubernetes.Interface, f func(*v1.Container)) error {
	pachd, err := k8s.AppsV1().Deployments("default").Get(ctx, "pachd", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get pachd: %w", err)
	}
	old := pachd.DeepCopy()

	for i, c := range pachd.Spec.Template.Spec.Containers {
		if c.Name == "pachd" {
			f(&pachd.Spec.Template.Spec.Containers[i])
		}
	}
	pachd, err = k8s.AppsV1().Deployments("default").Update(ctx, pachd, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("update pachd: %w", err)
	}

	if false {
		diff := cmp.Diff(old.Spec.Template.Spec, pachd.Spec.Template.Spec)
		log.Printf("diff: %s", diff)
	}

	log.Printf("waiting for rollout")
	if err := exec.CommandContext(ctx, "kubectl", "rollout", "status", "deployment", "pachd").Run(); err != nil {
		return fmt.Errorf("wait for update: %w", err)
	}
	log.Printf("...rollout done")

	return nil
}
