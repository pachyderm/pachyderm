// SPDX-FileCopyrightText: Pachyderm, Inc. <info@pachyderm.com>
// SPDX-License-Identifier: Apache-2.0

package helmtest

import (
	"errors"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/instrumenta/kubeval/kubeval"
)

func TestDashImageTag(t *testing.T) {

	helmChartPath := "../pachyderm"

	options := &helm.Options{
		SetValues: map[string]string{
			"dash.image.tag":              "0.5.5",
			"pachd.storage.backend":       "GOOGLE",
			"pachd.storage.google.bucket": "bucket",
		},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "deployment", []string{"templates/dash/deployment.yaml"})

	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(t, output, &deployment)

	expectedContainerImage := "pachyderm/dash:0.5.5"
	deploymentContainers := deployment.Spec.Template.Spec.Containers
	if deploymentContainers[0].Image != expectedContainerImage {
		t.Fatalf("Rendered container image (%s) is not expected (%s)", deploymentContainers[0].Image, expectedContainerImage)
	}
}

func TestEtcdImageTag(t *testing.T) {

	helmChartPath := "../pachyderm"

	options := &helm.Options{
		SetValues: map[string]string{
			"etcd.image.tag":              "blah",
			"pachd.storage.backend":       "GOOGLE",
			"pachd.storage.google.bucket": "bucket",
		},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "statefulset", []string{"templates/etcd/statefulset.yaml"})

	var statefulSet appsv1.StatefulSet
	helm.UnmarshalK8SYaml(t, output, &statefulSet)

	expectedContainerImage := "pachyderm/etcd:blah"
	deploymentContainers := statefulSet.Spec.Template.Spec.Containers
	if deploymentContainers[0].Image != expectedContainerImage {
		t.Fatalf("Rendered container image (%s) is not expected (%s)", deploymentContainers[0].Image, expectedContainerImage)
	}
}

func TestPachdImageTag(t *testing.T) {
	helmChartPath := "../pachyderm"

	options := &helm.Options{
		SetValues: map[string]string{
			"pachd.image.tag":             "blah1234",
			"pachd.storage.backend":       "GOOGLE",
			"pachd.storage.google.bucket": "bucket",
		},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "deployment", []string{"templates/pachd/deployment.yaml"})

	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(t, output, &deployment)

	expectedContainerImage := "pachyderm/pachd:blah1234"
	deploymentContainers := deployment.Spec.Template.Spec.Containers
	if deploymentContainers[0].Image != expectedContainerImage {
		t.Fatalf("Rendered container image (%s) is not expected (%s)", deploymentContainers[0].Image, expectedContainerImage)
	}
}

func TestPachdImageTagDeploymentEnv(t *testing.T) {
	helmChartPath := "../pachyderm"

	expectedTag := "blah1234"
	expectedPachdContainerImage := "pachyderm/pachd:" + expectedTag
	expectedWorkerContainerImage := "pachyderm/worker:" + expectedTag
	options := &helm.Options{
		SetValues: map[string]string{
			"pachd.image.tag":             expectedTag,
			"pachd.storage.backend":       "GOOGLE",
			"pachd.storage.google.bucket": "bucket",
		},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "deployment", []string{"templates/pachd/deployment.yaml"})

	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(t, output, &deployment)

	deploymentContainers := deployment.Spec.Template.Spec.Containers

	//fmt.Printf("%+v\n", deploymentContainers[0].Env)
	pachdVersionEnvVar := "PACHD_VERSION"
	err, version := GetEnvVarByName(deploymentContainers[0].Env, pachdVersionEnvVar)

	if err != nil {
		t.Fatalf("Could not find env var")
	}
	workerImageEnvVar := "WORKER_IMAGE"
	err, workerImage := GetEnvVarByName(deploymentContainers[0].Env, workerImageEnvVar)
	if err != nil {
		t.Fatalf("Could not find env var")
	}
	workerSidecarImageEnvVar := "WORKER_SIDECAR_IMAGE"
	err, workerSidecarImage := GetEnvVarByName(deploymentContainers[0].Env, workerSidecarImageEnvVar)
	if err != nil {
		t.Fatalf("Could not find env var")
	}
	if deploymentContainers[0].Image != expectedPachdContainerImage {
		t.Fatalf("Rendered Pachd Image (%s) is not expected (%s)", deploymentContainers[0].Image, expectedPachdContainerImage)
	}
	if version != expectedTag {
		t.Fatalf("Rendered EnvVar (%s) value (%s) is not expected (%s)", workerImageEnvVar, version, expectedTag)
	}
	if workerImage != expectedWorkerContainerImage {
		t.Fatalf("Rendered EnvVar (%s) value (%s) is not expected (%s)", workerImageEnvVar, workerImage, expectedWorkerContainerImage)
	}

	if workerSidecarImage != expectedPachdContainerImage {
		t.Fatalf("Rendered EnvVar (%s) value (%s) is not expected (%s)", workerSidecarImageEnvVar, workerImage, expectedPachdContainerImage)
	}

	res, err := kubeval.Validate([]byte(output))
	if err != nil {
		t.Fatalf("could not validate output: %v", err)
	}
	for _, res := range res {
		if !res.ValidatedAgainstSchema {
			t.Errorf("file %s failed to validate: %v", res.FileName, res.Errors)
		}
	}
}

