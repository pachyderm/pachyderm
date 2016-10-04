package metrics

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/version"

	"github.com/dancannon/gorethink"
	"go.pedge.io/lion/proto"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

var metrics = &Metrics{}

func latestMetricsFromDB(dbClient *gorethink.Session, metrics *Metrics) {
	term := gorethink.DB(dbClient.Database())
	cursor, err := term.Object(
		"Repos",
		term.Table("Repos").GetAll().Count(),
		"Files",
		term.Table("Files").GetAll().Count(),
		"Jobs",
		term.Table("Jobs").GetAll().Count(),
		"Pipelines",
		term.Table("Pipelines").GetAll().Count(),
	).Run(dbClient)
	cursor.One(&metrics)
}

// ReportMetrics blocks and reports metrics, if modified, to the
// given kubernetes client every 15 seconds.
func ReportMetrics(clusterID string, kubeClient *kube.Client, dbClient *gorethink.Session) {
	metrics.ID = clusterID
	metrics.PodID = uuid.NewWithoutDashes()
	metrics.Version = version.PrettyPrintVersion(version.Version)
	for {
		externalMetrics(kubeClient, metrics)
		latestMetricsFromDB(dbClient, metrics)
		protolion.Info(metrics)
		fmt.Printf("Writing metrics: %v\n", metrics)
		reportSegment(metrics)
		<-time.After(15 * time.Second)
	}
}
