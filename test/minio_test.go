// SPDX-FileCopyrightText: 2021 Pachyderm, Inc. <info@pachyderm.com>
// SPDX-License-Identifier: Apache-2.0

package helmtest

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	appsV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func TestMinio(t *testing.T) {
	var (
		bucket       = "foo-bucket"
		endpoint     = "https://endpoint.test/"
		id           = "bar-id"
		secret       = "baz-secret"
		secure       = "quux-secure"
		signature    = "quux-sig"
		objects, err = manifestToObjects(helm.RenderTemplate(t,
			&helm.Options{
				SetStrValues: map[string]string{
					"pachd.storage.backend":         "MINIO",
					"pachd.storage.minio.bucket":    bucket,
					"pachd.storage.minio.endpoint":  endpoint,
					"pachd.storage.minio.id":        id,
					"pachd.storage.minio.secret":    secret,
					"pachd.storage.minio.secure":    secure,
					"pachd.storage.minio.signature": signature,
				}},
			"../pachyderm/", "release-name", nil))
		checks = map[string]bool{
			"bucket":          false,
			"endpoint":        false,
			"id":              false,
			"secret":          false,
			"secure":          false,
			"signature":       false,
			"STORAGE_BACKEND": false,
			"MINIO_BUCKET":    false,
			"MINIO_ENDPOINT":  false,
			"MINIO_ID":        false,
			"MINIO_SECRET":    false,
			"MINIO_SECURE":    false,
			"MINIO_SIGNATURE": false,
		}
	)
	if err != nil {
		t.Fatalf("could not render templates to objects: %v", err)
	}
	for _, object := range objects {
		switch object := object.(type) {
		case *v1.Secret:
			if object.Name != "pachyderm-storage-secret" {
				continue
			}
			if s := string(object.Data["minio-bucket"]); s != bucket {
				t.Errorf("expected bucket to be %q but was %q", bucket, s)
			}
			checks["bucket"] = true
			if s := string(object.Data["minio-endpoint"]); s != endpoint {
				t.Errorf("expected endpoint to be %q but was %q", endpoint, s)
			}
			checks["endpoint"] = true
			if s := string(object.Data["minio-id"]); s != id {
				t.Errorf("expected id to be %q but was %q", id, s)
			}
			checks["id"] = true
			if s := string(object.Data["minio-secret"]); s != secret {
				t.Errorf("expected secret to be %q but was %q", secret, s)
			}
			checks["secret"] = true
			if s := string(object.Data["minio-secure"]); s != secure {
				t.Errorf("expected secure to be %q but was %q", secure, s)
			}
			checks["secure"] = true
			if s := string(object.Data["minio-signature"]); s != signature {
				t.Errorf("expected signature to be %q but was %q", signature, s)
			}
			checks["signature"] = true
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
						if e.Value != "MINIO" {
							t.Errorf("expected STORAGE_BACKEND to be %q, not %q", "GOOGLE", e.Value)
						}
						checks["STORAGE_BACKEND"] = true
					case "MINIO_BUCKET":
						checks["MINIO_BUCKET"] = true
					case "MINIO_ENDPOINT":
						checks["MINIO_ENDPOINT"] = true
					case "MINIO_ID":
						checks["MINIO_ID"] = true
					case "MINIO_SECRET":
						checks["MINIO_SECRET"] = true
					case "MINIO_SECURE":
						checks["MINIO_SECURE"] = true
					case "MINIO_SIGNATURE":
						checks["MINIO_SIGNATURE"] = true
					}
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
