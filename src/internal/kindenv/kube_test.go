package kindenv

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetConfig(t *testing.T) {
	testKubeClient = fake.NewSimpleClientset(
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
				Annotations: map[string]string{
					clusterVersionKey:                  "1",
					clusterRegistryPullKey:             "pull",
					clusterRegistryPushKey:             "push",
					clusterHostnameKey:                 "pachyderm.example",
					tlsKey:                             "false",
					exposedPortsKey:                    "1,2,3,4,5,6,7,8,9,10",
					portBindingPrefix + pachydermProxy: "80,443",
				},
			},
		},
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-namespace-1",
				Annotations: map[string]string{
					tlsKey:                             "true",
					exposedPortsKey:                    "2,3,4,5,6,7,8,9,10",
					portBindingPrefix + pachydermProxy: "1",
				},
			},
		},
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-namespace-2",
				Annotations: map[string]string{
					tlsKey:                             "false",
					exposedPortsKey:                    "12,13,14,15,16,17,18,19,20",
					portBindingPrefix + pachydermProxy: "11",
				},
			},
		},
	)
	t.Cleanup(func() { testKubeClient = nil })

	cluster := &Cluster{
		name:       "test",
		kubeconfig: testSentinel,
	}
	t.Cleanup(func() {
		if err := cluster.Close(); err != nil {
			t.Fatalf("close cluster: %v", err)
		}
	})

	testData := []struct {
		namespace string
		want      *ClusterConfig
	}{
		{
			namespace: "default",
			want: &ClusterConfig{
				Version:       1,
				ImagePushPath: "push",
				ImagePullPath: "pull",
				Hostname:      "pachyderm.example",
				TLS:           false,
			},
		},
		{
			namespace: "test-namespace-1",
			want: &ClusterConfig{
				Version:       1,
				ImagePushPath: "push",
				ImagePullPath: "pull",
				Hostname:      "pachyderm.example",
				TLS:           true,
			},
		},
		{
			namespace: "test-namespace-2",
			want: &ClusterConfig{
				Version:       1,
				ImagePushPath: "push",
				ImagePullPath: "pull",
				Hostname:      "pachyderm.example",
				TLS:           false,
			},
		},
	}
	for _, test := range testData {
		t.Run(test.namespace, func(t *testing.T) {
			ctx := pctx.TestContext(t)
			got, err := cluster.GetConfig(ctx, test.namespace)
			if err != nil {
				t.Errorf("GetConfig: %v\n", err)
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("GetConfig: config: -want +got\n%s", diff)
			}
		})
	}
}

func TestAllocatePort(t *testing.T) {
	testKubeClient = fake.NewSimpleClientset(
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
				Annotations: map[string]string{
					clusterVersionKey:                  "1",
					clusterRegistryPullKey:             "pull",
					clusterRegistryPushKey:             "push",
					clusterHostnameKey:                 "pachyderm.example",
					tlsKey:                             "false",
					exposedPortsKey:                    "1,2,3,4,5,6,7,8,9,10",
					portBindingPrefix + pachydermProxy: "80,443",
				},
			},
		},
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-namespace-1",
				Annotations: map[string]string{
					tlsKey:                             "true",
					exposedPortsKey:                    "12,13,14,15,16,17,18,19,20",
					portBindingPrefix + pachydermProxy: "11",
				},
			},
		},
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-namespace-2",
				Annotations: map[string]string{
					tlsKey:                             "false",
					exposedPortsKey:                    "22,23,24,25,26,27,28,29,30",
					portBindingPrefix + pachydermProxy: "21",
				},
			},
		},
	)
	t.Cleanup(func() { testKubeClient = nil })

	cluster := &Cluster{
		name:       "test",
		kubeconfig: testSentinel,
	}
	t.Cleanup(func() {
		if err := cluster.Close(); err != nil {
			t.Fatalf("close cluster: %v", err)
		}
	})

	testData := []struct {
		name, namespace, service, want string
		wantAnnotations                map[string]string
	}{
		{
			name:      "default proxy port",
			namespace: "default",
			service:   pachydermProxy,
			want:      "80,443",
			wantAnnotations: map[string]string{
				clusterVersionKey:                  "1",
				clusterRegistryPullKey:             "pull",
				clusterRegistryPushKey:             "push",
				clusterHostnameKey:                 "pachyderm.example",
				tlsKey:                             "false",
				exposedPortsKey:                    "1,2,3,4,5,6,7,8,9,10",
				portBindingPrefix + pachydermProxy: "80,443",
			},
		},
		{
			name:      "namespace-1 proxy port",
			namespace: "test-namespace-1",
			service:   pachydermProxy,
			want:      "11",
			wantAnnotations: map[string]string{
				tlsKey:                             "true",
				exposedPortsKey:                    "12,13,14,15,16,17,18,19,20",
				portBindingPrefix + pachydermProxy: "11",
			},
		},
		{
			name:      "namespace-2 proxy port",
			namespace: "test-namespace-2",
			service:   pachydermProxy,
			want:      "21",
			wantAnnotations: map[string]string{
				tlsKey:                             "false",
				exposedPortsKey:                    "22,23,24,25,26,27,28,29,30",
				portBindingPrefix + pachydermProxy: "21",
			},
		},
		{
			name:      "default foo port",
			namespace: "default",
			service:   "foo",
			want:      "1",
			wantAnnotations: map[string]string{
				clusterVersionKey:                  "1",
				clusterRegistryPullKey:             "pull",
				clusterRegistryPushKey:             "push",
				clusterHostnameKey:                 "pachyderm.example",
				tlsKey:                             "false",
				exposedPortsKey:                    "2,3,4,5,6,7,8,9,10",
				portBindingPrefix + pachydermProxy: "80,443",
				portBindingPrefix + "foo":          "1",
			},
		},
		{
			name:      "default bar port",
			namespace: "default",
			service:   "bar",
			want:      "2",
			wantAnnotations: map[string]string{
				clusterVersionKey:                  "1",
				clusterRegistryPullKey:             "pull",
				clusterRegistryPushKey:             "push",
				clusterHostnameKey:                 "pachyderm.example",
				tlsKey:                             "false",
				exposedPortsKey:                    "3,4,5,6,7,8,9,10",
				portBindingPrefix + pachydermProxy: "80,443",
				portBindingPrefix + "foo":          "1",
				portBindingPrefix + "bar":          "2",
			},
		},
		{
			name:      "namespace-1 foo port",
			namespace: "test-namespace-1",
			service:   "foo",
			want:      "12",
			wantAnnotations: map[string]string{
				tlsKey:                             "true",
				exposedPortsKey:                    "13,14,15,16,17,18,19,20",
				portBindingPrefix + pachydermProxy: "11",
				portBindingPrefix + "foo":          "12",
			},
		},
	}
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			ctx := pctx.TestContext(t)
			got, err := cluster.AllocatePort(ctx, test.namespace, test.service)
			if err != nil {
				t.Errorf("AllocatePort(%v, %v): %v", test.namespace, test.service, err)
			}
			if want := test.want; got != want {
				t.Errorf("AllocatePort(%v, %v):\n  got: %v\n want: %v", test.namespace, test.service, got, want)
			}
			ns, err := testKubeClient.CoreV1().Namespaces().Get(ctx, test.namespace, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("Namespaces().Get(%v): %v", test.namespace, err)
			}
			if diff := cmp.Diff(test.wantAnnotations, ns.GetAnnotations()); diff != "" {
				t.Errorf("annotations(%s): diff (-want +got)\n:%v", test.namespace, diff)
			}
		})
	}
}
