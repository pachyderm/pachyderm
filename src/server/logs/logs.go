// Package logs implements the log service gRPC server.
package logs

import (
	"context"
	"encoding/base64"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/logs"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil"
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
	adapter.addLevelFilter(filter.GetLevel())

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
		// The caller of compileRequest ensures that Filter is not nil.  After
		// compileRequest, we re-read the level to add a global level filter.
		if f := request.GetFilter(); f != nil && f.GetLevel() == logs.LogLevel_LOG_LEVEL_UNSET {
			// If the log level filter is unset, we default to INFO level logs.  The
			// worker defaults to running at DEBUG now and this avoids spam unless
			// explicitly requested.
			request.Filter.Level = logs.LogLevel_LOG_LEVEL_INFO
		}
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
	level             logs.LogLevel
	// If pass returns false, the message will not be published.
	pass passFunc
}

func (a *adapter) addLevelFilter(l logs.LogLevel) {
	if l != logs.LogLevel_LOG_LEVEL_UNSET && a.level == logs.LogLevel_LOG_LEVEL_UNSET {
		a.level = l
	}
}

func stringAsLogLevel(s string) logs.LogLevel {
	switch s {
	case "debug":
		return logs.LogLevel_LOG_LEVEL_DEBUG
	case "info":
		return logs.LogLevel_LOG_LEVEL_INFO
	default:
		return logs.LogLevel_LOG_LEVEL_ERROR
	}

}

func (a *adapter) publish(ctx context.Context, labels loki.LabelSet, entry *loki.Entry) (bool, error) {
	if entry == nil {
		return false, nil
	}

	// Fix some Loki configuration errors that would prevent parsing.
	var obj map[string]any
	entry.Line, obj = lokiutil.RepairLine(entry.Line)

	// If the line is not JSON, synthesize some.
	if obj == nil {
		obj = map[string]any{
			"message":  entry.Line,
			"time":     entry.Timestamp.Format(time.RFC3339Nano),
			"severity": "info",
		}
	}

	// Build the base object to return.
	msg := &logs.LogMessage{
		Verbatim: &logs.VerbatimLogMessage{
			Line:      []byte(entry.Line),
			Timestamp: timestamppb.New(entry.Timestamp),
		},
		Object:          new(structpb.Struct),
		NativeTimestamp: timestamppb.New(entry.Timestamp),
	}

	// If there is JSON, convert it to a structpb.Struct, or synthesize something.
	var err error
	msg.Object, err = structpb.NewStruct(obj)
	if err != nil {
		msg.Object = &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"message":  structpb.NewStringValue(entry.Line),
				"time":     structpb.NewStringValue(entry.Timestamp.Format(time.RFC3339Nano)),
				"severity": structpb.NewStringValue("error"),
				"#error":   structpb.NewStringValue(err.Error()),
			},
		}
		obj["severity"] = "error"
	}

	// Grab the severity from the log message, for filtering below.
	severity := logs.LogLevel_LOG_LEVEL_ERROR
	if v, ok := obj["severity"]; ok {
		if s, ok := v.(string); ok {
			severity = stringAsLogLevel(s)
		}
	}

	// Add loki metadata to the object.
	for k, v := range labels {
		msg.Object.Fields["#"+k] = structpb.NewStringValue(v)
	}

	// Extract the native timestamp, if possible.
	if v, ok := obj["time"]; ok {
		if sv, ok := v.(string); ok {
			if t, err := time.Parse(time.RFC3339Nano, sv); err == nil && !t.IsZero() {
				msg.NativeTimestamp = timestamppb.New(t)
			}
		}
	}
	if v, ok := obj["ts"]; ok {
		if sv, ok := v.(string); ok {
			if t, err := time.Parse(time.RFC3339Nano, sv); err == nil && !t.IsZero() {
				msg.NativeTimestamp = timestamppb.New(t)
			}
		}
	}
	if msg.NativeTimestamp == nil || msg.NativeTimestamp.AsTime().IsZero() {
		msg.NativeTimestamp = nil
	}

	// Convert the object to a PPS log message for backwards compatibility.
	msg.PpsLogMessage = objectToLogMessage(obj)
	if msg.PpsLogMessage.GetTs() == nil && msg.NativeTimestamp != nil {
		msg.PpsLogMessage.Ts = msg.NativeTimestamp
	}
	if proto.Equal(msg.PpsLogMessage, &pps.LogMessage{}) {
		// If the PPS log message didn't capture any fields, remove it entirely.  This is
		// rare; it will only happen for valid JSON objects that don't contain "ts" and
		// "message".
		msg.PpsLogMessage = nil
	}

	// Filter by log level, if requested.
	if a.level == logs.LogLevel_LOG_LEVEL_INFO && severity < logs.LogLevel_LOG_LEVEL_INFO {
		return false, nil
	}
	if a.level == logs.LogLevel_LOG_LEVEL_ERROR && severity < logs.LogLevel_LOG_LEVEL_ERROR {
		return false, nil
	}

	// Run other filters.
	if a.pass != nil && !a.pass(map[string]string(labels), msg) {
		return false, nil
	}

	// Finally, write out the unfiltered line.
	if err := a.responsePublisher.Publish(ctx, &logs.GetLogsResponse{
		ResponseType: &logs.GetLogsResponse_Log{
			Log: msg,
		},
	}); err != nil {
		return false, errors.WithStack(fmt.Errorf("%w response with parsed json object: %w", ErrPublish, err))
	}
	return true, nil
}

func extract[T any](obj map[string]any, dst *T, key string) {
	if raw, ok := obj[key]; ok {
		if v, ok := raw.(T); ok {
			*dst = v
		}
	}
}

func objectToLogMessage(obj map[string]any) *pps.LogMessage {
	result := new(pps.LogMessage)
	extract(obj, &result.ProjectName, "projectName")
	extract(obj, &result.PipelineName, "pipelineName")
	extract(obj, &result.JobId, "jobId")
	extract(obj, &result.WorkerId, "workerId")
	extract(obj, &result.DatumId, "datumId")
	extract(obj, &result.Master, "master")
	extract(obj, &result.User, "user")
	extract(obj, &result.Message, "message")

	var ts string
	extract(obj, &ts, "ts")
	if t, err := time.Parse(time.RFC3339Nano, ts); err == nil {
		result.Ts = timestamppb.New(t)
	}

	var dataList []any
	extract(obj, &dataList, "data")
	for _, v := range dataList {
		data, ok := v.(map[string]any)
		if !ok {
			continue
		}
		file := new(pps.InputFile)
		var hashString string
		extract(data, &hashString, "hash")
		file.Hash, _ = base64.StdEncoding.DecodeString(hashString)
		extract(data, &file.Path, "path")
		result.Data = append(result.Data, file)
	}

	return result
}
