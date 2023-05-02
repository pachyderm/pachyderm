package lokiutil

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

func TestQueryRange(t *testing.T) {
	ctx := context.Background()
	config, kc, namespace := minikubetestenv.AcquireKubernetesCluster(t)
	transport, upgrader, err := spdy.RoundTripperFor(config)
	require.NoError(t, err)
	// config.Host is not guaranteed to be a URL, but if it’s not then we are stuck.
	u, err := url.Parse(config.Host)
	require.NoError(t, err, "could not parse %q as URL", config.Host)
	u.Path = fmt.Sprintf("/api/v1/namespaces/%s/pods/%s-loki-0/portforward", namespace, namespace)
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, u)
	stopCh := make(chan struct{})
	readyCh := make(chan struct{})
	streams, _, _, _ := genericclioptions.NewTestIOStreams()
	svc, err := kc.CoreV1().Services(namespace).Get(ctx, namespace+"-loki", metav1.GetOptions{})
	require.NoError(t, err, "could not get Loki service")
	var lokiPort int32
	for _, port := range svc.Spec.Ports {
		lokiPort = port.Port
	}
	require.False(t, lokiPort == 0, "could not find Loki port")

	fw, err := portforward.New(dialer, []string{fmt.Sprintf("0:%d", lokiPort)}, stopCh, readyCh, streams.Out, streams.ErrOut)
	require.NoError(t, err, "could forward Loki port")
	errCh := make(chan error)
	go func(ctx context.Context, fw *portforward.PortForwarder, stopCh chan<- struct{}, readyCh <-chan struct{}, errCh chan<- error) {
		defer close(errCh)
		<-readyCh
		ports, err := fw.GetPorts()
		if err != nil {
			close(stopCh)
			errCh <- errors.Wrap(err, "could not get ports")
			return
		}
		var c = new(loki.Client)
		require.True(t, len(ports) == 1, "did not get exactly one Loki port")
		for _, p := range ports {
			c.Address = fmt.Sprintf("http://localhost:%d", p.Local)
		}

		if err := QueryRange(ctx, c, "{job=~\".+\"}", time.Now().Add(-60*24*time.Hour), time.Now(), false, func(t time.Time, line string) error {
			return nil
		}); err != nil {
			close(stopCh)
			errCh <- errors.Wrap(err, "could not query large time range")
			return
		}

		qr, err := c.QueryRange(ctx, "{job=~\".+\"}", 1, time.Now().Add(-60*24*time.Hour), time.Time{}, "forward", 0, 0, false)
		if err != nil {
			close(stopCh)
			errCh <- errors.Wrap(err, "could not get max query length")
			return
		}
		log.Println("QQQ QR", qr)
		close(stopCh)
	}(ctx, fw, stopCh, readyCh, errCh)
	err = fw.ForwardPorts()
	require.NoError(t, err, "could not forward ports")

	if err, ok := <-errCh; ok {
		require.NoError(t, err, "error querying")
	}
}
