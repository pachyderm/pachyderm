package logs

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/logs"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/protoutil"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth"
)

// LogService implements the core logs functionality.
type LogService struct {
	GetLokiClient func() (*loki.Client, error)
	AuthServer    authserver.APIServer
}

var (
	// ErrUnimplemented is returned whenever requested functionality is planned but unimplemented.
	ErrUnimplemented = errors.New("unimplemented")
	// ErrPublish is returned whenever publishing fails (say, due to a closed client).
	ErrPublish = errors.New("error publishing")
	// ErrBadRequest indicates that the client made an erroneous request.
	ErrBadRequest = errors.New("bad request")
)

func userLogQueryValidateErr(requestType, missingArg string) error {
	return errors.Wrapf(ErrBadRequest, "%s UserLogQuery request requires a %s argument", requestType, missingArg)
}

func adminLogQueryValidateErr(requestType, missingArg string) error {
	return errors.Wrapf(ErrBadRequest, "%s AdminLogQuery request requires a %s argument", requestType, missingArg)
}

// GetLogs gets logs according its request and publishes them.  The pattern is
// similar to that used when handling an HTTP request.
func (ls LogService) GetLogs(ctx context.Context, request *logs.GetLogsRequest, publisher ResponsePublisher) error {
	if request == nil {
		request = &logs.GetLogsRequest{}
	}

	filter := request.Filter
	if filter == nil {
		filter = new(logs.LogFilter)
		request.Filter = filter
	}
	if filter.Limit > math.MaxInt {
		return errors.Wrapf(ErrBadRequest, "limit %d > maxint", filter.Limit)
	}

	c, err := ls.GetLokiClient()
	if err != nil {
		return errors.Wrap(err, "loki client error")
	}
	var (
		adapter = adapter{responsePublisher: publisher}
		logQL   string
	)
	logQL, adapter.pass, err = ls.compileRequest(ctx, request)
	if err != nil {
		return errors.Wrap(err, "cannot convert request to LogQL")
	}
	var from, to time.Time
	if t := filter.GetTimeRange().GetFrom(); t != nil {
		from = t.AsTime()
	}
	if t := filter.GetTimeRange().GetUntil(); t != nil {
		to = t.AsTime()
	}
	prev, next, prevOffset, nextOffset, err := c.BatchedQueryRange(ctx, logQL, from, to, uint(filter.GetTimeRange().GetOffset()), int(filter.Limit), adapter.publish)
	if err != nil {
		return errors.Wrap(err, "BatchedQueryRange")
	}
	if request.WantPagingHint {
		older, newer := protoutil.Clone(request), protoutil.Clone(request)
		if !prev.IsZero() {
			if older.Filter == nil {
				older.Filter = &logs.LogFilter{}
			}
			older.Filter.TimeRange = &logs.TimeRangeLogFilter{
				Until:  timestamppb.New(prev),
				Offset: uint64(prevOffset),
			}
		} else {
			older = nil
		}
		if !next.IsZero() {
			if newer.Filter == nil {
				newer.Filter = &logs.LogFilter{}
			}
			newer.Filter.TimeRange = &logs.TimeRangeLogFilter{
				From:   timestamppb.New(next),
				Offset: uint64(nextOffset),
			}
		} else {
			newer = nil
		}
		hint := &logs.GetLogsResponse{
			ResponseType: &logs.GetLogsResponse_PagingHint{
				PagingHint: &logs.PagingHint{
					Older: older,
					Newer: newer,
				},
			},
		}
		if err := publisher.Publish(ctx, hint); err != nil {
			return errors.Wrap(err, "send paging hint")
		}
	}
	return nil
}

