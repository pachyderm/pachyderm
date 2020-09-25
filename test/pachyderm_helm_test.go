package helmtest

import (
	"errors"
	"fmt"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
)

func TestDashImageTag(t *testing.T) {

	helmChartPath := "../pachyderm"

	options := &helm.Options{
		SetValues: map[string]string{"dash.image.tag": "0.5.5"},
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
		SetValues: map[string]string{"etcd.image.tag": "blah"},
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
		SetValues: map[string]string{"pachd.image.tag": "1234"},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "deployment", []string{"templates/pachd/deployment.yaml"})

	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(t, output, &deployment)

	expectedContainerImage := "pachyderm/pachd:1234"
	deploymentContainers := deployment.Spec.Template.Spec.Containers
	if deploymentContainers[0].Image != expectedContainerImage {
		t.Fatalf("Rendered container image (%s) is not expected (%s)", deploymentContainers[0].Image, expectedContainerImage)
	}
}

func TestPachdImageTagDeploymentEnv(t *testing.T) {
	helmChartPath := "../pachyderm"

	expectedTag := "1234"
	expectedContainerImage := "pachyderm/pachd:" + expectedTag
	options := &helm.Options{
		SetValues: map[string]string{"pachd.image.tag": expectedTag},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, "deployment", []string{"templates/pachd/deployment.yaml"})

	var deployment appsv1.Deployment
	fmt.Printf("%+v\n", deployment)
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

	if version != expectedTag {
		t.Fatalf("Rendered EnvVar (%s) value (%s) is not expected (%s)", workerImageEnvVar, version, expectedTag)
	}
	if workerImage != expectedContainerImage {
		t.Fatalf("Rendered EnvVar (%s) value (%s) is not expected (%s)", workerImageEnvVar, workerImage, expectedContainerImage)
	}

	if workerSidecarImage != expectedContainerImage {
		t.Fatalf("Rendered EnvVar (%s) value (%s) is not expected (%s)", workerSidecarImageEnvVar, workerImage, expectedContainerImage)
	}

}

func TestSetNamespaceWorkerRoleBinding(t *testing.T) {

	helmChartPath := "../pachyderm"

	expectedNamespace := "kspace"

	options := &helm.Options{
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
			"pachd.storage.backend":                      "MICROSOFT",
			"pachd.storage.microsoft.microsoftContainer": microsoftContainer},
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
			"pachd.storage.backend":             "GOOGLE",
			"pachd.storage.google.googleBucket": googleBucket},
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
			"pachd.storage.backend":           "MINIO",
			"pachd.storage.minio.minioBucket": minioBucket},
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
