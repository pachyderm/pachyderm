// SPDX-FileCopyrightText: Pachyderm, Inc. <info@pachyderm.com>
// SPDX-License-Identifier: Apache-2.0

package helmtest

import (
	"fmt"
	"testing"

	helm "github.com/gruntwork-io/terratest/modules/helm"
	v1 "k8s.io/api/core/v1"
)

const (
	ChartDir           = "../pachyderm"
	ReleaseName        = "config-secret"
	TemplatePath       = "templates/pachd/config-secret.yaml"
	ExpectedSecretName = "pachyderm-bootstrap-config"
	ExpectedLicense    = "$ENTERPRISE_LICENSE_KEY"
)

type Expectation func(*testing.T, *v1.Secret)

func RenderSecret(
	t *testing.T,
	setValuesOverride map[string]string,
	extraHelmArgs []string,
) *v1.Secret {
	options := helm.Options{
		SetValues: map[string]string{
			"deployTarget":                         "LOCAL",
			"pachd.activateEnterprise":             "true",
			"pachd.enterpriseLicenseKeySecretName": "shh",
		},
	}

	for k, v := range setValuesOverride {
		options.SetValues[k] = v
	}

	output, err := helm.RenderTemplateE(
		t,
		&options,
		ChartDir,
		ReleaseName,
		[]string{TemplatePath},
		extraHelmArgs...,
	)
	if err != nil {
		t.Fatal(err)
	}

	var secret *v1.Secret
	helm.UnmarshalK8SYaml(t, output, &secret)
	return secret
}

func CheckCommon(t *testing.T, secret *v1.Secret) {
	if ExpectedSecretName != secret.Name {
		t.Errorf("Unexpected secret name '%s'", secret.Name)
	}
	if ExpectedLicense != secret.StringData["license"] {
		t.Errorf("Unexpected secret name '%s'", secret.Name)
	}
}

func TestConfigSecretNamespace(t *testing.T) {
	setValuesOverride := map[string]string{}
	expectedNamespace := "foo-namespace"
	extraHelmArgs := []string{
		fmt.Sprintf("--namespace=%s", expectedNamespace),
	}

	secret := RenderSecret(t, setValuesOverride, extraHelmArgs)
	CheckCommon(t, secret)

	if expectedNamespace != secret.Namespace {
		t.Errorf("Unexpected namespace '%s'", secret.Namespace)
	}
}

type TestCase struct {
	setValuesOverride map[string]string
	extraHelmArgs     []string
	expectations      []Expectation
}

func TestConfigSecretEnterpriseSecret(t *testing.T) {
	testCases := []TestCase{
		{
			map[string]string{
				"pachd.enterpriseSecretSecretName": "shhhhh!",
			},
			[]string{},
			[]Expectation{
				func(t *testing.T, secret *v1.Secret) {
					expectedEnterpriseSecret := "$ENTERPRISE_SECRET"
					enterpriseSecret := secret.StringData["enterpriseSecret"]
					if expectedEnterpriseSecret != enterpriseSecret {
						t.Errorf("Unexpected enterpriseSecret '%s'", enterpriseSecret)
					}
				},
			},
		},
		{
			map[string]string{
				"pachd.enterpriseSecretSecretName": "shhhhh!",
			},
			[]string{},
			[]Expectation{
				func(t *testing.T, secret *v1.Secret) {
					expectedEnterpriseSecret := "$ENTERPRISE_SECRET"
					enterpriseSecret := secret.StringData["enterpriseSecret"]
					if expectedEnterpriseSecret != enterpriseSecret {
						t.Errorf("Unexpected enterpriseSecret '%s'", enterpriseSecret)
					}
				},
			},
		},
	}

	for _, testCase := range testCases {
		secret := RenderSecret(t, testCase.setValuesOverride, testCase.extraHelmArgs)
		CheckCommon(t, secret)
		for _, expectation := range testCase.expectations {
			expectation(t, secret)
		}
	}
}

// rootTokenLength := len(secret.StringData["rootToken"])
// if rootTokenLength != 32 {
// 	t.Errorf("Unexpected root token length '%d'", rootTokenLength)
// }
