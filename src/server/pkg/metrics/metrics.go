package metrics

import (
	"fmt"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/version"
	db "github.com/pachyderm/pachyderm/src/server/pfs/db"

	"github.com/dancannon/gorethink"
	"go.pedge.io/lion/proto"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

var metrics = &Metrics{}

func latestMetricsFromDB(dbClient *gorethink.Session, pfsDbName string, ppsDbName string, metrics *Metrics) {
	cursor, err := gorethink.Object(
		"Repos",
		gorethink.DB(pfsDbName).Table("Repos").Count(),
		"Commits",
		gorethink.DB(pfsDbName).Table("Commits").Count(), // TODO: want archived commits as well
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
	metrics.ID = clusterID
	metrics.PodID = uuid.NewWithoutDashes()
	metrics.Version = version.PrettyPrintVersion(version.Version)
	for {
		externalMetrics(kubeClient, metrics)
		latestMetricsFromDB(dbClient, pfsDbName, ppsDbName, metrics)
		protolion.Info(metrics)
		fmt.Printf("!!! Writing metrics: %v\n", metrics)
		reportSegment(metrics)
		<-time.After(15 * time.Second)
	}
}
