// SPDX-FileCopyrightText: 2021 Pachyderm, Inc. <info@pachyderm.com>
// SPDX-License-Identifier: Apache-2.0

package helmtest

import (
	"path"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	appsV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func TestLocal(t *testing.T) {
	var (
		hostPath     = "/this/is/a/host/path/"
		objects, err = manifestToObjects(helm.RenderTemplate(t,
			&helm.Options{
				SetStrValues: map[string]string{
					"pachd.storage.backend":        "LOCAL",
					"pachd.storage.local.hostPath": hostPath,
				}},
			"../pachyderm/", "release-name", nil))
		checks = map[string]bool{
			"STORAGE_BACKEND":   false,
			"STORAGE_HOST_PATH": false,
			"headless service":  false,
		}
	)
	if err != nil {
		t.Fatalf("could not render templates to objects: %v", err)
	}
	for _, object := range objects {
		switch object := object.(type) {
		case *appsV1.Deployment:
			if object.Name != "pachd" {
				continue
			}
			for _, c := range object.Spec.Template.Spec.Containers {
				if c.Name != "pachd" {
					continue
				}
				for _, e := range c.Env {
					switch e.Name {
					case "STORAGE_BACKEND":
						if e.Value != "LOCAL" {
							t.Errorf("expected STORAGE_BACKEND to be %q, not %q", "GOOGLE", e.Value)
						}
						checks["STORAGE_BACKEND"] = true
					case "STORAGE_HOST_PATH":
						if e.Value != path.Join(hostPath, "pachd") {
							t.Errorf("expected STORAGE_HOST_PATH to be %q, not %q", path.Join(hostPath, "pachd"), e.Value)
						}
						checks["STORAGE_HOST_PATH"] = true
					}
				}
			}
		case *v1.Service:
			if object.Name != "etcd-headless" {
				continue
			}
			checks["headless service"] = true
		}
	}
	for check := range checks {
		if !checks[check] {
			t.Errorf("%q incomplete", check)
		}
	}
}
