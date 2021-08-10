// SPDX-FileCopyrightText: Pachyderm, Inc. <info@pachyderm.com>
// SPDX-License-Identifier: Apache-2.0
package helmtest

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"gopkg.in/yaml.v2"
)

func TestSecretHelperGenerate(t *testing.T) {
	values := map[string]string{
		"isTest":                           "true",
		"deployTarget":                     "LOCAL",
		"pachd.clusterDeploymentID":        "",
		"pachd.enterpriseSecret":           "",
		"console.config.oauthClientSecret": "",
	}

	rendered := helm.RenderTemplate(t, &helm.Options{SetStrValues: values}, "../pachyderm", "release-name", []string{"templates/tests/helpersTestTemplate.yaml"})

	renderedAgain := helm.RenderTemplate(t, &helm.Options{SetStrValues: values}, "../pachyderm", "release-name", []string{"templates/tests/helpersTestTemplate.yaml"})

	if rendered != renderedAgain {
		t.Errorf("rendered: %s; rendered again: %s", rendered, renderedAgain)
	}

	renderedData := make(map[interface{}]interface{})
	err := yaml.Unmarshal([]byte(rendered), &renderedData)
	if err != nil {
		t.Error(err)
	}

	renderedAgainData := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(renderedAgain), &renderedAgainData)
	if err != nil {
		t.Error(err)
	}

	if renderedData["enterpriseSecret"] == "" {
		t.Errorf("enterpriseSecret should not render to an empty string")
	}
	if renderedData["consoleSecret"] == "" {
		t.Errorf("consoleSecret should not render to an empty string")
	}
	if renderedData["clusterDeploymentId"] == "" {
		t.Errorf("clusterDeploymentId should not render to an empty string")
	}
}

func TestSecretHelperProvideDefault(t *testing.T) {
	// execute
	values := map[string]string{
		"isTest":                           "true",
		"deployTarget":                     "LOCAL",
		"pachd.clusterDeploymentID":        "ClusterDeploymentID",
		"pachd.enterpriseSecret":           "enterprisePassword",
		"console.config.oauthClientSecret": "OAuthClientSecret",
	}

	rendered := helm.RenderTemplate(t, &helm.Options{SetStrValues: values}, "../pachyderm", "release-name", []string{"templates/tests/helpersTestTemplate.yaml"})

	renderedData := make(map[interface{}]interface{})
	err := yaml.Unmarshal([]byte(rendered), &renderedData)
	if err != nil {
		t.Error(err)
	}

	if renderedData["enterpriseSecret"] != values["pachd.enterpriseSecret"] {
		t.Errorf("rendered enterpriseSecret value: %s, should match the passed in value: %s.",
			renderedData["enterpriseSecret"],
			values["pachd.enterpriseSecret"])
	}
	if renderedData["consoleSecret"] != values["console.config.oauthClientSecret"] {
		t.Errorf("rendered consoleSecret value: %s, should match the passed in value: %s.",
			renderedData["consoleSecret"],
			values["console.config.oauthClientSecret"])
	}
	if renderedData["clusterDeploymentId"] != values["pachd.clusterDeploymentID"] {
		t.Errorf("rendered clusterDeploymentId value: %s, should match the passed in value: %s.",
			renderedData["clusterDeploymentId"],
			values["pachd.clusterDeploymentID"])
	}
}