func TestSetNamespaceWorkerRoleBinding(t *testing.T) {

	helmChartPath := "../pachyderm"

	expectedNamespace := "kspace"

	options := &helm.Options{
		SetValues: map[string]string{
			"pachd.storage.backend":       "GOOGLE",
			"pachd.storage.google.bucket": "bucket",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", expectedNamespace),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "rolebinding", []string{"templates/pachd/rbac/worker-rolebinding.yaml"})

	var rolebinding rbacv1.RoleBinding
	helm.UnmarshalK8SYaml(t, output, &rolebinding)

	subjects := rolebinding.Subjects
	if subjects[0].Namespace != expectedNamespace {
		t.Fatalf("Rendered namespace (%s) is not expected (%s)", subjects[0].Namespace, expectedNamespace)
	}

}

func TestSetNamespaceServiceAccount(t *testing.T) {

	helmChartPath := "../pachyderm"

	expectedNamespace := "kspace1234"

	options := &helm.Options{
		SetValues: map[string]string{
			"pachd.storage.backend":       "GOOGLE",
			"pachd.storage.google.bucket": "bucket",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", expectedNamespace),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "serviceaccount", []string{"templates/pachd/rbac/serviceaccount.yaml"})

	var serviceAccount v1.ServiceAccount
	helm.UnmarshalK8SYaml(t, output, &serviceAccount)

	metadata := serviceAccount.ObjectMeta

	if metadata.Namespace != expectedNamespace {
		t.Fatalf("Rendered namespace (%s) is not expected (%s)", metadata.Namespace, expectedNamespace)
	}
}

func TestSetNamespaceWorkerServiceAccount(t *testing.T) {

	helmChartPath := "../pachyderm"

	expectedNamespace := "kspaceworker"

	options := &helm.Options{
		SetValues: map[string]string{
			"pachd.storage.backend":       "GOOGLE",
			"pachd.storage.google.bucket": "bucket",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", expectedNamespace),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "workerserviceaccount", []string{"templates/pachd/rbac/worker-serviceaccount.yaml"})

	var serviceAccount v1.ServiceAccount
	helm.UnmarshalK8SYaml(t, output, &serviceAccount)

	metadata := serviceAccount.ObjectMeta

	if metadata.Namespace != expectedNamespace {
		t.Fatalf("Rendered namespace (%s) is not expected (%s)", metadata.Namespace, expectedNamespace)
	}
}

func TestSetNamespaceClusterRoleBinding(t *testing.T) {

	helmChartPath := "../pachyderm"

	expectedNamespace := "kspace123"

	options := &helm.Options{
		SetValues: map[string]string{
			"pachd.storage.backend":       "GOOGLE",
			"pachd.storage.google.bucket": "bucket",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", expectedNamespace),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "clusterrolebinding", []string{"templates/pachd/rbac/clusterrolebinding.yaml"})

	var rolebinding rbacv1.ClusterRoleBinding
	helm.UnmarshalK8SYaml(t, output, &rolebinding)

	subjects := rolebinding.Subjects
	if subjects[0].Namespace != expectedNamespace {
		t.Fatalf("Rendered namespace (%s) is not expected (%s)", subjects[0].Namespace, expectedNamespace)
	}

}

func TestAmazonStorageSecretsAmazonRegion(t *testing.T) {
	/*
		amazonRegion:
		amazonBucket:
		amazonID:
		amazonSecret:
		amazonToken:
		amazonDistribution:
		customEndpoint:
		retries:
		timeout:
		uploadACL:
		reverse:
		partSize:
		maxUploadParts:
		disableSSL:
		noVerifySSL:
		#TODO Vault env vars
		#TODO iamRole - Check all places IAM role rendered
	*/
	helmChartPath := "../pachyderm"
	amazonRegion := "blah"
	options := &helm.Options{
		SetValues: map[string]string{
			"pachd.storage.backend":             "AMAZON",
			"pachd.storage.amazon.amazonRegion": amazonRegion},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "secret", []string{"templates/pachd/storage-secret.yaml"})

	var secret v1.Secret
	helm.UnmarshalK8SYaml(t, output, &secret)

	//TODO Check Result

	//fmt.Printf("%+v\n", blah)
	//fmt.Print(secret.Data["amazon-region"])
}

func TestMicrosoftStorageSecrets(t *testing.T) {

	helmChartPath := "../pachyderm"
	microsoftContainer := "blah"
	options := &helm.Options{
		SetValues: map[string]string{
			"pachd.storage.backend":             "MICROSOFT",
			"pachd.storage.microsoft.container": microsoftContainer},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "secret", []string{"templates/pachd/storage-secret.yaml"})

	var secret v1.Secret
	helm.UnmarshalK8SYaml(t, output, &secret)

	//TODO Check Result

	//fmt.Printf("%+v\n", blah)
}

func TestGoogleStorageSecrets(t *testing.T) {

	helmChartPath := "../pachyderm"
	googleBucket := "blah"
	options := &helm.Options{
		SetValues: map[string]string{
			"pachd.storage.backend":       "GOOGLE",
			"pachd.storage.google.bucket": googleBucket},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "secret", []string{"templates/pachd/storage-secret.yaml"})

	var secret v1.Secret
	helm.UnmarshalK8SYaml(t, output, &secret)

	//TODO Check result

	//fmt.Printf("%+v\n", blah)
}

func TestMinioStorageSecrets(t *testing.T) {

	helmChartPath := "../pachyderm"
	minioBucket := "blah"
	options := &helm.Options{
		SetValues: map[string]string{
			"pachd.storage.backend":      "MINIO",
			"pachd.storage.minio.bucket": minioBucket},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "secret", []string{"templates/pachd/storage-secret.yaml"})

	var secret v1.Secret
	helm.UnmarshalK8SYaml(t, output, &secret)

	// TODO Check result

	//fmt.Printf("%+v\n", blah)
}

func GetEnvVarByName(envVars []v1.EnvVar, name string) (error, string) {
	for _, v := range envVars {
		if v.Name == name {
			return nil, v.Value
		}
	}
	return errors.New("Not found"), ""
}

type service struct {
	name     string
	selector map[string]string
	ports    []string
}

func (s service) match(c container) bool {
	for k, v := range s.selector {
		if c.labels[k] != v {
			return false
		}
	}
	return true
}

type container struct {
	name   string
	index  int
	labels map[string]string
	ports  map[string]bool
}

func TestServicePorts(t *testing.T) {
	var (
		hostPath     = "/this/is/a/host/path/"
		objects, err = manifestToObjects(helm.RenderTemplate(t,
			&helm.Options{
				SetStrValues: map[string]string{
					"pachd.storage.backend":        "LOCAL",
					"pachd.storage.local.hostPath": hostPath,
				}},
			"../pachyderm/", "release-name", nil))
		services   []service
		containers []container
	)
	if err != nil {
		t.Fatalf("could not render templates to objects: %v", err)
	}
	for _, object := range objects {
		switch object := object.(type) {
		case *v1.Service:
			var ports []string
			for _, p := range object.Spec.Ports {
				if p.TargetPort.Type == intstr.String {
					ports = append(ports, p.TargetPort.StrVal)
				}
			}
			if len(ports) == 0 {
				continue
			}
			services = append(services, service{
				name:     object.Name,
				selector: object.Spec.Selector,
				ports:    ports,
			})
		case *appsv1.Deployment:
			for i, c := range object.Spec.Template.Spec.Containers {
				var ports = make(map[string]bool)
				for _, p := range c.Ports {
					if p.Name != "" {
						ports[p.Name] = true
					}
				}
				if len(ports) == 0 {
					continue
				}
				containers = append(containers, container{
					name:   object.Name,
					index:  i,
					labels: object.Spec.Template.Labels,
					ports:  ports,
				})
			}
		case *appsv1.StatefulSet:
			for i, c := range object.Spec.Template.Spec.Containers {
				var ports = make(map[string]bool)
				for _, p := range c.Ports {
					if p.Name != "" {
						ports[p.Name] = true
					}
				}
				if len(ports) == 0 {
					continue
				}
				containers = append(containers, container{
					name:   object.Name,
					index:  i,
					labels: object.Spec.Template.Labels,
					ports:  ports,
				})
			}
		case *v1.ReplicationController:
			for i, c := range object.Spec.Template.Spec.Containers {
				var ports = make(map[string]bool)
				for _, p := range c.Ports {
					if p.Name != "" {
						ports[p.Name] = true
					}
				}
				if len(ports) == 0 {
					continue
				}
				containers = append(containers, container{
					name:   object.Name,
					index:  i,
					labels: object.Spec.Template.Labels,
					ports:  ports,
				})
			}
		case *appsv1.DaemonSet:
			for i, c := range object.Spec.Template.Spec.Containers {
				var ports = make(map[string]bool)
				for _, p := range c.Ports {
					if p.Name != "" {
						ports[p.Name] = true
					}
				}
				if len(ports) == 0 {
					continue
				}
				containers = append(containers, container{
					name:   object.Name,
					index:  i,
					labels: object.Spec.Template.Labels,
					ports:  ports,
				})
			}
		}
	}
	for _, s := range services {
		var satisfied = make(map[string]bool)
		for _, c := range containers {
			if !s.match(c) {
				continue
			}
			for _, p := range s.ports {
				if c.ports[p] {
					satisfied[p] = true
				}
			}
		}
		for _, p := range s.ports {
			if !satisfied[p] {
				t.Errorf("nothing satisfies service %s/%s", s.name, p)
			}
		}
	}
}

func TestGOMAXPROCS(t *testing.T) {
	var (
		hostPath     = "/this/is/a/host/path/"
		objects, err = manifestToObjects(helm.RenderTemplate(t,
			&helm.Options{
				SetStrValues: map[string]string{
					"pachd.storage.backend":        "LOCAL",
					"pachd.storage.local.hostPath": hostPath,
				}},
			"../pachyderm/", "release-name", nil))
		sawGoMaxProcs bool
	)
	if err != nil {
		t.Fatalf("could not render templates to objects: %v", err)
	}
	for _, object := range objects {
		switch object := object.(type) {
		case *appsv1.Deployment:
			for _, e := range object.Spec.Template.Spec.Containers[0].Env {
				if e.Name == "GOMAXPROCS" {
					t.Error("GOMAXPROCS exists when it should not")
				}
			}
		}
	}

	objects, err = manifestToObjects(helm.RenderTemplate(t,
		&helm.Options{
			ValuesFiles: []string{"../examples/hub-values.yaml"},
		},
		"../pachyderm/", "release-name", nil))
	if err != nil {
		t.Fatalf("could not render templates to objects: %v", err)
	}
	for _, object := range objects {
		switch object := object.(type) {
		case *appsv1.Deployment:
			for _, e := range object.Spec.Template.Spec.Containers[0].Env {
				if e.Name == "GOMAXPROCS" {
					if e.Value != "3" {
						t.Error("GOMAXPROCS ≠ 3")
					}
					sawGoMaxProcs = true
				}
			}
		}
	}
	if !sawGoMaxProcs {
		t.Error("GOMAXPROCS does not exists when it should")
	}
}

//Tests todo
/*
Overall
- Check existence of values / non existence of values / overrides / defaults

- Namespace correctly set
- Service correctly wired through
- Secrets correctly wired through

TODO
- Storage Class Tests


*/
