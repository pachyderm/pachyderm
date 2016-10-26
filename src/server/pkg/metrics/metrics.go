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

type Reporter struct {
	clusterID  string
	kubeClient *kube.Client
	dbClient   *gorethink.Session
	pfsDbName  string
	ppsDbName  string
}

func NewReporter(clusterID string, kubeClient *kube.Client, address string, pfsDbName string, ppsDbName string) (*Reporter, error) {
	dbClient, err := db.DbConnect(address)
	if err != nil {
		return nil, fmt.Errorf("Error connected to DB when reporting metrics: %v\n", err)
	}
	return &Reporter{
		clusterID:  clusterID,
		kubeClient: kubeClient,
		dbClient:   dbClient,
		pfsDbName:  pfsDbName,
		ppsDbName:  ppsDbName,
	}, nil
}

// If we're not reporting metrics, incrementing should do nothing
var metricsEnabled = false

// Segment API allows for map[string]interface{} for a single user's traits
// But we only care about things that are countable for the moment
// map userID -> action name -> count
type countableActions map[string]interface{}
type countableUserActions map[string]countableActions

type incrementUserAction struct {
	action string
	user   string
}

var userActions = make(countableUserActions)
var incrementActionChannel = make(chan *incrementUserAction, 0)

//IncrementUserAction updates a counter per user per action for an API method by name
func IncrementUserAction(ctx context.Context, action string) {
	if !metricsEnabled {
		return
	}
	md, ok := metadata.FromContext(ctx)
	if ok && md["userid"] != nil && len(md["userid"]) > 0 {
		userID := md["userid"][0]
		incrementActionChannel <- &incrementUserAction{
			action: action,
			user:   userID,
		}
	}
}

//ReportSingleAction is used in the few places we need to report metrics from the client
func ReportSingleAction(action string) {
	oneShotClient := analytics.New(segmentAPIKey)
	cfg, err := config.Read()
	if err != nil {
		// Errors are non fatal when reporting metrics
		return
	}
	actions := make(countableActions)
	actions[action] = uint64(1)
	reportUserMetricsToSegment(oneShotClient, cfg.UserID, actions, "")
	oneShotClient.Close()
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

// ReportMetrics blocks and reports metrics every 15 seconds
func (r *Reporter) ReportMetrics() {
	metricsEnabled = true
	reportingTicker := time.NewTicker(time.Second * 15)
	for {
		select {
		case incrementAction := <-incrementActionChannel:
			if userActions[incrementAction.user] == nil {
				userActions[incrementAction.user] = make(countableActions)
			}
			val, ok := userActions[incrementAction.user][incrementAction.action]
			if !ok {
				val = uint64(0)
			}
			userActions[incrementAction.user][incrementAction.action] = val.(uint64) + uint64(1)
			break
		case <-reportingTicker.C:
			r.reportToSegment()
		}
	}
}

func (r *Reporter) reportToSegment() {
	if len(userActions) > 0 {
		batchOfUserActions := make(countableUserActions)
		// copy the existing stats into a new object so we can make the segment
		// request asynchronously
		for user, actions := range userActions {
			singleUserActions := make(countableActions)
			for name, count := range actions {
				singleUserActions[name] = count
			}
			batchOfUserActions[user] = singleUserActions
		}
		go r.reportUserMetrics(batchOfUserActions)
		userActions = make(countableUserActions)
	}
	go r.reportClusterMetrics()
}

func (r *Reporter) reportUserMetrics(batchOfUserActions countableUserActions) {
	if len(batchOfUserActions) > 0 {
		for userID, actions := range batchOfUserActions {
			reportUserMetricsToSegment(persistentClient, userID, actions, r.clusterID)
		}
	}
}

func (r *Reporter) reportClusterMetrics() {
	metrics := &Metrics{}
	r.dbMetrics(metrics)
	externalMetrics(r.kubeClient, metrics)
	metrics.ClusterID = r.clusterID
	metrics.PodID = uuid.NewWithoutDashes()
	metrics.Version = version.PrettyPrintVersion(version.Version)
	reportClusterMetricsToSegment(persistentClient, metrics)
}
