package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/emptypb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cmd"
	"sigs.k8s.io/kind/pkg/log"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/version"
)

const KindTestClusterName = "pachyderm-test"

// settings:
// - namespace
// - pachyderm version (version #, tag, "latest", "nightly", or "local")
// - kubernetes version
// - whether to deploy pachyderm or not (e.g. if using minikubetestenv)
// - local storage vs. minio
// - generic helm args
// - auth or no auth
// - include mount-server image in build?
// -

// deployOp represents a single deployment operation. The variables in this
// struct, consequently, fall into two categories:
//  1. Parameters that are set by the user (e.g. pachVersion, kubeVersion)
//  2. Intermediate (or cached) results calculated by the program (e.g. k8sManifest)
//
// Note that 'err' is a special case of (2), as if it is set, it means that the
// overall operation failed, and other intermdiate results are unset or invalid.
type deployOp struct {
	//////////////////////////////
	//// User-Provided Fields ////
	//////////////////////////////

	// pachVersion specifies, approximately, the version of pachyderm that should
	// be deployed.  This can be a tag or docker image sha (which will be pulled
	// if it doesn't exist locally, and then pushed), or it can be any of the
	// special strings:
	// - "latest" (which uses the most recent minor release of Pachyderm),
	// - "nightly" (which uses the most recently nightly release of Pachyderm), or
	// - "local" (which uses the special tag ":local").
	pachVersion string
	// objectStorage specifies the type of object storage to use. This can be "minio" or "local"
	objectStorage string

	//////////////////////////////
	//// Intermediate Results ////
	//////////////////////////////

	// ctx is the context associated with the overall restart operation. It's
	// used in 'pachClient' and 'kubeClient' below, as well as other operations
	// that require a context.
	//
	// This is set lazily by the Ctx() accessor, so that it can be left undefined
	// when not useful (such as printing out a manifest). Therefore it should
	// always be accessed via the Ctx() accessor.
	ctx context.Context

	// cancel is the cancel function associated with 'ctx' above. It likewise
	// is set lazily by op.Ctx(), and should only be called via op.Cancel(), in
	// case it was never initialized.
	cancel func()

	// pachClient is a pre-initialized client for pachd, to be shared across
	// various subtasks. Like ctx, kubeClient and others, this is initialized
	// lazily by the PachClient() method, and should only be accessed through it
	// (even internally, by other methods, in case it hasn't been initialized
	// yet). This is so that it's not initialized for operations that don't
	// require it, such as printing out a manifest, allowing those operations to
	// be run even in the absence of a kube/pach cluster.)
	pachClient *client.APIClient

	// kubeClient is a pre-initialized client for Kubernetes, to be shared across
	// various subtasks. Like ctx, pachClient, and others, this is initialized
	// lazily by the KubeClient() method, and should only be accessed through it
	// (even internally, by other methods, in case it hasn't been initialized
	// yet). This is so that it's not initialized for operations that don't
	// require it, such as printing out a manifest, allowing those operations to
	// be run even in the absence of a kube/pach cluster.)
	kubeClient *kubernetes.Clientset

	// kindProvider is a pre-initialized kind client (or cluster provider,
	// actually, as kind is local), to be shared across various subtasks. Like
	// ctx, pachClient, and others, this is initialized lazily by the
	// KindProvider() method, and should only be accessed through it (even
	// internally, by other methods, in case it hasn't been initialized yet).
	// This is so that it's not initialized for operations that don't require
	// it, such as printing out a manifest, allowing those operations to be run
	// even in the absence of kind or Docker.)
	kindProvider *cluster.Provider

	// k8sManifest is the kubernetes manifest that will be deployed. Like ctx,
	// pachClient, and others, This is initialized lazily by K8sManifest() and
	// should only be accessed through it (even internally, by other methods, in
	// case it hasn't been initialized yet).
	k8sManifest string

	// err is set if the overall operation has failed, and indicates the cause.
	err error

	// oldPachdPodName is the name of the pachd pod that will be replaced
	// (including pod ID and so on). This is used to determined whether 'pachctl
	// version' is talking to the old pod or the new pod.
	oldPachdPodName string

	// oldPachVersion is the version of pachyderm that is being replaced. If
	// this is the same as 'pachVersion', then we don't need to push a new
	// image.
	oldPachVersion string
}