func (ls LogService) compileRequest(ctx context.Context, request *logs.GetLogsRequest) (string, passFunc, error) {
	if request == nil {
		return "", nil, errors.New("nil request")
	}
	query := request.Query
	if query == nil {
		return "", nil, nil
	}
	switch query := query.QueryType.(type) {
	case *logs.LogQuery_User:
		return ls.compileUserLogQueryReq(ctx, query.User, true, request.GetFilter().GetUserLogsOnly())
	case *logs.LogQuery_Admin:
		if ls.AuthServer != nil {
			if err := ls.AuthServer.CheckClusterIsAuthorized(ctx, auth.Permission_CLUSTER_GET_LOKI_LOGS); err != nil && !auth.IsErrNotActivated(err) {
				return "", nil, errors.EnsureStack(err)
			}
		}
		switch query := query.Admin.GetAdminType().(type) {
		case *logs.AdminLogQuery_Logql:
			if query.Logql == "" {
				return "", nil, adminLogQueryValidateErr("Logql", "Logql")
			}
			return query.Logql, func(map[string]string, *logs.LogMessage) bool { return true }, nil
		case *logs.AdminLogQuery_Pod:
			if query.Pod == "" {
				return "", nil, adminLogQueryValidateErr("Pod", "Pod")
			}
			return fmt.Sprintf(`{suite="pachyderm",pod=%q}`, query.Pod), func(map[string]string, *logs.LogMessage) bool { return true }, nil
		case *logs.AdminLogQuery_PodContainer:
			if query.PodContainer.Pod == "" {
				return "", nil, adminLogQueryValidateErr("PodContainer", "Pod")
			}
			if query.PodContainer.Container == "" {
				return "", nil, adminLogQueryValidateErr("PodContainer", "Container")
			}
			return fmt.Sprintf(`{suite="pachyderm",pod=%q,container=%q}`, query.PodContainer.Pod, query.PodContainer.Container),
				func(map[string]string, *logs.LogMessage) bool { return true }, nil
		case *logs.AdminLogQuery_App:
			if query.App == "" {
				return "", nil, adminLogQueryValidateErr("App", "App")
			}
			return fmt.Sprintf(`{suite="pachyderm",app=%q}`, query.App), func(map[string]string, *logs.LogMessage) bool { return true }, nil
		case *logs.AdminLogQuery_Master:
			pipeline := query.Master.Pipeline
			project := query.Master.Project
			if pipeline == "" {
				return "", nil, adminLogQueryValidateErr("Master", "Pipeline")
			}
			if project == "" {
				return "", nil, adminLogQueryValidateErr("Master", "Project")
			}
			return fmt.Sprintf(`{app="pipeline",suite="pachyderm",container="user",pipelineProject=%q,pipelineName=%q}`, project, pipeline), func(labels map[string]string, msg *logs.LogMessage) bool {
				if msg.GetPpsLogMessage().Master {
					return true
				}
				ff := msg.GetObject().GetFields()
				if ff != nil {
					v, ok := ff["master"]
					if ok {
						if v.GetBoolValue() {
							return true
						}
					}
				}
				return false
			}, nil
		case *logs.AdminLogQuery_Storage:
			pipeline := query.Storage.Pipeline
			project := query.Storage.Project
			if pipeline == "" {
				return "", nil, adminLogQueryValidateErr("Storage", "Pipeline")
			}
			if project == "" {
				return "", nil, adminLogQueryValidateErr("Storage", "Project")
			}
			return fmt.Sprintf(`{app="pipeline",suite="pachyderm",container="storage",pipelineProject=%q,pipelineName=%q}`, project, pipeline),
				func(map[string]string, *logs.LogMessage) bool {
					return true
				}, nil
		case *logs.AdminLogQuery_User:
			return ls.compileUserLogQueryReq(ctx, query.User, false, request.GetFilter().GetUserLogsOnly())
		default:
			return "", nil, nil
		}
	default:
		return "", nil, errors.Wrapf(ErrUnimplemented, "%T", query)
	}
}

func (ls LogService) compileUserLogQueryReq(ctx context.Context, query *logs.UserLogQuery, checkAuth bool, userOnly bool) (string, func(map[string]string, *logs.LogMessage) bool, error) {
	switch query := query.GetUserType().(type) {
	case *logs.UserLogQuery_Pipeline:
		pipeline := query.Pipeline.Pipeline
		project := query.Pipeline.Project
		if checkAuth {
			if pass := ls.authPipelineLogs(ctx, pipeline, project, nil); !pass {
				return "", nil, nil
			}
		}
		return ls.compilePipelineLogsReq(project, pipeline, userOnly)
	case *logs.UserLogQuery_JobDatum:
		authCache := make(map[string]bool)
		return ls.compileJobDatumsLogsReq(ctx, query.JobDatum.Job, query.JobDatum.Datum, checkAuth, authCache, userOnly)
	case *logs.UserLogQuery_Datum:
		authCache := make(map[string]bool)
		return ls.compileDatumsLogsReq(ctx, query.Datum, checkAuth, authCache, userOnly)
	case *logs.UserLogQuery_Project:
		authCache := make(map[string]bool)
		return ls.compileProjectLogsReq(ctx, query.Project, checkAuth, authCache, userOnly)
	case *logs.UserLogQuery_Job:
		authCache := make(map[string]bool)
		return ls.compileJobLogsReq(ctx, query.Job, checkAuth, authCache, userOnly)
	case *logs.UserLogQuery_PipelineJob:
		project := query.PipelineJob.Pipeline.Project
		pipeline := query.PipelineJob.Pipeline.Pipeline
		if checkAuth {
			if pass := ls.authPipelineLogs(ctx, pipeline, project, nil); !pass {
				return "", nil, nil
			}
		}
		job := query.PipelineJob.Job
		return ls.compilePipelineJobLogsReq(project, pipeline, job, userOnly)
	case *logs.UserLogQuery_PipelineDatum:
		project := query.PipelineDatum.Pipeline.Project
		pipeline := query.PipelineDatum.Pipeline.Pipeline
		if checkAuth {
			if pass := ls.authPipelineLogs(ctx, pipeline, project, nil); !pass {
				return "", nil, nil
			}
		}
		datum := query.PipelineDatum.Datum
		return ls.compilePipelineDatumLogsReq(project, pipeline, datum, userOnly)
	default:
		return "", nil, errors.Wrapf(ErrUnimplemented, "%T", query)
	}
}

