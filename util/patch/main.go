package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"
	"os/exec"

	jsonpatch "github.com/evanphx/json-patch"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes/scheme"
)

var (
	k8sPodFile = flag.String("pod", "pod.yaml", "k8s yaml/json of pod to patch")
	patchFile  = flag.String("patch", "patch.json", "json patch")
)

func main() {
	flag.Parse()
	k8sPodJSON, err := os.ReadFile(*k8sPodFile)
	if err != nil {
		log.Fatal(err)
	}
	patch, err := os.ReadFile(*patchFile)
	if err != nil {
		log.Fatal(err)
	}

	pod := v1.Pod{}
	if _, _, err := scheme.Codecs.UniversalDeserializer().Decode(k8sPodJSON, &schema.GroupVersionKind{
		Version: "v1",
		Kind:    "Pod",
	}, &pod); err != nil {
		log.Fatal(err)
	}

	jsonPodSpec, err := json.Marshal(&pod.Spec)
	if err != nil {
		log.Fatal(err)
	}

	podSpec := v1.PodSpec{}

	patcher, err := jsonpatch.DecodePatch(patch)
	if err != nil {
		log.Fatal(err)
	}
	jsonPodSpec, err = patcher.Apply(jsonPodSpec)
	if err != nil {
		log.Fatal(err)
	}

	if err := json.Unmarshal(jsonPodSpec, &podSpec); err != nil {
		log.Fatal(err)
	}

	fh, err := os.CreateTemp("", "k8s-pod-*.yaml")
	if err != nil {
		log.Fatal(err)
	}
	name := fh.Name()
	defer func() {
		log.Printf("cleaning up temp file %v", name)
		if err := os.Remove(fh.Name()); err != nil {
			log.Fatal(err)
		}
	}()
	log.Printf("raw output is in %v", name)

	pod.Spec = podSpec
	if err := (&printers.YAMLPrinter{}).PrintObj(&pod, fh); err != nil {
		log.Fatal(err)
	}
	if err := fh.Close(); err != nil {
		log.Fatal(err)
	}
	cmd := exec.Command("diff", "-u", *k8sPodFile, name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		exitErr := new(exec.ExitError)
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() != 1 {
				log.Fatal(err)
			}
			// if 1, there was a diff, which we printed.
		} else {
			log.Fatal(err)
		}
	} else {
		log.Println("patch had no effect")
	}
}
