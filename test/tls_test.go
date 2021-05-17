// SPDX-FileCopyrightText: Pachyderm, Inc. <info@pachyderm.com>
// SPDX-License-Identifier: Apache-2.0

package helmtest

import (
	"fmt"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func TestEnablePachTLSNoName(t *testing.T) {
	helmChartPath := "../pachyderm"
	options := &helm.Options{
		SetValues: map[string]string{
			"pachd.storage.backend": "LOCAL",
			"pachd.tls.enabled":     "true",
		},
	}

	_, err := helm.RenderTemplateE(t, options, helmChartPath, "secret", []string{"templates/pachd/deployment.yaml"})
	if err == nil {
		t.Error("Template should error")
	}
}

func TestEnablePachTLSExistingSecret(t *testing.T) {
	helmChartPath := "../pachyderm"
	options := &helm.Options{
		SetValues: map[string]string{
			"pachd.storage.backend": "LOCAL",
			"pachd.tls.enabled":     "true",
			"pachd.tls.secretName":  "blah",
		},
	}

	manifest := helm.RenderTemplate(t, options, helmChartPath, "secret", []string{"templates/pachd/deployment.yaml"})
	fmt.Println(manifest)
}

func TestEnableDashTLSNoName(t *testing.T) {
	helmChartPath := "../pachyderm"
	options := &helm.Options{
		SetValues: map[string]string{
			"pachd.storage.backend": "LOCAL",
			"dash.tls.enabled":      "true",
			"dash.url":              "http://blah.com",
		},
	}

	_, err := helm.RenderTemplateE(t, options, helmChartPath, "secret", []string{"templates/dash/ingress.yaml"})
	if err == nil {
		t.Error("Template should error")
	}
}

func TestEnableDashTLSExistingSecret(t *testing.T) {
	helmChartPath := "../pachyderm"
	options := &helm.Options{
		SetValues: map[string]string{
			"pachd.storage.backend": "LOCAL",
			"dash.tls.enabled":      "true",
			"dash.tls.secretName":   "blah",
			"dash.url":              "http://blah.com",
		},
	}

	manifest := helm.RenderTemplate(t, options, helmChartPath, "secret", []string{"templates/dash/ingress.yaml"})

	fmt.Println(manifest)
}

func TestDashTLSNewSecretNoEnable(t *testing.T) {
	helmChartPath := "../pachyderm"
	options := &helm.Options{
		SetValues: map[string]string{
			"pachd.storage.backend":  "LOCAL",
			"dash.tls.newSecret.crt": "blah",
			"dash.url":               "http://blah.com",
		},
	}

	_, err := helm.RenderTemplateE(t, options, helmChartPath, "secret", []string{"templates/dash/ingress.yaml"})
	if err == nil {
		t.Error("Template should error")
	}
}

func TestPachTLSNewSecretNoEnable(t *testing.T) {
	helmChartPath := "../pachyderm"
	options := &helm.Options{
		SetValues: map[string]string{
			"pachd.storage.backend":   "LOCAL",
			"pachd.tls.newSecret.crt": "blah",
		},
	}

	_, err := helm.RenderTemplateE(t, options, helmChartPath, "deployment", []string{"templates/pachd/deployment.yaml"})
	fmt.Println(err)
	if err == nil {
		t.Error("Template should error")
	}
}

func TestPachTLSNewSecret(t *testing.T) {
	helmChartPath := "../pachyderm"
	secretName := "newsecret1234"
	key := "blah"
	crt := "blah"
	options := &helm.Options{
		SetValues: map[string]string{
			"pachd.storage.backend":      "LOCAL",
			"pachd.tls.enabled":          "true",
			"pachd.tls.newSecret.create": "true",
			"pachd.tls.newSecret.crt":    crt,
			"pachd.tls.newSecret.key":    key,
			"pachd.tls.secretName":       secretName,
		},
	}
	manifests := []string{
		"templates/pachd/tls-secret.yaml",
		"templates/pachd/deployment.yaml",
	}

	renderedManifest := helm.RenderTemplate(t, options, helmChartPath, "deployment", manifests)
	objects, _ := manifestToObjects(renderedManifest)

	for _, obj := range objects {
		switch obj := obj.(type) {
		case *v1.Secret:
			if obj.Name != secretName {
				t.Error("Secret name must match passed in value")
			}
			gotCRT := obj.Data["tls.crt"]
			gotKey := obj.Data["tls.key"]

			if key != string(gotKey) {
				t.Error("Key does not match in secret")
			}
			if crt != string(gotCRT) {
				t.Error("Cert does not match in secret")
			}

		case *appsv1.Deployment:
			for _, c := range obj.Spec.Template.Spec.Containers {
				//Check volume mount pach-tls-cert exists and set correctly
				//Check volume pach-tls-cert and set to correct secret name
				if c.Name == "pachd" {
					want := v1.VolumeMount{
						Name:      "pachd-tls-cert",
						MountPath: "/pachd-tls-cert",
					}

					if !ensureVolumeMountPresent(want, c.VolumeMounts) {
						t.Error("TLS Secret Volume Mount not found in deployment")
					}

				}
			}
			volumes := obj.Spec.Template.Spec.Volumes
			wantVol := v1.Volume{
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: secretName,
					},
				},
				Name: "pachd-tls-cert",
			}

			if !ensureVolumePresent(wantVol, volumes) {
				t.Error("TLS Volume not found")
			}
		}
	}
}