func matchesUserOnlyRequest(userOnly bool, msg *logs.LogMessage) bool {
	if !userOnly {
		return true
	}
	if msg.GetPpsLogMessage().GetUser() {
		return true
	}
	if ff := msg.GetObject().GetFields(); ff != nil {
		v, ok := ff["user"]
		if ok {
			if v.GetBoolValue() {
				return true
			}
		}
	}
	return false
}

func filterUserLogs(userOnly bool, next func(map[string]string, *logs.LogMessage) bool) func(map[string]string, *logs.LogMessage) bool {
	return func(labels map[string]string, msg *logs.LogMessage) bool {
		return next(labels, msg) && matchesUserOnlyRequest(userOnly, msg)
	}
}

func (ls LogService) compilePipelineLogsReq(project, pipeline string, userOnly bool) (string, func(map[string]string, *logs.LogMessage) bool, error) {
	if pipeline == "" {
		return "", nil, userLogQueryValidateErr("Pipeline", "Pipeline")
	}
	if project == "" {
		return "", nil, userLogQueryValidateErr("Pipeline", "Project")
	}
	filter := func(labels map[string]string, msg *logs.LogMessage) bool {
		return matchesUserOnlyRequest(userOnly, msg)
	}
	return fmt.Sprintf(`{app="pipeline",suite="pachyderm",container="user",pipelineProject=%q,pipelineName=%q}`, project, pipeline), filter, nil
}

func (ls LogService) compileJobDatumsLogsReq(ctx context.Context, job, datum string, checkAuth bool, authCache map[string]bool, userOnly bool) (string, func(map[string]string, *logs.LogMessage) bool, error) {
	if job == "" {
		return "", nil, userLogQueryValidateErr("JobDatum", "Job")
	}
	if datum == "" {
		return "", nil, userLogQueryValidateErr("JobDatum", "Datum")
	}
	return fmt.Sprintf(`{suite="pachyderm",app="pipeline"} |= %q`, datum), filterUserLogs(userOnly, func(labels map[string]string, msg *logs.LogMessage) bool {
		logDatumId := msg.GetPpsLogMessage().GetDatumId()
		logJobId := msg.GetPpsLogMessage().GetJobId()
		if logJobId != "" && logDatumId != "" {
			if logDatumId != datum && logJobId != job {
				return false
			}
		} else {
			ff := msg.GetObject().GetFields()
			if ff != nil {
				d, okDatum := ff["datumId"]
				j, okJob := ff["jobId"]
				if !okJob || !okDatum {
					return false
				}
				if d.GetStringValue() != datum || j.GetStringValue() != job {
					return false
				}
			}
		}
		if checkAuth {
			return ls.authLogMessage(ctx, labels, msg, authCache)
		}
		return true
	}), nil
}

func (ls LogService) compileDatumsLogsReq(ctx context.Context, datum string, checkAuth bool, authCache map[string]bool, userOnly bool) (string, func(map[string]string, *logs.LogMessage) bool, error) {
	if datum == "" {
		return "", nil, userLogQueryValidateErr("Datum", "Datum")
	}
	return fmt.Sprintf(`{suite="pachyderm",app="pipeline"} |= %q`, datum), filterUserLogs(userOnly, func(labels map[string]string, msg *logs.LogMessage) bool {
		if msg.GetPpsLogMessage().GetDatumId() != "" && msg.GetPpsLogMessage().GetDatumId() != datum {
			return false
		}
		ff := msg.GetObject().GetFields()
		if ff != nil {
			v, ok := ff["datumId"]
			if !ok {
				return false
			}
			if v.GetStringValue() != datum {
				return false
			}
		}
		if checkAuth {
			return ls.authLogMessage(ctx, labels, msg, authCache)
		}
		return true
	}), nil
}