// Option is the type of optional values passed to NewDeployOp
type Option func(*deployOp) error

// WithPachVersion is an option for NewDeployOp that sets the version of the
// Pachyderm instance that will be deployed. The argument can be any released
// version of Pachyderm (e.g. "2.8.1"), or any of the special strings:
//   - "local" (indicating that a locally-build Pachyderm image should be
//     deployed; commonly used on Core)
//   - "latest" (indicating that the most recent minor release of Pachyderm should
//     be deployed) (TODO)
//   - nightly (indicating that the most recent nightly release of Pachyderm
//     should be deployed)
func WithPachVersion(v string) Option {
	return func(op *deployOp) error {
		op.pachVersion = v
		return nil
	}
}

func WithObjectStorage(v string) Option {
	// WithPachVersion is an option for NewDeployOp that sets the version of the
	// Pachyderm instance that will be deployed.
	return func(op *deployOp) error {
		op.pachVersion = v
		return nil
	}
}

func NewDeployOp(verbose bool, opts ...Option) (*deployOp, error) {
	result := &deployOp{
		pachVersion:   "local",
		objectStorage: "minio",
	}
	for _, op := range opts {
		if err := op(result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// ctx and cancel are captured by result.Ctx(), and will not be initialized
// if that is never called
func (op *deployOp) Ctx() context.Context {
	if op.ctx != nil {
		return op.ctx
	}
	op.ctx, op.cancel = context.WithCancel(context.Background())
	return op.ctx
}

func (op *deployOp) Cancel() {
	if op.cancel != nil {
		op.cancel()
	}
}

// KubeClient is an accessor for op.kubeClient that lazily initializes the
// field if it's not already set.  All access to op.kubeClient should be through
// this method, even internally, in case it hasn't been initialized yet.
func (op *deployOp) KubeClient() *kubernetes.Clientset {
	if op.kubeClient != nil {
		return op.kubeClient
	}

	// Load the Kubernetes configuration from the default location
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		op.err = errors.Wrap(err, "could not load kubernetes config")
	}

	op.kubeClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		op.err = errors.Wrap(err, "could not create kubernetes client")
	}
	return op.kubeClient
}

// PachClient is an accessor for op.pachClient that lazily initializes the
// field if it's not already set.  All access to op.kubeClient should be through
// this method, even internally, in case it hasn't been initialized yet.
func (op *deployOp) PachClient() *client.APIClient {
	if op.pachClient != nil {
		return op.pachClient
	}

	pachctlCfg := &pachctl.Config{}
	var err error
	op.pachClient, err = pachctlCfg.NewOnUserMachine(op.Ctx(), false)
	if err != nil {
		op.err = errors.Wrap(err, "could not create Pachyderm client")
	}
	return op.pachClient
}

// KindProvider is an accessor for op.kindProvider that lazily initializes the
// field if it's not already set.  All access to op.kindProvider should be through
// this method, even internally, in case it hasn't been initialized yet.
func (op *deployOp) KindProvider() *cluster.Provider {
	if op.kindProvider != nil {
		return op.kindProvider
	}
	var kindLogger log.Logger = log.NoopLogger{}
	if verbose {
		// TODO(msteffen) literally no idea what good UX is here. Ideally we
		// would log to stdout instead of stderr, but I can't access
		// kind/pgk/internal/cli to change the logging destination. Am I holding
		// this wrong?
		kindLogger = cmd.NewLogger()
	}
	op.kindProvider = cluster.NewProvider(
		cluster.ProviderWithLogger(kindLogger),
		cluster.ProviderWithDocker(), // TODO(msteffen): is this necessary or appropriate?
	)
	return op.kindProvider
}

// helmChartSource is a helper for K8sManifest that returns the location of the
// Pachyderm helm chart it should use.
func (op *deployOp) helmChartSource() string {
	if op.pachVersion == "latest" {
		// // TODO(msteffen): this codepath has no automated tests. Maybe set up
		// // mock GitHub? (has worked in the past, and I bet the releases endpoint
		// // is simple)
		// client := github.NewClient(nil)
		// // Not sure if this is the correct call -- want to filter out alphas,
		// // nightlies, etc.
		// release, _, err := client.Repositories.GetLatestRelease(context.Background(), "pachyderm", "pachyderm")
		// if err != nil {
		// 	panic(err)
		// }
		// return fmt.Sprintf("https://github.com/pachyderm/helmchart/releases/download/pachyderm-%s/pachyderm-%s.tgz", *release.TagName, *release.TagName)
		panic("not implemented")
	}
	if op.pachVersion == "nightly" {
		// look up from dockerhub or something, somehow
		panic("not implemented")
	}
	if op.pachVersion == "local" {
		return "./etc/helm/pachyderm"
	}
	panic("do not know how to get helm chart for pachVersion " + op.pachVersion)
}

// getHelmArgs is a helper for K8sManifest that returns the arguments it should
// pass to helm
func (op *deployOp) getHelmArgs() []string {
	args := []string{
		"--set", "pachd.image.tag=local",
		"--set", "pachd.enterpriseLicenseKey=" + os.Getenv("ENT_ACT_CODE"), // used in CI -- weird name, though
		"--set", "pachd.lokiDeploy=false",
		"--set", "pachd.lokiLogging=false",
		"--set", "pachd.clusterDeploymentID=dev",
		"--set", "proxy.service.type=NodePort",
		"--set", "pachd.rootToken=iamroot",
	}

	if op.objectStorage == "minio" {
		args = append(args,
			"--set", "deployTarget=custom",
			"--set", "pachd.storage.backend=MINIO",
			"--set", "pachd.storage.minio.bucket=pachyderm-test",
			"--set", "pachd.storage.minio.endpoint=minio.default.svc.cluster.local:9000",
			"--set", "pachd.storage.minio.id=minioadmin",
			"--set", "pachd.storage.minio.secret=minioadmin",
			"--set-string", "pachd.storage.minio.signature=",
			"--set-string", "pachd.storage.minio.secure=false",
		)
	} else if op.objectStorage == "local" {
		args = append(args,
			"--set", "deployTarget=LOCAL",
		)
	}

	return args
}

// K8sManifest is an accessor for op.k8sManifest that lazily initializes the
// field if it's not already set.  All access to op.k8sManifest should be through
// this method, even internally, in case it hasn't been initialized yet.
func (op *deployOp) K8sManifest() string {
	if op.err != nil {
		return ""
	}
	if op.k8sManifest != "" {
		return op.k8sManifest
	}

	helmArgs := []string{"template", "pach", op.helmChartSource()}
	helmArgs = append(helmArgs, op.getHelmArgs()...)
	cmd := exec.Command("helm", helmArgs...)
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	if err := cmd.Run(); err != nil {
		op.err = errors.Wrap(err, "failed to generate k8s manifest with helm:\n"+stderr.String())
		return ""
	}
	op.k8sManifest = stdout.String()
	return op.k8sManifest
}

// getImages returns a slice of all docker images that 'op' will reference in
// it's Pachyderm deploy manifest *except* pachd and worker. These images can
// downloaded (via 'docker pull') and pushed into the cluster (via 'kind load').
func (op *deployOp) getImages() ([]string, error) {
	if op.err != nil {
		return nil, op.err
	}

	// Parse the helm chart (and minio k8s manifest, if in use) and extract all
	// images in it. Unfortunately, while 'helm install' can write its output as
	// JSON, 'helm template' cannot, so we query the output with yq.
	decoder := yqlib.NewYamlDecoder(yqlib.NewDefaultYamlPreferences())
	yamlDocs := []io.Reader{
		strings.NewReader(op.getK8sManifest()),
	}
	if op.objectStorage == "minio" {
		minioManifest, err := os.Open("etc/testing/minio.yaml")
		if err != nil {
			op.err = errors.Wrap(err, "could not read minio manifest to load minio image")
			return nil, op.err
		}
		yamlDocs = append(yamlDocs, minioManifest)
	}
	images := []string{"jaegertracing/all-in-one:1.10.1"}
	for _, doc := range yamlDocs {
		reader := bufio.NewReader(doc)
		if err := decoder.Init(reader); err != nil {
			op.err = errors.Wrap(err, "could not initialize yqlib decoder")
			return nil, op.err
		}

		for {
			// decoder.Decode returns one yaml document, but the k8s manifest
			// contains one document per k8s artifact (i.e. many documents
			// total), so we'll need to decode all of them.
			//
			// Taken from https://github.com/mikefarah/yq/blob/aaef27147fadc1cc2d59c7f3c940f5281a1a8654/pkg/yqlib/all_at_once_evaluator_test.go#L35-L56
			candidateNode, err := decoder.Decode()
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				op.err = errors.Wrap(err, "could not decode k8s manifest")
				return nil, op.err
			}
			yqExpr := ".. | .image? | select(. | . != null)"
			l, err := yqlib.NewAllAtOnceEvaluator().EvaluateNodes(yqExpr, candidateNode)
			if err != nil {
				op.err = errors.Wrapf(err, "could not evaluate yqlib expression %q against helm-generated manifest", yqExpr)
				return nil, op.err
			}
			for el := l.Front(); el != nil; el = el.Next() {
				image := el.Value.(*yqlib.CandidateNode).Value
				if strings.Contains(image, "pachyderm/pachd") && op.pachVersion == "local" {
					continue // this will be built and pushed in a different fn
				}
				images = append(images, image)
			}
		}
	}

	// sort & dedupe 'images' (wish there was a better way to do this)
	slices.Sort(images)
	i, j := 1, 1
	for ; j < len(images); j++ {
		if images[j] == images[j-1] {
			continue
		}
		if i != j {
			images[i] = images[j]
		}
		i++
	}
	if i < len(images) {
		return images[:i], nil
	}
	return images, nil
}

