package metrics

import (
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/version"
	db "github.com/pachyderm/pachyderm/src/server/pfs/db"

	"github.com/dancannon/gorethink"
	"go.pedge.io/lion/proto"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

func dbMetrics(dbClient *gorethink.Session, pfsDbName string, ppsDbName string, metrics *Metrics) {
	cursor, err := gorethink.Object(
		"Repos",
		gorethink.DB(pfsDbName).Table("Repos").Count(),
		"Commits",
		gorethink.DB(pfsDbName).Table("Commits").Count(),
		"ArchivedCommits",
		gorethink.DB(pfsDbName).Table("Commits").Filter(
			map[string]interface{}{
				"Archived": true,
			},
		).Count(),
		"CancelledCommits",
		gorethink.DB(pfsDbName).Table("Commits").Filter(
			map[string]interface{}{
				"Cancelled": true,
			},
		).Count(),
		"Files",
		gorethink.DB(pfsDbName).Table("Diffs").Group("Path").Ungroup().Count(),
		"Jobs",
		gorethink.DB(ppsDbName).Table("JobInfos").Count(),
		"Pipelines",
		gorethink.DB(ppsDbName).Table("PipelineInfos").Count(),
	).Run(dbClient)
	if err != nil {
		protolion.Errorf("Error Fetching Metrics:%+v", err)
	}
	cursor.One(&metrics)
}

// ReportMetrics blocks and reports metrics, if modified, to the
// given kubernetes client every 15 seconds.
func ReportMetrics(clusterID string, kubeClient *kube.Client, address string, pfsDbName string, ppsDbName string) {
	dbClient, err := db.DbConnect(address)
	if err != nil {
		protolion.Errorf("Error connected to DB when reporting metrics: %v\n", err)
		return
	}
	for {
		metrics := &Metrics{}
		dbMetrics(dbClient, pfsDbName, ppsDbName, metrics)
		externalMetrics(kubeClient, metrics)
		metrics.ID = clusterID
		metrics.PodID = uuid.NewWithoutDashes()
		metrics.Version = version.PrettyPrintVersion(version.Version)
		reportSegment(metrics)
		<-time.After(15 * time.Second)
	}
}