func (ls LogService) compileProjectLogsReq(ctx context.Context, project string, checkAuth bool, authCache map[string]bool, userOnly bool) (string, func(map[string]string, *logs.LogMessage) bool, error) {
	if project == "" {
		return "", nil, userLogQueryValidateErr("Project", "Project")
	}
	return fmt.Sprintf(`{suite="pachyderm",app="pipeline",pipelineProject=%q}`, project), filterUserLogs(userOnly, func(labels map[string]string, msg *logs.LogMessage) bool {
		if checkAuth {
			return ls.authLogMessage(ctx, labels, msg, authCache)
		}
		return true
	}), nil
}

func (ls LogService) compileJobLogsReq(ctx context.Context, job string, checkAuth bool, authCache map[string]bool, userOnly bool) (string, func(map[string]string, *logs.LogMessage) bool, error) {
	if job == "" {
		return "", nil, userLogQueryValidateErr("Job", "Job")
	}
	return `{suite="pachyderm",app="pipeline"}`, filterUserLogs(userOnly, func(labels map[string]string, msg *logs.LogMessage) bool {
		if msg.GetPpsLogMessage().GetJobId() != "" && msg.GetPpsLogMessage().GetJobId() != job {
			return false
		}
		ff := msg.GetObject().GetFields()
		if ff != nil {
			v, ok := ff["jobId"]
			if !ok {
				return false
			}
			if v.GetStringValue() != job {
				return false
			}
		}
		if checkAuth {
			return ls.authLogMessage(ctx, labels, msg, authCache)
		}
		return true
	}), nil
}

func (ls LogService) compilePipelineJobLogsReq(project, pipeline, job string, userOnly bool) (string, func(map[string]string, *logs.LogMessage) bool, error) {
	if project == "" {
		return "", nil, userLogQueryValidateErr("PipelineJob", "Project")
	}
	if pipeline == "" {
		return "", nil, userLogQueryValidateErr("PipelineJob", "Pipeline")
	}
	if job == "" {
		return "", nil, userLogQueryValidateErr("PipelineJob", "Job")
	}
	return fmt.Sprintf(`{suite="pachyderm",pipelineProject=%q,pipelineName=%q}`, project, pipeline), filterUserLogs(userOnly, func(labels map[string]string, msg *logs.LogMessage) bool {
		if msg.GetPpsLogMessage().GetJobId() == job {
			return true
		}
		ff := msg.GetObject().GetFields()
		if ff != nil {
			v, ok := ff["jobId"]
			if ok {
				if v.GetStringValue() == job {
					return true
				}
			}
		}
		return false
	}), nil
}

func (ls LogService) compilePipelineDatumLogsReq(project, pipeline, datum string, userOnly bool) (string, func(map[string]string, *logs.LogMessage) bool, error) {
	if project == "" {
		return "", nil, userLogQueryValidateErr("PipelineDatum", "Project")
	}
	if pipeline == "" {
		return "", nil, userLogQueryValidateErr("PipelineDatum", "Pipeline")
	}
	if datum == "" {
		return "", nil, userLogQueryValidateErr("PipelineDatum", "Datum")
	}
	return fmt.Sprintf(`{suite="pachyderm",pipelineProject=%q,pipelineName=%q}`, project, pipeline), filterUserLogs(userOnly, func(labels map[string]string, msg *logs.LogMessage) bool {
		if msg.GetPpsLogMessage().GetDatumId() == datum {
			return true
		}
		ff := msg.GetObject().GetFields()
		if ff != nil {
			v, ok := ff["datumId"]
			if ok {
				if v.GetStringValue() == datum {
					return true
				}
			}
		}
		return false
	}), nil
}

func (ls LogService) authLogMessage(ctx context.Context, labels map[string]string, msg *logs.LogMessage, cache map[string]bool) bool {
	var (
		authed, labelsDenied, ppsLogMessageDenied, objectDenied bool
	)
	if labels != nil && labels["pipelineProject"] != "" && labels["pipelineName"] != "" {
		authed = true
		labelsDenied = !ls.authPipelineLogs(ctx, labels["pipelineName"], labels["pipelineProject"], cache)
	}
	if ppsLogMessage := msg.GetPpsLogMessage(); ppsLogMessage != nil && ppsLogMessage.ProjectName != "" && ppsLogMessage.PipelineName != "" {
		authed = true
		ppsLogMessageDenied = !ls.authPPSLogMessage(ctx, ppsLogMessage, cache)
	}
	ff := msg.GetObject().GetFields()
	if ff != nil {
		pipeline, ok := ff["pipelineName"]
		if ok {
			project, ok := ff["projectName"]
			if ok {
				authed = true
				objectDenied = !ls.authPipelineLogs(ctx, pipeline.GetStringValue(), project.GetStringValue(), cache)
			}
		}
	}
	// At least one auth source must be checked, and none must fail.
	return authed && !(labelsDenied || ppsLogMessageDenied || objectDenied)
}