func (op *deployOp) getOldPachVersion() (string, error) {
	// Call the 'Version' API to get the Pachyderm version
	resp, err := op.PachClient().GetVersion(op.Ctx(), &emptypb.Empty{})
	if err != nil {
		return "", err
	}

	return version.PrettyPrintVersion(resp), nil
}

func (op *deployOp) getPachdPodName() (string, error) {
	// Retrieve the list of pachd pods (matching the below label selector)
	labelSelector := labels.Set{
		"suite": "pachyderm",
		"app":   "pachd",
	}.AsSelector()
	podList, err := op.KubeClient().CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return "", err
	}

	// Check if there is a running pachd pod
	if len(podList.Items) == 0 {
		return "", fmt.Errorf("no running pachd pod found")
	}

	// Get the name of the first pachd pod
	pachdPodName := podList.Items[0].Name

	return pachdPodName, nil
}

func (op *deployOp) maybeBuildOrPullPachyderm() {
	// Not implemented
}

// maybeDeleteK8sCluster deletes the kind cluster if it exists.
// Taken from: https://github.com/kubernetes-sigs/kind/blob/8f6cb8f45de2c56e4d13de1b4678f1d538755208/pkg/cmd/kind/delete/cluster/deletecluster.go#L68-L77
func (op *deployOp) maybeDeleteK8sCluster() error {
	if err := op.KindProvider().Delete(KindTestClusterName, ""); err != nil {
		op.err = errors.Wrap(err, "failed to delete kind cluster")
		return op.err
	}
	return nil
}

