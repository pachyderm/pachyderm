package metrics

import (
	"fmt"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/version"

	"github.com/segmentio/analytics-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	kube "k8s.io/client-go/kubernetes"
)

//Reporter is used to submit user & cluster metrics to segment
type Reporter struct {
	segmentClient *analytics.Client
	clusterID     string
	kubeClient    *kube.Clientset
}

// NewReporter creates a new reporter and kicks off the loop to report cluster
// metrics
func NewReporter(clusterID string, kubeClient *kube.Clientset) *Reporter {
	reporter := &Reporter{
		segmentClient: newPersistentClient(),
		clusterID:     clusterID,
		kubeClient:    kubeClient,
	}
	go reporter.reportClusterMetrics()
	return reporter
}

//ReportUserAction pushes the action into a queue for reporting,
// and reports the start, finish, and error conditions
func ReportUserAction(ctx context.Context, r *Reporter, action string) func(time.Time, error) {
	if r == nil {
		// This happens when stubbing out metrics for testing, e.g. src/server/pfs/server/server_test.go
		return func(time.Time, error) {}
	}
	// If we report nil, segment sees it, but mixpanel omits the field
	r.reportUserAction(ctx, fmt.Sprintf("%vStarted", action), 1)
	return func(start time.Time, err error) {
		if err == nil {
			r.reportUserAction(ctx, fmt.Sprintf("%vFinished", action), time.Since(start).Seconds())
		} else {
			r.reportUserAction(ctx, fmt.Sprintf("%vErrored", action), err.Error())
		}
	}
}

func getKeyFromMD(md metadata.MD, key string) (string, error) {
	if md[key] != nil && len(md[key]) > 0 {
		return md[key][0], nil
	}
	return "", fmt.Errorf("error extracting userid from metadata. userid is empty")
}

func (r *Reporter) reportUserAction(ctx context.Context, action string, value interface{}) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		// metadata API downcases all the key names
		userID, err := getKeyFromMD(md, "userid")
		if err != nil {
			// The FUSE client will never have a userID, so normal usage will produce a lot of these errors
			return
		}
		prefix, err := getKeyFromMD(md, "prefix")
		if err != nil {
			log.Errorln(err)
			return
		}
		reportUserMetricsToSegment(
			r.segmentClient,
			userID,
			prefix,
			action,
			value,
			r.clusterID,
		)
	}
}

// Helper method called by (Start|Finish)ReportAndFlushUserAction. Like those
// functions, it is used by the pachctl binary and runs on users' machines
// TODO(msteffen): Wrap config parsing in a library
func reportAndFlushUserAction(action string, value interface{}) func() {
	metricsDone := make(chan struct{})
	go func() {
		client := newSegmentClient()
		defer client.Close()
		cfg, err := config.Read()
		if err != nil || cfg == nil || cfg.UserID == "" {
			log.Errorf("Error reading userid from ~/.pachyderm/config: %v", err)
			// metrics errors are non fatal
			return
		}
		reportUserMetricsToSegment(client, cfg.UserID, "user", action, value, "")
		close(metricsDone)
	}()
	return func() {
		select {
		case <-metricsDone:
			return
		case <-time.After(time.Second * 5):
			return
		}
	}
}

// StartReportAndFlushUserAction immediately reports the metric but does
// not block execution. It returns a wait function which waits or times
// out after 5s.
// It is used by the pachctl binary and runs on users' machines
func StartReportAndFlushUserAction(action string, value interface{}) func() {
	return reportAndFlushUserAction(fmt.Sprintf("%vStarted", action), value)
}

// FinishReportAndFlushUserAction immediately reports the metric but does
// not block execution. It returns a wait function which waits or times
// out after 5s.
// It is used by the pachctl binary and runs on users' machines
func FinishReportAndFlushUserAction(action string, err error, start time.Time) func() {
	var wait func()
	if err != nil {
		wait = reportAndFlushUserAction(fmt.Sprintf("%vErrored", action), err)
	} else {
		wait = reportAndFlushUserAction(fmt.Sprintf("%vFinished", action), time.Since(start).Seconds())
	}
	return wait
}

func (r *Reporter) reportClusterMetrics() {
	for {
		time.Sleep(reportingInterval)
		metrics := &Metrics{}
		externalMetrics(r.kubeClient, metrics)
		metrics.ClusterID = r.clusterID
		metrics.PodID = uuid.NewWithoutDashes()
		metrics.Version = version.PrettyPrintVersion(version.Version)
		reportClusterMetricsToSegment(r.segmentClient, metrics)
	}
}
