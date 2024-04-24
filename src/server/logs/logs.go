package logs

import (
	"context"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/logs"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
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

type logDirection string

const (
	forwardLogDirection  logDirection = "forward"
	backwardLogDirection logDirection = "backward"
)

// GetLogs gets logs according its request and publishes them.  The pattern is
// similar to that used when handling an HTTP request.
func (ls LogService) GetLogs(ctx context.Context, request *logs.GetLogsRequest, publisher ResponsePublisher) error {
	var direction = forwardLogDirection

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
	switch {
	case filter.TimeRange == nil || (filter.TimeRange.From == nil && filter.TimeRange.Until == nil):
		now := time.Now()
		filter.TimeRange = &logs.TimeRangeLogFilter{
			From:  timestamppb.New(now.Add(-700 * time.Hour)),
			Until: timestamppb.New(now),
		}
	case filter.TimeRange.From == nil:
		filter.TimeRange.From = timestamppb.New(filter.TimeRange.Until.AsTime().Add(-700 * time.Hour))
	case filter.TimeRange.Until == nil:
		filter.TimeRange.Until = timestamppb.New(filter.TimeRange.From.AsTime().Add(700 * time.Hour))
	}

	c, err := ls.GetLokiClient()
	if err != nil {
		return errors.Wrap(err, "loki client error")
	}
	start := filter.TimeRange.From.AsTime()
	end := filter.TimeRange.Until.AsTime()
	if start.Equal(end) {
		return errors.Errorf("start equals end (%v)", start)
	}
	if start.After(end) {
		direction = backwardLogDirection
		start, end = end, start
	}

	var (
		adapter = adapter{responsePublisher: publisher}
		logQL   string
	)
	logQL, adapter.pass, err = ls.compileRequest(ctx, request)
	if err != nil {
		return errors.Wrap(err, "cannot convert request to LogQL")
	}
	if err = doQuery(ctx, c, logQL, int(filter.Limit), start, end, direction, adapter.publish); err != nil {
		var invalidBatchSizeErr ErrInvalidBatchSize
		switch {
		case errors.As(err, &invalidBatchSizeErr):
			// try to requery
			err = doQuery(ctx, c, logQL, invalidBatchSizeErr.RecommendedBatchSize(), start, end, direction, adapter.publish)
			if err != nil {
				return errors.Wrap(err, "invalid batch size requery failed")
			}
		default:
			return errors.Wrap(err, "doQuery failed")
		}
	}
	if request.WantPagingHint {
		var newer, older loki.Entry
		switch direction {
		case forwardLogDirection:
			newer = adapter.last
		case backwardLogDirection:
			newer = adapter.first
		default:
			return errors.Errorf("invalid direction %q", direction)
		}
		// request a record immediately prior to the page
		err := doQuery(ctx, c, logQL, 1, start.Add(-700*time.Hour), start, backwardLogDirection, func(ctx context.Context, e loki.Entry) (bool, error) {
			older = e
			return false, nil
		})
		if err != nil {
			return errors.Wrap(err, "hint doQuery failed")
		}
		hint := &logs.PagingHint{
			Older: proto.Clone(request).(*logs.GetLogsRequest),
			Newer: proto.Clone(request).(*logs.GetLogsRequest),
		}
		if !older.Timestamp.IsZero() {
			hint.Older.Filter.TimeRange.Until = timestamppb.New(older.Timestamp)
		}
		if !newer.Timestamp.IsZero() {
			hint.Newer.Filter.TimeRange.From = timestamppb.New(newer.Timestamp)
		}
		if request.Filter.TimeRange.From != nil && request.Filter.TimeRange.Until != nil {
			delta := request.Filter.TimeRange.Until.AsTime().Sub(request.Filter.TimeRange.From.AsTime())
			if !older.Timestamp.IsZero() {
				hint.Older.Filter.TimeRange.From = timestamppb.New(older.Timestamp.Add(-delta))
			}
			if !newer.Timestamp.IsZero() {
				hint.Newer.Filter.TimeRange.Until = timestamppb.New(newer.Timestamp.Add(delta))
			}
		}
		if err := publisher.Publish(ctx, &logs.GetLogsResponse{
			ResponseType: &logs.GetLogsResponse_PagingHint{
				PagingHint: hint,
			},
		}); err != nil {
			return errors.WithStack(fmt.Errorf("%w paging hint: %w", ErrPublish, err))
		}
	}

	return nil
}

func (ls LogService) compileRequest(ctx context.Context, request *logs.GetLogsRequest) (string, func(*logs.LogMessage) bool, error) {
	if request == nil {
		return "", nil, errors.New("nil request")
	}
	query := request.Query
	if query == nil {
		return "", nil, nil
	}
	switch query := query.QueryType.(type) {
	case *logs.LogQuery_User:
		return ls.compileUserLogQueryReq(ctx, query.User, true)
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
			return query.Logql, func(*logs.LogMessage) bool { return true }, nil
		case *logs.AdminLogQuery_Pod:
			if query.Pod == "" {
				return "", nil, adminLogQueryValidateErr("Pod", "Pod")
			}
			return fmt.Sprintf(`{suite="pachyderm",pod=%q}`, query.Pod), func(*logs.LogMessage) bool { return true }, nil
		case *logs.AdminLogQuery_PodContainer:
			if query.PodContainer.Pod == "" {
				return "", nil, adminLogQueryValidateErr("PodContainer", "Pod")
			}
			if query.PodContainer.Container == "" {
				return "", nil, adminLogQueryValidateErr("PodContainer", "Container")
			}
			return fmt.Sprintf(`{suite="pachyderm",pod=%q,container=%q}`, query.PodContainer.Pod, query.PodContainer.Container),
				func(*logs.LogMessage) bool { return true }, nil
		case *logs.AdminLogQuery_App:
			if query.App == "" {
				return "", nil, adminLogQueryValidateErr("App", "App")
			}
			return fmt.Sprintf(`{suite="pachyderm",app=%q}`, query.App), func(*logs.LogMessage) bool { return true }, nil
		case *logs.AdminLogQuery_Master:
			pipeline := query.Master.Pipeline
			project := query.Master.Project
			if pipeline == "" {
				return "", nil, adminLogQueryValidateErr("Master", "Pipeline")
			}
			if project == "" {
				return "", nil, adminLogQueryValidateErr("Master", "Project")
			}
			return fmt.Sprintf(`{app="pipeline",suite="pachyderm",container="user",pipelineProject=%q,pipelineName=%q}`, project, pipeline), func(msg *logs.LogMessage) bool {
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
				func(*logs.LogMessage) bool {
					return true
				}, nil
		case *logs.AdminLogQuery_User:
			return ls.compileUserLogQueryReq(ctx, query.User, false)
		default:
			return "", nil, nil
		}
	default:
		return "", nil, errors.Wrapf(ErrUnimplemented, "%T", query)
	}
}

func (ls LogService) compileUserLogQueryReq(ctx context.Context, query *logs.UserLogQuery, checkAuth bool) (string, func(*logs.LogMessage) bool, error) {
	switch query := query.GetUserType().(type) {
	case *logs.UserLogQuery_Pipeline:
		pipeline := query.Pipeline.Pipeline
		project := query.Pipeline.Project
		if checkAuth {
			if pass := ls.authPipelineLogs(ctx, pipeline, project, nil); !pass {
				return "", nil, nil
			}
		}
		return ls.compilePipelineLogsReq(project, pipeline)
	case *logs.UserLogQuery_Datum:
		authCache := make(map[string]bool)
		return ls.compileDatumsLogsReq(ctx, query.Datum, checkAuth, authCache)
	case *logs.UserLogQuery_Project:
		authCache := make(map[string]bool)
		return ls.compileProjectLogsReq(ctx, query.Project, checkAuth, authCache)
	case *logs.UserLogQuery_Job:
		authCache := make(map[string]bool)
		return ls.compileJobLogsReq(ctx, query.Job, checkAuth, authCache)
	case *logs.UserLogQuery_PipelineJob:
		project := query.PipelineJob.Pipeline.Project
		pipeline := query.PipelineJob.Pipeline.Pipeline
		if checkAuth {
			if pass := ls.authPipelineLogs(ctx, pipeline, project, nil); !pass {
				return "", nil, nil
			}
		}
		job := query.PipelineJob.Job
		return ls.compilePipelineJobLogsReq(project, pipeline, job)
	default:
		return "", nil, errors.Wrapf(ErrUnimplemented, "%T", query)
	}
}

func (ls LogService) compilePipelineLogsReq(project, pipeline string) (string, func(*logs.LogMessage) bool, error) {
	if pipeline == "" {
		return "", nil, userLogQueryValidateErr("Pipeline", "Pipeline")
	}
	if project == "" {
		return "", nil, userLogQueryValidateErr("Pipeline", "Project")
	}
	return fmt.Sprintf(`{app="pipeline",suite="pachyderm",container="user",pipelineProject=%q,pipelineName=%q}`, project, pipeline),
		func(msg *logs.LogMessage) bool {
			if msg.GetPpsLogMessage().GetUser() {
				return true
			}
			ff := msg.GetObject().GetFields()
			if ff != nil {
				v, ok := ff["user"]
				if ok {
					if v.GetBoolValue() {
						return true
					}
				}
			}
			return false
		}, nil
}

func (ls LogService) compileDatumsLogsReq(ctx context.Context, datum string, checkAuth bool, authCache map[string]bool) (string, func(*logs.LogMessage) bool, error) {
	if datum == "" {
		return "", nil, userLogQueryValidateErr("Datum", "Datum")
	}
	return fmt.Sprintf(`{suite="pachyderm",app="pipeline"} |= %q`, datum), func(msg *logs.LogMessage) bool {
		if msg.GetPpsLogMessage().GetDatumId() != "" && msg.GetPpsLogMessage().GetDatumId() != datum {
			return false
		}
		ff := msg.GetObject().GetFields()
		if ff != nil {
			v, ok := ff["datumId"]
			if ok {
				if v.GetStringValue() != datum {
					return false
				}
			} else {
				return false
			}
		}
		if checkAuth {
			return ls.authLogMessage(ctx, msg, authCache)
		}
		return true
	}, nil
}

func (ls LogService) compileProjectLogsReq(ctx context.Context, project string, checkAuth bool, authCache map[string]bool) (string, func(*logs.LogMessage) bool, error) {
	if project == "" {
		return "", nil, userLogQueryValidateErr("Project", "Project")
	}
	return fmt.Sprintf(`{suite="pachyderm",app="pipeline",pipelineProject=%q}`, project), func(msg *logs.LogMessage) bool {
		if checkAuth {
			return ls.authLogMessage(ctx, msg, authCache)
		}
		return true
	}, nil
}

func (ls LogService) compileJobLogsReq(ctx context.Context, job string, checkAuth bool, authCache map[string]bool) (string, func(*logs.LogMessage) bool, error) {
	if job == "" {
		return "", nil, userLogQueryValidateErr("Job", "Job")
	}
	return `{suite="pachyderm",app="pipeline"}`, func(msg *logs.LogMessage) bool {
		if msg.GetPpsLogMessage().GetJobId() != "" && msg.GetPpsLogMessage().GetJobId() != job {
			return false
		}
		ff := msg.GetObject().GetFields()
		if ff != nil {
			v, ok := ff["jobId"]
			if ok {
				if v.GetStringValue() != job {
					return false
				}
			} else {
				return false
			}
		}
		if checkAuth {
			return ls.authLogMessage(ctx, msg, authCache)
		}
		return true
	}, nil
}

func (ls LogService) compilePipelineJobLogsReq(project, pipeline, job string) (string, func(*logs.LogMessage) bool, error) {
	if project == "" {
		return "", nil, userLogQueryValidateErr("PipelineJob", "Project")
	}
	if pipeline == "" {
		return "", nil, userLogQueryValidateErr("PipelineJob", "Pipeline")
	}
	if job == "" {
		return "", nil, userLogQueryValidateErr("PipelineJob", "Job")
	}
	return fmt.Sprintf(`{suite="pachyderm",pipelineProject=%q,pipelineName=%q}`, project, pipeline), func(msg *logs.LogMessage) bool {
		if msg.GetPpsLogMessage().JobId == job {
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
	}, nil

}

func (ls LogService) authLogMessage(ctx context.Context, msg *logs.LogMessage, cache map[string]bool) bool {
	pipeline := msg.GetPpsLogMessage().GetPipelineName()
	project := msg.GetPpsLogMessage().GetProjectName()
	if pipeline != "" && project != "" {
		return ls.authPipelineLogs(ctx, pipeline, project, cache)
	}
	ff := msg.GetObject().GetFields()
	if ff != nil {
		pipeline, ok := ff["pipelineName"]
		if ok {
			project, ok := ff["projectName"]
			if ok {
				return ls.authPipelineLogs(ctx, pipeline.GetStringValue(), project.GetStringValue(), cache)
			}
		}
	}
	return false
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
	pass := true
	if err := ls.AuthServer.CheckRepoIsAuthorized(ctx, repo, auth.Permission_REPO_READ); err != nil && !auth.IsErrNotActivated(err) {
		log.Info(ctx, "unauthorized to retrieve logs for pipeline", zap.Error(err), zap.String("pipeline", project+"@"+pipeline))
		pass = false
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

// An adapter publishes log entries to a ResponsePublisher in a specified format.
type adapter struct {
	responsePublisher ResponsePublisher
	first, last       loki.Entry
	gotFirst          bool
	pass              func(*logs.LogMessage) bool
}

func (a *adapter) publish(ctx context.Context, entry loki.Entry) (bool, error) {
	var msg = &logs.LogMessage{
		Verbatim: &logs.VerbatimLogMessage{
			Line:      []byte(entry.Line),
			Timestamp: timestamppb.New(entry.Timestamp),
		},
	}
	if !a.gotFirst {
		a.gotFirst = true
		a.first = entry
	}
	msg.Object = new(structpb.Struct)
	if err := msg.Object.UnmarshalJSON([]byte(entry.Line)); err != nil {
		log.Error(ctx, "failed to unmarshal json into protobuf Struct", zap.Error(err), zap.String("line", entry.Line))
		msg.Object = nil
	} else {
		if val := msg.Object.Fields["time"].GetStringValue(); val != "" {
			if t, err := time.Parse(time.RFC3339Nano, val); err == nil {
				msg.NativeTimestamp = timestamppb.New(t)
			}
		}
		if val := msg.Object.Fields["log"].GetStringValue(); val != "" {
			obj := new(structpb.Struct)
			if err := obj.UnmarshalJSON([]byte(val)); err != nil {
				log.Error(ctx, "failed to unmarshal json into protobuf Struct", zap.Error(err), zap.String("log", val))
			} else {
				msg.Object = obj
			}
		}
	}
	msg.PpsLogMessage = new(pps.LogMessage)
	m := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: true,
	}
	if err := m.Unmarshal([]byte(entry.Line), msg.PpsLogMessage); err != nil {
		log.Error(ctx, "failed to unmarshal json into PpsLogMessage", zap.Error(err), zap.String("line", entry.Line))
		msg.PpsLogMessage = nil
	} else if msg.PpsLogMessage.Ts != nil {
		msg.NativeTimestamp = msg.PpsLogMessage.Ts
	}

	if a.pass != nil && !a.pass(msg) {
		return true, nil
	}
	if err := a.responsePublisher.Publish(ctx, &logs.GetLogsResponse{
		ResponseType: &logs.GetLogsResponse_Log{
			Log: msg,
		},
	}); err != nil {
		return false, errors.WithStack(fmt.Errorf("%w response with parsed json object: %w", ErrPublish, err))
	}
	return false, nil
}
