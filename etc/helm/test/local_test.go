// SPDX-FileCopyrightText: Pachyderm, Inc. <info@pachyderm.com>
// SPDX-License-Identifier: Apache-2.0

package helmtest

import (
	"path"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	appsV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func TestLocalDefaultHostPathIsStableAndShared(t *testing.T) {
	options := &helm.Options{
		SetStrValues: map[string]string{
			"deployTarget": "LOCAL",
		},
	}
	templates := []string{"templates/pachd/deployment.yaml", "templates/pachw/deployment.yaml"}
	var renders [][]string
	for i := 0; i < 2; i++ {
		objects, err := manifestToObjects(helm.RenderTemplate(t, options, "../pachyderm/", "release-name", templates))
		if err != nil {
			t.Fatalf("could not render templates to objects: %v", err)
		}
		var paths []string
		for _, object := range objects {
			deployment, ok := object.(*appsV1.Deployment)
			if !ok {
				continue
			}
			for _, volume := range deployment.Spec.Template.Spec.Volumes {
				if volume.Name == "pach-disk" && volume.HostPath != nil {
					paths = append(paths, volume.HostPath.Path)
				}
			}
			for _, c := range deployment.Spec.Template.Spec.Containers {
				for _, e := range c.Env {
					if e.Name == "STORAGE_HOST_PATH" {
						paths = append(paths, e.Value)
					}
				}
			}
		}
		if expected, got := 4, len(paths); expected != got {
			t.Fatalf("expected %d local storage paths (host path volume and STORAGE_HOST_PATH of pachd and pachw); got %d: %v", expected, got, paths)
		}
		renders = append(renders, paths)
	}
	want := renders[0][0]
	for _, paths := range renders {
		for _, got := range paths {
			if got != want {
				t.Errorf("expected every local storage path to be %q; got %q (all renders: %v)", want, got, renders)
			}
		}
	}
}

func TestLocal(t *testing.T) {
	var (
		expectedStorageBackend = "LOCAL"
		hostPath               = "/this/is/a/host/path/"
		secret                 = "secret-name"
		objects, err           = manifestToObjects(helm.RenderTemplate(t,
			&helm.Options{
				SetStrValues: map[string]string{
					"deployTarget":                 expectedStorageBackend,
					"pachd.storage.local.hostPath": hostPath,
					"global.imagePullSecrets[0]":   secret,
					"pachd.enterpriseLicenseKey":   "licenseKey",
				},
				SetValues: map[string]string{
					"oidc.mockIDP": "true",
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
							t.Errorf("expected STORAGE_BACKEND to be %q, not %q", expectedStorageBackend, e.Value)
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
