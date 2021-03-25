package helmtest

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	appsV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

/*
pachd:
  image:
    tag: 1.12.3
  storage:
    backend: GOOGLE
    google:
      googleBucket: "fake-bucket"
      googleCred: "fake-creds"

*/
func TestGoogleServiceAccount(t *testing.T) {
	helmChartPath := "../pachyderm"

	expectedServiceAccount := "my-fine-sa"
	options := &helm.Options{
		SetValues: map[string]string{
			"pachd.image.tag":                         "1.12.3",
			"pachd.storage.backend":                   "GOOGLE",
			"pachd.storage.google.bucket":             "fake-bucket",
			"pachd.storage.google.serviceAccountName": expectedServiceAccount,
		},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "blah", []string{"templates/pachd/rbac/serviceaccount.yaml"})

	var serviceAccount v1.ServiceAccount

	helm.UnmarshalK8SYaml(t, output, &serviceAccount)

	manifestServiceAccount := serviceAccount.Annotations["iam.gke.io/gcp-service-account"]
	if manifestServiceAccount != expectedServiceAccount {
		t.Fatalf("Google service account expected (%s) actual (%s) ", expectedServiceAccount, manifestServiceAccount)
	}
}

func TestGoogleWorkerServiceAccount(t *testing.T) {
	helmChartPath := "../pachyderm"

	expectedServiceAccount := "my-fine-sa"
	options := &helm.Options{
		SetValues: map[string]string{
			"pachd.image.tag":                         "1.12.3",
			"pachd.storage.backend":                   "GOOGLE",
			"pachd.storage.google.bucket":             "fake-bucket",
			"pachd.storage.google.serviceAccountName": expectedServiceAccount,
		},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "blah", []string{"templates/pachd/rbac/worker-serviceaccount.yaml"})

	var serviceAccount v1.ServiceAccount

	helm.UnmarshalK8SYaml(t, output, &serviceAccount)

	manifestServiceAccount := serviceAccount.Annotations["iam.gke.io/gcp-service-account"]
	if manifestServiceAccount != expectedServiceAccount {
		t.Fatalf("Google service account expected (%s) actual (%s) ", expectedServiceAccount, manifestServiceAccount)
	}
}

