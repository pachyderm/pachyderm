package client

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"sync"

	"github.com/pachyderm/pachyderm/src/client/pkg/config"

	"github.com/facebookgo/pidfile"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // enables support for configs with auth
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

const (
	pachdLocalPort         = 30650
	pachdRemotePort        = 650
	samlAcsLocalPort       = 30654
	dashUILocalPort        = 30080
	dashWebSocketLocalPort = 30081
	pfsLocalPort           = 30652
	s3gatewayLocalPort     = 30600
)

// PortForwarder handles proxying local traffic to a kubernetes pod
type PortForwarder struct {
	core          corev1.CoreV1Interface
	client        rest.Interface
	config        *rest.Config
	namespace     string
	logger        *io.PipeWriter
	stopChansLock *sync.Mutex
	stopChans     []chan struct{}
	shutdown      bool
}

// NewPortForwarder creates a new port forwarder
func NewPortForwarder(namespace string) (*PortForwarder, error) {
	cfg, err := config.Read()
	if err != nil {
		return nil, fmt.Errorf("could not read config: %v", err)
	}
	_, context, err := cfg.ActiveContext()
	if err != nil {
		return nil, fmt.Errorf("could not get active context: %v", err)
	}

	if namespace == "" {
		namespace = context.Namespace
	}
	if namespace == "" {
		namespace = "default"
	}

	kubeConfig := config.KubeConfig(context)

	kubeClientConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return nil, err
	}

	core := client.CoreV1()

	return &PortForwarder{
		core:          core,
		client:        core.RESTClient(),
		config:        kubeClientConfig,
		namespace:     namespace,
		logger:        log.StandardLogger().Writer(),
		stopChansLock: &sync.Mutex{},
		stopChans:     []chan struct{}{},
		shutdown:      false,
	}, nil
}

// Run starts the port forwarder. Returns after initialization is begun,
// returning any initialization errors.
func (f *PortForwarder) Run(appName string, localPort, remotePort uint16) error {
	podNameSelector := map[string]string{
		"suite": "pachyderm",
		"app":   appName,
	}

	podList, err := f.core.Pods(f.namespace).List(metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(podNameSelector)),
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		},
	})
	if err != nil {
		return err
	}
	if len(podList.Items) == 0 {
		return fmt.Errorf("No pods found for app %s", appName)
	}

	// Choose a random pod
	podName := podList.Items[rand.Intn(len(podList.Items))].Name

	url := f.client.Post().
		Resource("pods").
		Namespace(f.namespace).
		Name(podName).
		SubResource("portforward").
		URL()

	transport, upgrader, err := spdy.RoundTripperFor(f.config)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", url)
	ports := []string{fmt.Sprintf("%d:%d", localPort, remotePort)}
	readyChan := make(chan struct{}, 1)
	stopChan := make(chan struct{}, 1)

	// Ensure that the port forwarder isn't already shutdown, and append the
	// shutdown channel so this forwarder can be closed
	f.stopChansLock.Lock()
	if f.shutdown {
		f.stopChansLock.Unlock()
		return fmt.Errorf("port forwarder is shutdown")
	}
	f.stopChans = append(f.stopChans, stopChan)
	f.stopChansLock.Unlock()

	fw, err := portforward.New(dialer, ports, stopChan, readyChan, ioutil.Discard, f.logger)
	if err != nil {
		return err
	}

	errChan := make(chan error, 1)
	go func() { errChan <- fw.ForwardPorts() }()

	select {
	case err = <-errChan:
		return fmt.Errorf("port forwarding failed: %v", err)
	case <-fw.Ready:
		return nil
	}
}

// RunForDaemon creates a port forwarder for the pachd daemon.
func (f *PortForwarder) RunForDaemon(localPort, remotePort uint16) error {
	if localPort == 0 {
		localPort = pachdLocalPort
	}
	if remotePort == 0 {
		remotePort = pachdRemotePort
	}
	return f.Run("pachd", localPort, remotePort)
}

// RunForSAMLACS creates a port forwarder for SAML ACS.
func (f *PortForwarder) RunForSAMLACS(localPort uint16) error {
	if localPort == 0 {
		localPort = samlAcsLocalPort
	}
	return f.Run("pachd", localPort, 654)
}

// RunForDashUI creates a port forwarder for the dash UI.
func (f *PortForwarder) RunForDashUI(localPort uint16) error {
	if localPort == 0 {
		localPort = dashUILocalPort
	}
	return f.Run("dash", localPort, 8080)
}

// RunForDashWebSocket creates a port forwarder for the dash websocket.
func (f *PortForwarder) RunForDashWebSocket(localPort uint16) error {
	if localPort == 0 {
		localPort = dashWebSocketLocalPort
	}
	return f.Run("dash", localPort, 8081)
}

// RunForPFS creates a port forwarder for PFS over HTTP.
func (f *PortForwarder) RunForPFS(localPort uint16) error {
	if localPort == 0 {
		localPort = pfsLocalPort
	}
	return f.Run("pachd", localPort, 30652)
}

// RunForS3Gateway creates a port forwarder for the s3gateway.
func (f *PortForwarder) RunForS3Gateway(localPort uint16) error {
	if localPort == 0 {
		localPort = s3gatewayLocalPort
	}
	return f.Run("pachd", localPort, 600)
}

// Lock uses pidfiles to ensure that only one port forwarder is running across
// one or more `pachctl` instances
func (f *PortForwarder) Lock() error {
	pidfile.SetPidfilePath(path.Join(os.Getenv("HOME"), ".pachyderm/port-forward.pid"))
	return pidfile.Write()
}

// Close shuts down port forwarding.
func (f *PortForwarder) Close() {
	defer f.logger.Close()

	f.stopChansLock.Lock()
	defer f.stopChansLock.Unlock()

	if f.shutdown {
		panic("port forwarder already shutdown")
	}

	f.shutdown = true

	for _, stopChan := range f.stopChans {
		close(stopChan)
	}
}
