package k8s

import (
	"testing"

	ourstar "github.com/pachyderm/pachyderm/v2/src/internal/starlark"
	"github.com/pachyderm/pachyderm/v2/src/internal/starlark/startest"
	"go.starlark.net/starlark"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	dfake "k8s.io/client-go/dynamic/fake"
	kfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/kubectl/pkg/scheme"
)

func TestClient(t *testing.T) {
	objects := []runtime.Object{
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-pod-abc123",
				Namespace: "default",
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "test",
					},
				},
			},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
			},
		},
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-test-pod",
				Namespace: "default",
				Labels: map[string]string{
					"suite": "test",
				},
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "test",
					},
				},
			},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
			},
		},
	}
	sc := kfake.NewSimpleClientset(objects...)
	sc.Fake.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{
					Name:         "pods",
					SingularName: "pod",
					Namespaced:   true,
					Kind:         "Pod",
					Verbs:        metav1.Verbs{"get", "list", "watch"},
				},
			},
		},
	}
	dc := dfake.NewSimpleDynamicClient(scheme.Scheme, objects...)
	module, err := NewClientset("default", sc, dc)
	if err != nil {
		t.Fatal(err)
	}
	startest.RunTest(t, "k8s_test.star", ourstar.Options{
		Predefined: starlark.StringDict{
			"k8s": module,
		},
	})
}