func TestGoogleValues(t *testing.T) {
	var (
		bucket              = "fake-bucket"
		cred                = `INSERT JSON HERE`
		pachdServiceAccount = "128"
		serviceAccount      = "a-service-account"
		helmChartPath       = "../pachyderm"
		checks              = map[string]bool{
			"bucket":          false,
			"cred":            false,
			"service account": false,
			"STORAGE_BACKEND": false,
			"GOOGLE_BUCKET":   false,
			"GOOGLE_CRED":     false,
		}

		options = &helm.Options{
			SetStrValues: map[string]string{
				"pachd.serviceAccount.name":               pachdServiceAccount,
				"pachd.image.tag":                         "1.12.3",
				"pachd.storage.backend":                   "GOOGLE",
				"pachd.storage.google.bucket":             bucket,
				"pachd.storage.google.cred":               cred,
				"pachd.storage.google.serviceAccountName": serviceAccount,
				// this certificate was generated for the examples FIXME: generate one for tests
				"tls.crt": `
    -----BEGIN CERTIFICATE-----
    MIICzDCCAbSgAwIBAgIJAPfq8ZWr7H6zMA0GCSqGSIb3DQEBBQUAMBcxFTATBgNV
    BAMTDGZha2UuZXhhbXBsZTAeFw0yMTAzMTYxMjQxMTdaFw0zMTAzMTQxMjQxMTda
    MBcxFTATBgNVBAMTDGZha2UuZXhhbXBsZTCCASIwDQYJKoZIhvcNAQEBBQADggEP
    ADCCAQoCggEBANCEbJS5qgvPUsMpwcV085R1bB4mBkqQs19eOQjU+CZQQPEdJa2/
    rcNnE1xqpNdhvqi7uTQ2AA5jIXG3igrq1HbDpnqcvePQoWGsvT7G26wSqcJsL3ab
    VrSbz9exlmCVxABtu/B1NFhHfRTb6Qeipaa0fPoibWfPKszvlmpJNSv8NzoaUpM5
    j6lfeytAvQ1yC0R5VcodRpsPaOgzV/xvMNd73fQ8HB3vWBR43RdIcUNZt4Plpt/G
    5nmWvNBQZhxTrEPGi7pNRbdfJFU6FGM3zjZ2TjaQ3Z1+AFhcgCmYAk/sknsYDfG1
    YaR/QYMs4PHNRjLbEniHHR+DXgh5QRpA3RUCAwEAAaMbMBkwFwYDVR0RBBAwDoIM
    ZmFrZS5leGFtcGxlMA0GCSqGSIb3DQEBBQUAA4IBAQCzk7BYoeBOmbv81x0SBbQ2
    8QH+tFfvgUDYH5suYlV4VhXTj6s/zbTbyhNHn8hZTqGgvmdH7AFgLRBNdQLaC+LM
    J9srlwyORG3/0yJ4+cagWiJBaFLdyrCRTueDWzvP8whbdRz4EKDyXmbfnK4X12xp
    0iwaXMsmYLSWc6HFrffF4TIFNqpGmjtax8JSWlM3XUzFNegO3CfmpxT24vVPVhkv
    mIp8Cb/3yloIIwBzbEDq3oOGaOQaQtZXJQDSvM0Bks/FsSz6qbiq9W8QsP7KP9Jc
    W9erM+ku5QK8I62yLpJH9XaWaNS82yVoozs/pyj/obSTbFxgKapSD02knFXzelCs
    -----END CERTIFICATE-----
`,
				// this key was generated for the examples FIXME: generate one for tests
				"tls.key": `
    -----BEGIN RSA PRIVATE KEY-----
    MIIEowIBAAKCAQEA0IRslLmqC89SwynBxXTzlHVsHiYGSpCzX145CNT4JlBA8R0l
    rb+tw2cTXGqk12G+qLu5NDYADmMhcbeKCurUdsOmepy949ChYay9PsbbrBKpwmwv
    dptWtJvP17GWYJXEAG278HU0WEd9FNvpB6KlprR8+iJtZ88qzO+Wakk1K/w3OhpS
    kzmPqV97K0C9DXILRHlVyh1Gmw9o6DNX/G8w13vd9DwcHe9YFHjdF0hxQ1m3g+Wm
    38bmeZa80FBmHFOsQ8aLuk1Ft18kVToUYzfONnZONpDdnX4AWFyAKZgCT+ySexgN
    8bVhpH9Bgyzg8c1GMtsSeIcdH4NeCHlBGkDdFQIDAQABAoIBABOBv/Kt59WRAKoX
    VvRU+5CQ55tubTo+jTlHxEgqPEjBS0IDOwolG2ljVDFaHK+1ijOY1DupLZoq9e8A
    f56D13qA1Ss1TKJqWx6bHV0pF1XirRTuMAaFg7gDt47zIyFIAX0Uxvc4z7vOfEoe
    RI+dTKfqzKJN5DRI8jUX2Nd6n8nMdEDRZYVHvwLh6soJ58Lkdt5faOLrnK6U+yie
    8r0SObsuW4CZY1Is0gcD0WHM2gnBacFLtgw+Ec3CSn0Scp0peAQUOhONDWnC6ESH
    YJxWFtOGBrFnuOqNgVEzCKQrPJKdOg4yLDo6TVv8T2xUlfTM9k5czkweVjRJhZtJ
    jrElLQECgYEA9lzedeN0HGTIvhy9lJW+GqYz4OxSP7kxJoHL0jAtFJayNhswFFcD
    AQkRQkQ9ogHOCD+OCcxtG03CGMQeaVx1vSsfAHj+nytyBwfXXNWmONCxJHgVw0/e
    72UBqfF8+HxEb7UyWzVytVf9q2MrXQnCf26HClin0OekURGeZ4+phVUCgYEA2KyP
    5piWC9CDS1wQbHQE3wEitJzNLElkmJlXToNwYhedExOXlDen5STMV9LS3VgIYrNO
    v0Ze2BSnE2SikmpuxlIOttCa5ZAL7a7wraCUaogD9BzUce986oGu2sMSR9bWqiq5
    4SZ2jM21w1xXb2N5LBfABj+pwRsycJmGjBkl+MECgYAE4R3+07xu+4gGS+dtU/Hp
    8TTB1axjWrWgf52b0hxydfGdpLg1DuweTyGqYFOgK8z62NdlVkkq60VW3DuF9rDW
    SE5a4gqY+HFPtlYLnqemJGv9vusfbSuLLkL0LLY+7aclVz9iExLsiIubo2EufIz/
    nR7Lk6nvN0dH28N5ZZ0D6QKBgHpX7aUKMWcYZJpPsKJcXEfDL2KGSz+fbWLQ6sBV
    bUamCLY10NgLGQ1E2vEYBKKgy5NXpbZROMqP1ssXfshnuobW3KITZfMLhADAT/vp
    +QOyK3FSOg7faExNz3qMvSy9PVa2a2CbREM7AFAAOwqVQ11HR9D/b42vGqsDtTo0
    FQHBAoGBAJ/tsZrh95LsnQjuNtq78O2uChcYIkZl2KQY8fk+GQo+NBffkhafAlyX
    hGMZufzcFJ+/LeMFpzmxsZOck8EUqJ2gNkofvt/SMgUhGHqVg57jzX3opKjVnbpW
    hUpG89j0CbfXSYYioe3Z2GKLzzuLtnKCszSYkoAgifqOpxaR3k92
    -----END RSA PRIVATE KEY-----
`,
			},
		}
		output = helm.RenderTemplate(t, options, helmChartPath, "release-name", nil)
	)
	objects, err := manifestToObjects(output)
	if err != nil {
		t.Fatal(err)
	}

	for _, object := range objects {
		switch resource := object.(type) {
		case *v1.Secret:
			if resource.Name != "pachyderm-storage-secret" {
				continue
			}
			if b := string(resource.Data["google-bucket"]); b != bucket {
				t.Errorf("expected bucket to be %q but was %q", bucket, b)
			}
			checks["bucket"] = true
			if c := string(resource.Data["google-cred"]); c != cred {
				t.Errorf("expected cred to be %q but was %q", cred, c)
			}
			checks["cred"] = true
		case *v1.ServiceAccount:
			if resource.Name != pachdServiceAccount {
				continue
			}
			if sa := resource.Annotations["iam.gke.io/gcp-service-account"]; sa != serviceAccount {
				t.Errorf("expected service account to be %q but was %q", serviceAccount, sa)
			}
			checks["service account"] = true
		case *appsV1.Deployment:
			if resource.Name != "pachd" {
				continue
			}
			for _, c := range resource.Spec.Template.Spec.Containers {
				if c.Name != "pachd" {
					continue
				}
				for _, e := range c.Env {
					switch e.Name {
					case "STORAGE_BACKEND":
						if e.Value != "GOOGLE" {
							t.Errorf("expected STORAGE_BACKEND to be %q, not %q", "GOOGLE", e.Value)
						}
						checks["STORAGE_BACKEND"] = true
					case "GOOGLE_BUCKET":
						checks["GOOGLE_BUCKET"] = true
					case "GOOGLE_CRED":
						checks["GOOGLE_CRED"] = true
					}
				}
			}
		}
	}
	for check := range checks {
		if !checks[check] {
			t.Errorf("check %q not performed", check)
		}
	}
}