func (ls LogService) authPPSLogMessage(ctx context.Context, msg *pps.LogMessage, cache map[string]bool) bool {
	pipeline := msg.GetPipelineName()
	project := msg.GetProjectName()
	return ls.authPipelineLogs(ctx, pipeline, project, cache)
}

func (ls LogService) authPipelineLogs(ctx context.Context, pipeline, project string, cache map[string]bool) bool {
	if ls.AuthServer == nil {
		return true
	}
	repo := &pfs.Repo{Type: pfs.UserRepoType, Project: &pfs.Project{Name: project}, Name: pipeline}
	if cache != nil {
		if pass, ok := cache[repo.String()]; ok {
			return pass
		}
	}
	var pass bool
	err := ls.AuthServer.CheckRepoIsAuthorized(ctx, repo, auth.Permission_REPO_READ)
	if err != nil && !auth.IsErrNotActivated(err) {
		log.Info(ctx, "unauthorized to retrieve logs for pipeline", zap.Error(err), zap.String("pipeline", project+"@"+pipeline))
		pass = false
	} else {
		pass = true
	}
	if cache != nil {
		cache[repo.String()] = pass
	}
	return pass
}

type ResponsePublisher interface {
	// Publish publishes a single GetLogsResponse to the client.
	Publish(context.Context, *logs.GetLogsResponse) error
}

// A passFunc receives a map of log labels to values and the log message, and
// return false if the log message does not pass — i.e., if it should be
// skipped.
type passFunc func(labels map[string]string, msg *logs.LogMessage) bool

// An adapter publishes log entries to a ResponsePublisher in a specified format.
type adapter struct {
	responsePublisher ResponsePublisher
	// If pass returns false, the message will not be published.
	pass passFunc
}

func (a *adapter) publish(ctx context.Context, labels loki.LabelSet, entry *loki.Entry) (bool, error) {
	if entry == nil {
		return false, nil
	}
	msg := &logs.LogMessage{
		Verbatim: &logs.VerbatimLogMessage{
			Line:      []byte(entry.Line),
			Timestamp: timestamppb.New(entry.Timestamp),
		},
		Object: new(structpb.Struct),
	}
	if strings.HasPrefix(entry.Line, "{") && strings.HasSuffix(entry.Line, "}") {
		if err := msg.Object.UnmarshalJSON([]byte(entry.Line)); err != nil {
			log.Debug(ctx, "failed to unmarshal json into protobuf Struct", zap.Error(err), zap.String("line", entry.Line))
			msg.Object = nil
		} else if val := msg.Object.Fields["time"].GetStringValue(); val != "" {
			if t, err := time.Parse(time.RFC3339Nano, val); err == nil {
				msg.NativeTimestamp = timestamppb.New(t)
			}
		}
		msg.PpsLogMessage = new(pps.LogMessage)
		m := protojson.UnmarshalOptions{
			AllowPartial:   true,
			DiscardUnknown: true,
		}
		if err := m.Unmarshal([]byte(entry.Line), msg.PpsLogMessage); err != nil {
			msg.PpsLogMessage = nil
		} else if msg.PpsLogMessage.Ts != nil {
			msg.NativeTimestamp = msg.PpsLogMessage.Ts
		}
	}
	if msg.Object == nil || msg.Object.Fields == nil {
		msg.Object = &structpb.Struct{Fields: make(map[string]*structpb.Value)}
	}
	for k, v := range labels {
		msg.Object.Fields["#"+k] = structpb.NewStringValue(v)
	}
	if len(msg.Object.AsMap()) == 0 {
		msg.Object = nil
	}
	if a.pass != nil && !a.pass(map[string]string(labels), msg) {
		return false, nil
	}
	if err := a.responsePublisher.Publish(ctx, &logs.GetLogsResponse{
		ResponseType: &logs.GetLogsResponse_Log{
			Log: msg,
		},
	}); err != nil {
		return false, errors.WithStack(fmt.Errorf("%w response with parsed json object: %w", ErrPublish, err))
	}
	return true, nil
}