// maybeCreateK8sCluster creates the kind cluster if it doesn't exist.
func (op *deployOp) maybeCreateK8sCluster() error {
	// TODO(msteffen) maybe switch to sigs.k8s.io/kind/pkg/cluster?

	// Check if the cluster already exists
	// The implementation of `kind get clusters` is here:
	// https://github.com/kubernetes-sigs/kind/blob/8f6cb8f45de2c56e4d13de1b4678f1d538755208/pkg/cmd/kind/get/clusters/clusters.go#L48-L63
	clusters, err := op.KindProvider().List()
	if err != nil {
		op.err = errors.Wrap(err, "could not list existing kind clusters")
		return op.err
	}
	for _, cluster := range clusters {
		if strings.Contains(cluster, KindTestClusterName) {
			return nil
		}
	}

	// Create the cluster
	// The implementation of `kind create cluster` is here:
	// https://github.com/kubernetes-sigs/kind/blob/8f6cb8f45de2c56e4d13de1b4678f1d538755208/pkg/cmd/kind/create/cluster/createcluster.go#L98-L123
	if err = op.KindProvider().Create(
		KindTestClusterName,
		cluster.CreateWithDisplayUsage(true),
		cluster.CreateWithDisplaySalutation(true),
	); err != nil {
		op.err = errors.Wrap(err, "failed to create cluster")
		return op.err
	}
	return nil
}

