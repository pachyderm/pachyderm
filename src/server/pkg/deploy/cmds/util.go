package cmds

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/pachyderm/pachyderm/src/server/pkg/serde"
	log "github.com/sirupsen/logrus"
	clientcmd "k8s.io/client-go/tools/clientcmd/api/v1"
)

func kubectl(stdin io.Reader, context *config.Context, args ...string) error {
	var environ []string = nil
	if context != nil {
		tmpfile, err := ioutil.TempFile("", "transient-kube-config-*.yaml")
		if err != nil {
			return errors.Wrapf(err, "failed to create transient kube config")
		}
		defer os.Remove(tmpfile.Name())

		config := clientcmd.Config{
			Kind:           "Config",
			APIVersion:     "v1",
			CurrentContext: "pachyderm-active-context",
			Contexts: []clientcmd.NamedContext{
				clientcmd.NamedContext{
					Name: "pachyderm-active-context",
					Context: clientcmd.Context{
						Cluster:   context.ClusterName,
						AuthInfo:  context.AuthInfo,
						Namespace: context.Namespace,
					},
				},
			},
		}

		var buf bytes.Buffer
		if err := encoder("yaml", &buf).Encode(config); err != nil {
			return errors.Wrapf(err, "failed to encode config")
		}

		tmpfile.Write(buf.Bytes())
		tmpfile.Close()

		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return errors.Wrapf(err, "failed to discover default kube config: could not get user home directory")
			}
			kubeconfig = path.Join(home, ".kube", "config")
			if _, err = os.Stat(kubeconfig); os.IsNotExist(err) {
				return errors.Wrapf(err, "failed to discover default kube config: %q does not exist", kubeconfig)
			}
		}
		kubeconfig = fmt.Sprintf("%s%c%s", kubeconfig, os.PathListSeparator, tmpfile.Name())

		// note that this will override `KUBECONFIG` (if it is already defined) in
		// the environment; see examples under
		// https://golang.org/pkg/os/exec/#Command
		environ = os.Environ()
		environ = append(environ, fmt.Sprintf("KUBECONFIG=%s", kubeconfig))

		if stdin == nil {
			stdin = os.Stdin
		}
	}

	ioObj := cmdutil.IO{
		Stdin:   stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Environ: environ,
	}

	args = append([]string{"kubectl"}, args...)
	return cmdutil.RunIO(ioObj, args...)
}

// Generates a random secure token, in hex
func generateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// Return the appropriate encoder for the given output format.
func encoder(output string, w io.Writer) serde.Encoder {
	if output == "" {
		output = "json"
	} else {
		output = strings.ToLower(output)
	}
	e, err := serde.GetEncoder(output, w,
		serde.WithIndent(2),
		serde.WithOrigName(true),
	)
	if err != nil {
		cmdutil.ErrorAndExit(err.Error())
	}
	return e
}

func kubectlCreate(dryRun bool, manifest []byte, opts *assets.AssetOpts) error {
	if dryRun {
		_, err := os.Stdout.Write(manifest)
		return err
	}
	// we set --validate=false due to https://github.com/kubernetes/kubernetes/issues/53309
	if err := kubectl(bytes.NewReader(manifest), nil, "apply", "-f", "-", "--validate=false", "--namespace", opts.Namespace); err != nil {
		return err
	}

	fmt.Println("\nPachyderm is launching. Check its status with \"kubectl get all\"")
	if opts.DashOnly || !opts.NoDash {
		fmt.Println("Once launched, access the dashboard by running \"pachctl port-forward\"")
	}
	fmt.Println("")

	return nil
}

