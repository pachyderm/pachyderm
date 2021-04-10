// SPDX-FileCopyrightText: Pachyderm, Inc. <info@pachyderm.com>
// SPDX-License-Identifier: Apache-2.0

package helmtest

import (
	"strconv"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	v1 "k8s.io/api/core/v1"
)

func TestAWS(t *testing.T) {
	objects, err := manifestToObjects(helm.RenderTemplate(t,
		&helm.Options{
			SetStrValues: map[string]string{
				"pachd.storage.backend": "AMAZON",
			},
		}, "../pachyderm", "release-name", nil))
	if err != nil {
		t.Error(err)
	}
	for _, obj := range objects {
		switch obj := obj.(type) {
		case *v1.Secret:
			if _, err := strconv.Atoi(string(obj.Data["part-size"])); err != nil {
				t.Error(err)
			}
		}
	}
}
