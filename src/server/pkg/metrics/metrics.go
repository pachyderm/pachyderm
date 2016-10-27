package metrics

import (
	"fmt"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/version"
	db "github.com/pachyderm/pachyderm/src/server/pfs/db"

	"github.com/dancannon/gorethink"
	"github.com/segmentio/analytics-go"
	"go.pedge.io/lion"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

var batchedSegmentClient *analytics.Client
var metricsEnabled = false
var clusterID string

//Reporter is used to submit user & cluster metrics to segment
type Reporter struct {
	kubeClient *kube.Client
	dbClient   *gorethink.Session
	pfsDbName  string
	ppsDbName  string
}

// InitializeReporter is used to setup metrics to be reported to segment
func InitializeReporter(thisClusterID string, kubeClient *kube.Client, address string, pfsDbName string, ppsDbName string) {

	dbClient, err := db.DbConnect(address)
	if err != nil {
		lion.Errorf("error connected to DB when reporting metrics: %v\n", err)
		return
	}
	metricsEnabled = true
	batchedSegmentClient = newPersistentClient()
	clusterID = thisClusterID
	reporter := &Reporter{
		kubeClient: kubeClient,
		dbClient:   dbClient,
		pfsDbName:  pfsDbName,
		ppsDbName:  ppsDbName,
	}
	go reporter.reportClusterMetrics()
}

//ReportUserAction pushes the action into a queue for reporting,
// and reports the start, finish, and error conditions
func ReportUserAction(ctx context.Context, action string) func(time.Time, error) {
	reportUserAction(ctx, fmt.Sprintf("%vStarted", action), nil)
	return func(start time.Time, err error) {
		if err == nil {
			reportUserAction(ctx, fmt.Sprintf("%vFinished", action), time.Since(start).Seconds())
		} else {
			reportUserAction(ctx, fmt.Sprintf("%vErrored", action), err.Error())
		}
	}
}

func reportUserAction(ctx context.Context, action string, value interface{}) {
	if !metricsEnabled {
		return
	}
	md, ok := metadata.FromContext(ctx)
	// metadata API downcases all the key names
	if ok {
		if md["userid"] != nil && len(md["userid"]) > 0 {
			userID := md["userid"][0]
			reportUserMetricsToSegment(
				batchedSegmentClient,
				userID,
				action,
				value,
				clusterID,
			)
		} else {
			lion.Errorln("error extracting userid from metadata. userid is empty\n")
		}
	} else {
		lion.Errorf("Error extracting userid metadata from context: %v\n", ctx)
	}
}

//ReportAndFlushUserAction immediately reports the metric
// It is used in the few places we need to report metrics from the client.
// It handles reporting the start, finish, and error conditions of the action
func ReportAndFlushUserAction(action string) func(time.Time, error) {
	reportAndFlushUserAction(fmt.Sprintf("%vStarted", action), nil)
	return func(start time.Time, err error) {
		if err == nil {
			reportAndFlushUserAction(fmt.Sprintf("%vFinished", action), time.Since(start).Seconds())
		} else {
			reportAndFlushUserAction(fmt.Sprintf("%vErrored", action), err.Error())
		}
	}
}

func reportAndFlushUserAction(action string, value interface{}) {
	client := newSegmentClient()
	defer client.Close()
	cfg, err := config.Read()
	if err != nil {
		lion.Errorf("Error reading userid from ~/.pachyderm/config: %v\n", err)
		// Errors are non fatal when reporting metrics
		return
	}
	reportUserMetricsToSegment(client, cfg.UserID, action, value, "")
}

func (r *Reporter) dbMetrics(metrics *Metrics) {
	cursor, err := gorethink.Object(
		"Repos",
		gorethink.DB(r.pfsDbName).Table("Repos").Count(),
		"Commits",
		gorethink.DB(r.pfsDbName).Table("Commits").Count(),
		"ArchivedCommits",
		gorethink.DB(r.pfsDbName).Table("Commits").Filter(
			map[string]interface{}{
				"Archived": true,
			},
		).Count(),
		"CancelledCommits",
		gorethink.DB(r.pfsDbName).Table("Commits").Filter(
			map[string]interface{}{
				"Cancelled": true,
			},
		).Count(),
		"Files",
		gorethink.DB(r.pfsDbName).Table("Diffs").Group("Path").Ungroup().Count(),
		"Jobs",
		gorethink.DB(r.ppsDbName).Table("JobInfos").Count(),
		"Pipelines",
		gorethink.DB(r.ppsDbName).Table("PipelineInfos").Count(),
	).Run(r.dbClient)
	if err != nil {
		lion.Errorf("Error Fetching Metrics:%+v", err)
	}
	cursor.One(&metrics)
}

func (r *Reporter) reportClusterMetrics() {
	for {
		time.Sleep(reportingInterval)
		metrics := &Metrics{}
		r.dbMetrics(metrics)
		externalMetrics(r.kubeClient, metrics)
		metrics.ClusterID = clusterID
		metrics.PodID = uuid.NewWithoutDashes()
		metrics.Version = version.PrettyPrintVersion(version.Version)
		reportClusterMetricsToSegment(batchedSegmentClient, metrics)
	}
}
