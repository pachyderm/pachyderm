// SPDX-FileCopyrightText: Pachyderm, Inc. <info@pachyderm.com>
// SPDX-License-Identifier: Apache-2.0

package helmtest

import (
	"path"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	appsV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func TestLocal(t *testing.T) {
	var (
		hostPath     = "/this/is/a/host/path/"
		secret       = "secret-name"
		objects, err = manifestToObjects(helm.RenderTemplate(t,
			&helm.Options{
				SetStrValues: map[string]string{
					"deployTarget":                 "LOCAL",
					"pachd.storage.local.hostPath": hostPath,
					"global.imagePullSecrets[0]":   secret,
					"pachd.enterpriseLicenseKey":   "licenseKey",
					"pachd.oidc.clientId":          "pachd",
					"pachd.oidc.issuer":            "http://issuer-ip:1658",
					"pachd.oidc.redirectURI":       "http://localhost:30657/authorization/callback",
				},
				SetValues: map[string]string{
					"pachd.oidc.mockIDP": "true",
				},
			},
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
			if expected, got := 1, len(object.Spec.Template.Spec.ImagePullSecrets); expected != got {
				t.Errorf("expected %d image pull secret; got %d`", expected, got)
			}
			for _, s := range object.Spec.Template.Spec.ImagePullSecrets {
				if expected, got := secret, s; expected != got.Name {
					t.Errorf("expected secret %q; got %q", expected, got)
				}
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

		case *v1.Secret:
			if object.Name == "pachyderm-bootstrap-config" {
				enterpriseSecret := object.StringData["enterpriseSecret"]
				if !strings.Contains(object.StringData["enterpriseClusters"], enterpriseSecret) {
					t.Errorf("enterprise secret %s should be present in the enterpriseClusters config %v", enterpriseSecret, object.StringData["enterpriseClusters"])
				}
				if !strings.Contains(object.StringData["enterpriseConfig"], enterpriseSecret) {
					t.Errorf("enterprise secret %s should be present in the enterpriseConfig %v", enterpriseSecret, object.StringData["enterpriseConfig"])
				}
			}
		}
	}
	for check := range checks {
		if !checks[check] {
			t.Errorf("%q incomplete", check)
		}
	}
}
