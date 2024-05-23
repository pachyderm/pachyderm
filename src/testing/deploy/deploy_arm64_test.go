//go:build k8s

package main

func init() {
	globalValueOverrides = map[string]string{
		"postgresql.image.repository": "pachyderm/postgresql",
		"postgresql.image.tag":        "13.3.0",
		"etcd.image.tag":              "v3.5.5",
		"pgbouncer.image.tag":         "1.16.1",
		"kubeEventTail.image.tag":     "0.0.7",
	}
}