// findEquivalentContext searches for a context in the existing config that
// references the same cluster as the context passed in. If no such context
// was found, default values are returned instead.
func findEquivalentContext(cfg *config.Config, to *config.Context) (string, *config.Context) {
	// first check the active context
	activeContextName, activeContext, _ := cfg.ActiveContext(false)
	if activeContextName != "" && to.EqualClusterReference(activeContext) {
		return activeContextName, activeContext
	}

	// failing that, search all contexts (sorted by name to be deterministic)
	contextNames := []string{}
	for contextName := range cfg.V2.Contexts {
		contextNames = append(contextNames, contextName)
	}
	sort.Strings(contextNames)
	for _, contextName := range contextNames {
		existingContext := cfg.V2.Contexts[contextName]

		if to.EqualClusterReference(existingContext) {
			return contextName, existingContext
		}
	}

	return "", nil
}

func contextCreate(namePrefix, namespace, serverCert string) error {
	kubeConfig, err := config.RawKubeConfig()
	if err != nil {
		return err
	}
	kubeContext := kubeConfig.Contexts[kubeConfig.CurrentContext]

	clusterName := ""
	authInfo := ""
	if kubeContext != nil {
		clusterName = kubeContext.Cluster
		authInfo = kubeContext.AuthInfo
	}

	cfg, err := config.Read(false)
	if err != nil {
		return err
	}

	newContext := &config.Context{
		Source:      config.ContextSource_IMPORTED,
		ClusterName: clusterName,
		AuthInfo:    authInfo,
		Namespace:   namespace,
		ServerCAs:   serverCert,
	}

	equivalentContextName, equivalentContext := findEquivalentContext(cfg, newContext)
	if equivalentContext != nil {
		cfg.V2.ActiveContext = equivalentContextName
		equivalentContext.Source = newContext.Source
		equivalentContext.ClusterDeploymentID = ""
		equivalentContext.ServerCAs = newContext.ServerCAs
		return cfg.Write()
	}

	// we couldn't find an existing context that is the same as the new one,
	// so we'll have to create it
	newContextName := namePrefix
	if _, ok := cfg.V2.Contexts[newContextName]; ok {
		newContextName = fmt.Sprintf("%s-%s", namePrefix, time.Now().Format("2006-01-02-15-04-05"))
	}

	cfg.V2.Contexts[newContextName] = newContext
	cfg.V2.ActiveContext = newContextName
	return cfg.Write()
}

// containsEmpty is a helper function used for validation (particularly for
// validating that creds arguments aren't empty
func containsEmpty(vals []string) bool {
	for _, val := range vals {
		if val == "" {
			return true
		}
	}
	return false
}

// getCompatibleVersion gets the compatible version of another piece of
// software, or falls back to a default
func getCompatibleVersion(displayName, subpath, defaultValue string) string {
	var relVersion string
	// This is the branch where to look.
	// When a new version needs to be pushed we can just update the
	// compatibility file in pachyderm repo branch. A (re)deploy will pick it
	// up. To make this work we have to point the URL to the branch (not tag)
	// in the repo.
	branch := version.BranchFromVersion(version.Version)
	if version.IsCustomRelease(version.Version) {
		relVersion = version.PrettyPrintVersionNoAdditional(version.Version)
	} else {
		relVersion = version.PrettyPrintVersion(version.Version)
	}

	url := fmt.Sprintf("https://raw.githubusercontent.com/pachyderm/pachyderm/compatibility%s/etc/%s/%s", branch, subpath, relVersion)
	resp, err := http.Get(url)
	if err != nil {
		log.Warningf("error looking up compatible version of %s, falling back to %s: %v", displayName, defaultValue, err)
		return defaultValue
	}

	// Error on non-200; for the requests we're making, 200 is the only OK
	// state
	if resp.StatusCode != 200 {
		log.Warningf("error looking up compatible version of %s, falling back to %s: unexpected return code %d", displayName, defaultValue, resp.StatusCode)
		return defaultValue
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warningf("error looking up compatible version of %s, falling back to %s: %v", displayName, defaultValue, err)
		return defaultValue
	}

	allVersions := strings.Split(strings.TrimSpace(string(body)), "\n")
	if len(allVersions) < 1 {
		log.Warningf("no compatible version of %s found, falling back to %s", displayName, defaultValue)
		return defaultValue
	}
	latestVersion := strings.TrimSpace(allVersions[len(allVersions)-1])
	return latestVersion
}
