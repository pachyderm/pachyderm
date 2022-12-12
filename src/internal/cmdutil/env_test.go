package cmdutil

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/api/resource"
)

type Cornucopia struct {
	String   string            `env:"STRING,default=hello"`
	Int      int               `env:"INT,default=42"`
	Resource resource.Quantity `env:"RESOURCE,default=1Gi"`
}

func TestPopulate(t *testing.T) {
	os.Setenv("STRING", "string")
	os.Setenv("RESOURCE", "1Mi")
	t.Cleanup(func() {
		os.Unsetenv("STRING")
		os.Unsetenv("RESOURCE")
	})
	want := Cornucopia{
		String:   "string",
		Int:      42,
		Resource: resource.MustParse("1Mi"),
	}

	var got Cornucopia
	if err := Populate(&got); err != nil {
		t.Errorf("populate: %v", err)
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("parsed value:\n%s", diff)
	}
}

func TestPopulateDefaults(t *testing.T) {
	want := Cornucopia{
		String:   "hello",
		Int:      42,
		Resource: resource.MustParse("1Gi"),
	}

	var got Cornucopia
	if err := PopulateDefaults(&got); err != nil {
		t.Errorf("populate: %v", err)
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("parsed value:\n%s", diff)
	}
}
