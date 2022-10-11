//go:build k8s

package server

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/minio/minio-go/v6"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	applyappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	applyv1 "k8s.io/client-go/applyconfigurations/core/v1"
	applymetav1 "k8s.io/client-go/applyconfigurations/meta/v1"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func get(t *testing.T, hc *http.Client, url string) error {
	t.Helper()
	ctx, c := context.WithTimeout(context.Background(), 5*time.Second)
	defer c()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return errors.Errorf("create http request for %v: %w", url, err)
	}

	res, err := hc.Do(req)
	if err != nil {
		return errors.Errorf("execute http request for %v: %w", url, err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Logf("warning: get %v: problem reading body: %v", url, err)
	}
	if got, want := res.StatusCode, http.StatusOK; got != want {
		t.Logf("get %v: body:\n%s", url, body)
		return errors.Errorf("get %v: response code\n  got: %v\n want: %v", url, got, want)
	}
	return nil
}

func proxyTest(t *testing.T, httpClient *http.Client, c *client.APIClient, secure bool) {
	t.Helper()
	httpPrefix := "http://"
	if secure {
		httpPrefix = "https://"
	}
	addr := fmt.Sprintf("%v:%d", c.GetAddress().Host, c.GetAddress().Port)

	// Test console.
	t.Run("TestConsole", func(t *testing.T) {
		require.NoErrorWithinTRetry(t, 60*time.Second, func() error {
			return get(t, httpClient, httpPrefix+addr+"/")
		}, "should be able to load console")
	})

	// Test OIDC.
	t.Run("TestOIDC", func(t *testing.T) {
		require.NoErrorWithinTRetry(t, 60*time.Second, func() error {
			return get(t, httpClient, httpPrefix+addr+"/dex/.well-known/openid-configuration")
		}, "should be able to load openid config")
	})

	testText := []byte("this is a test\n")
	for i := 0; i < 8*1024*1024; i++ {
		// Make sure the file doesn't fit in a single message.  Default message size is 4MB,
		// so 8MiB is comfortably large enough to ensure that both pachd and Envoy agree on
		// that it won't fit in a single message.
		testText = append(testText, (byte(i)%26)+'A')
	}

	// Test GRPC API.
	t.Run("TestGRPC", func(t *testing.T) {
		require.NoErrorWithinTRetry(t, 60*time.Second, func() error {
			if err := c.CreateRepo("test"); err != nil {
				return errors.Errorf("create repo: %w", err)
			}
			if err := c.PutFile(client.NewRepo("test").NewCommit("master", ""), "test.txt", bytes.NewReader(testText)); err != nil {
				return errors.Errorf("put file: %w", err)
			}
			return nil
		}, "should be able to upload a file over grpc")
	})

	// Test S3 API.
	t.Run("TestS3", func(t *testing.T) {
		require.NoErrorWithinTRetry(t, 60*time.Second, func() error {
			s3v4, err := minio.NewV4(addr, c.AuthToken(), c.AuthToken(), secure)
			if err != nil {
				return errors.Errorf("get s3v4 client: %w", err)
			}
			s3v2, err := minio.NewV2(addr, c.AuthToken(), c.AuthToken(), secure)
			if err != nil {
				return errors.Errorf("get s3v2 client: %w", err)
			}

			for name, client := range map[string]interface {
				GetObject(string, string, minio.GetObjectOptions) (*minio.Object, error)
				SetCustomTransport(http.RoundTripper)
			}{"v4": s3v4, "v2": s3v2} {
				client.SetCustomTransport(httpClient.Transport)

				obj, err := client.GetObject("master.test", "test.txt", minio.GetObjectOptions{})
				if err != nil {
					return errors.Errorf("s3%v: get object: %w", name, err)
				}
				got, err := io.ReadAll(obj)
				if err != nil {
					return errors.Errorf("s3%v: read all: %w", name, err)
				}
				want := testText
				if diff := cmp.Diff(got, want); diff != "" {
					return errors.Errorf("s3%v: content:\n%v", name, diff)
				}
			}
			return nil
		}, "should be able to retrieve files over S3")
	})
}

