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
