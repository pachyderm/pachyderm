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
	"go.pedge.io/lion/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

var globalReporter *Reporter

type Reporter struct {
	segmentClient *analytics.Client
	clusterID     string
	kubeClient    *kube.Client
	dbClient      *gorethink.Session
	pfsDbName     string
	ppsDbName     string
}

func InitializeReporter(clusterID string, kubeClient *kube.Client, address string, pfsDbName string, ppsDbName string) error {

	dbClient, err := db.DbConnect(address)
	if err != nil {
		return fmt.Errorf("Error connected to DB when reporting metrics: %v\n", err)
	}
	globalReporter = &Reporter{
		segmentClient: newPersistentClient(),
		clusterID:     clusterID,
		kubeClient:    kubeClient,
		dbClient:      dbClient,
		pfsDbName:     pfsDbName,
		ppsDbName:     ppsDbName,
	}
	go globalReporter.reportClusterMetrics()
	return nil
}

//ReportUserAction pushes the action into a queue for reporting
func ReportUserAction(ctx context.Context, action string, value interface{}) {
	if globalReporter == nil {
		// Metrics are disabled
		return
	}
	md, ok := metadata.FromContext(ctx)
	// metadata API downcases all the key names
	if ok && md["userid"] != nil && len(md["userid"]) > 0 {
		userID := md["userid"][0]
		reportUserMetricsToSegment(
			globalReporter.segmentClient,
			userID,
			action,
			value,
			globalReporter.clusterID,
		)
	}
}

//ReportAndFlushUserAction is used in the few places we need to report metrics from the client
func ReportAndFlushUserAction(action string, value interface{}) {
	// Need a new one each time, since we need to flush it
	client := newSegmentClient()
	defer client.Close()
	cfg, err := config.Read()
	if err != nil {
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
		protolion.Errorf("Error Fetching Metrics:%+v", err)
	}
	cursor.One(&metrics)
}

func (r *Reporter) reportClusterMetrics() {
	for {
		time.Sleep(reportingIntervalSeconds * time.Second)
		metrics := &Metrics{}
		r.dbMetrics(metrics)
		externalMetrics(r.kubeClient, metrics)
		metrics.ClusterID = r.clusterID
		metrics.PodID = uuid.NewWithoutDashes()
		metrics.Version = version.PrettyPrintVersion(version.Version)
		reportClusterMetricsToSegment(r.segmentClient, metrics)
	}
}