func deployFakeConsole(t *testing.T, ns string) {
	t.Helper()
	ctx := context.Background()
	c := testutil.GetKubeClient(t)
	lbl := map[string]string{
		"app":   "console",
		"suite": "pachyderm",
	}
	aopts := metav1.ApplyOptions{
		FieldManager: "proxy_test",
	}

	cm := applyv1.ConfigMap("fake-console-html", ns)
	cm.Data = map[string]string{
		"index.html": "<html><head><title>Hi</title></head><body><p>Hello from console!</p></body></html>",
	}
	_, err := c.CoreV1().ConfigMaps(ns).Apply(ctx, cm, aopts)
	require.NoError(t, err, "should create a configmap for the fake console HTML files")

	cm = applyv1.ConfigMap("fake-console-config", ns)
	cm.Data = map[string]string{
		"default.conf": `server {
			listen       4000;
			server_name  localhost;

			access_log  /dev/stdout  main;

			location / {
			    root   /usr/share/nginx/html;
			    index  index.html;
			}
		    }`,
	}
	_, err = c.CoreV1().ConfigMaps(ns).Apply(ctx, cm, aopts)
	require.NoError(t, err, "should create a configmap for the fake console config files")

	dep := applyappsv1.Deployment("console", ns)
	dep.ObjectMetaApplyConfiguration.Labels = lbl
	dep.Spec = applyappsv1.DeploymentSpec().WithReplicas(1).WithSelector(applymetav1.LabelSelector().WithMatchLabels(lbl))
	dep.Spec.Template = applyv1.PodTemplateSpec()
	dep.Spec.Template.ObjectMetaApplyConfiguration = applymetav1.ObjectMeta().WithName("console").WithNamespace(ns).WithLabels(lbl)
	dep.Spec.Template.Spec = applyv1.PodSpec().
		WithContainers(applyv1.Container().
			WithName("console").
			WithImage("nginx:latest").
			WithPorts(applyv1.ContainerPort().WithContainerPort(4000).WithName("console-http").WithProtocol(v1.ProtocolTCP)).
			WithVolumeMounts(
				applyv1.VolumeMount().WithName("html").WithMountPath("/usr/share/nginx/html"),
				applyv1.VolumeMount().WithName("config").WithMountPath("/etc/nginx/conf.d"),
			).
			WithReadinessProbe(applyv1.Probe().WithHTTPGet(
				applyv1.HTTPGetAction().WithPort(intstr.FromInt(4000)).WithPath("/")))).
		WithVolumes(
			applyv1.Volume().WithName("html").WithConfigMap(applyv1.ConfigMapVolumeSource().WithName("fake-console-html")),
			applyv1.Volume().WithName("config").WithConfigMap(applyv1.ConfigMapVolumeSource().WithName("fake-console-config")),
		)
	_, err = c.AppsV1().Deployments(ns).Apply(ctx, dep, aopts)
	require.NoError(t, err, "should create a 'console' deployment")

	t.Log("waiting for the console backend service to have endpoints")
	require.NoErrorWithinTRetry(t, 60*time.Second, func() error {
		ep, err := c.CoreV1().Endpoints(ns).Get(ctx, "console-proxy-backend", metav1.GetOptions{})
		if err != nil {
			return errors.Errorf("get console service endpoints: %v", err)
		}
		if ss := ep.Subsets; len(ss) > 0 {
			t.Logf("console endpoints: %v", ss)
			return nil
		}
		return errors.New("no endpoints yet")
	}, "console-proxy-backend should have endpoints")
}

func TestTrafficThroughProxy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()
	c, ns := minikubetestenv.AcquireCluster(t)
	deployFakeConsole(t, ns)
	testutil.ActivateAuthClient(t, c)
	require.NoError(t, testutil.ConfigureOIDCProvider(t, c, false))
	proxyTest(t, http.DefaultClient, c, false)
}

func TestTrafficThroughProxyTLS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		t.Fatalf("generate serial number: %v", err)
	}
	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"Pachyderm tests"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(5 * time.Minute),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	addr := minikubetestenv.GetPachAddress(t)
	if ip := net.ParseIP(addr.Host); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	} else {
		template.DNSNames = append(template.DNSNames, addr.Host)
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, pub, priv)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("marshal pkcs8 private key: %v", err)
	}

	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		t.Fatalf("parse cert: %v", err)
	}

	values := map[string]string{
		"proxy.tls.enabled":     "true",
		"proxy.tls.secretName":  "pachyderm-proxy-tls",
		"proxy.tls.secret.cert": string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})),
		"proxy.tls.secret.key":  string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})),
	}

	certpool := x509.NewCertPool()
	certpool.AddCert(cert)
	tlsConfig := &tls.Config{
		RootCAs: certpool,
	}

	c, ns := minikubetestenv.AcquireCluster(t, minikubetestenv.WithTLS, minikubetestenv.WithCertPool(certpool), minikubetestenv.WithValueOverrides(values))
	deployFakeConsole(t, ns)

	testutil.ActivateAuthClient(t, c)
	require.NoError(t, testutil.ConfigureOIDCProvider(t, c, false))

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   tlsConfig,
			ForceAttemptHTTP2: true,
		},
	}

	proxyTest(t, httpClient, c, true)
}
