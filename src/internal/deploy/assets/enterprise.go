package assets

import (
	"strconv"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// EnterpriseService returns a service for the Enterprise Server
func EnterpriseService(opts *AssetOpts) *v1.Service {
	prometheusAnnotations := map[string]string{
		"prometheus.io/scrape": "true",
		"prometheus.io/port":   strconv.Itoa(PrometheusPort),
	}
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: objectMeta(enterpriseServerName, labels(enterpriseServerName), prometheusAnnotations, opts.Namespace),
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeNodePort,
			Selector: map[string]string{
				"app": enterpriseServerName,
			},
			Ports: []v1.ServicePort{
				// NOTE: do not put any new ports before `api-grpc-port`, as
				// it'll change k8s SERVICE_PORT env var values
				{
					Port:     650, // also set in cmd/pachd/main.go
					Name:     "api-grpc-port",
					NodePort: 31650,
				},
				{
					Port:     651, // also set in cmd/pachd/main.go
					Name:     "trace-port",
					NodePort: 31651,
				},
				{
					Port:     OidcPort,
					Name:     "oidc-port",
					NodePort: 31000 + OidcPort,
				},
				{
					Port:     658,
					Name:     "identity-port",
					NodePort: 31658,
				},
				{
					Port:       656,
					Name:       "prometheus-metrics",
					NodePort:   31656,
					Protocol:   v1.ProtocolTCP,
					TargetPort: intstr.FromInt(656),
				},
			},
		},
	}
}