func (op *deployOp) maybeDeployPachyderm() {
	op.
}

func (op *deployOp) Deploy() {
	op.maybeDeleteK8sCluster()

	eg := &errgroup.Group{}
	eg.Go(op.maybeCreateK8sCluster)
	// eg.Go(op.maybeBuildOrPullPachyderm)
	// ###
	// # Wait for minikube to come up and for pachctl (and the pachd/worker images) to
	// # finish building (the prereqs for deploying)
	// ###
	// set +x
	// WHEEL='\-/|'; W=0
	// until minikube status; do
	//   echo -en "\e[G${WHEEL:$((W=(W+1)%4)):1} Waiting for Minikube to come up..."
	//   sleep 1
	// done
	// ts "minikube is available"

	// until pachctl version --client-only >/dev/null 2>&1; do
	//   echo -en "\e[G${WHEEL:$((W=(W+1)%4)):1} Waiting for pachctl to build..."
	//   hash -r
	//   sleep 1
	// done
	// hash -r
	// ts "pachctl is built"

	// maybe_push_images_to_kube
	// ts ">>> images pushed <<<"

	// echo ""
	// echo "###"
	// echo "# Deploying pachyderm version v$(pachctl version --client-only)"
	// echo "###"
	// set -x

	// # Deploy minio (what we use for local deployments now). This is necessary
	// # even when pachyderm isn't being deployed (and goes into the default
	// # namespace, so no $K8S_ARGS) as minikubetestenv, which we use for unit tests
	// # and obviates the need to deploy pachyderm directly, expects it.
	// kubectl apply -f "$(git root)/etc/testing/minio.yaml"

	// ###
	// # Deploy pachyderm into minikube if needed
	// ###
	// get_old_pach_version
	// maybe_undeploy_pachyderm
	// maybe_deploy_pachyderm
	// maybe_connect_to_pachyderm

	// ts "minikube is up"
	// notify-send -i /home/mjs/Pachyderm/logo_little.png "Minikube is up" -t 10000
}

// RN:
// ✅ fix restart_minikube images
// ✅ test with kind instead of minikube -- time it
//      - kind is faster; no need to support both
// ✅ test op.getImages()
//
// Symbols: ❌ ✅ ✀ ❓ 🚧
// ✀ ts
// ❓ init_go_env_vars
//     - unclear if this is necessary or happens in subcommands (e.g. set GOBIN every time?)
//     - this should become unnecessary with Bazel migration
// ✅ parse_args -- happens in main.go (cobra)
// ✅ set_helm_command - test that this is working
// ✅ maybe_delete_cluster
// ✅ maybe_create_kube_cluster
// ❌ get_old_pach_version (next)
// ❌ maybe_undeploy_pachyderm
// ❌ maybe_build_or_pull_pachyderm
// ❌ maybe_push_images_to_kube
// ❌ maybe_deploy_pachyderm
// ❌ maybe_connect_to_pachyderm
// ❌ __main__
