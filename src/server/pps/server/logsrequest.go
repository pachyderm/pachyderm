package server

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/getlogs"
)

type getLogsV2Handler struct {
  apiServer *apiServer
  maxLines uint64
  getLogs *getlogs.GetLogs
  kubeLogSource *getlogs.KubeLogSource
  lineMatchers *getlogs.MatchLines
}

func (handler *getLogsV2Handler) handlePipelineJob(ctx context.Context, query *pps.PipelineJobLogQuery) {
  if query == nil || query.GetPipeline == nil {
    return
  }
  handler.lineMatchers.MatchPipelineJob(query.GetPipeline().GetProject(), query.GetPipeline().GetPipeline(), query.GetJob())
}

func (handler *getLogsV2Handler) handlePipeline(ctx context.Context, query *pps.PipelineLogQuery) {
  if query == nil {
    return
  }
  handler.lineMatchers.MatchPipeline(query.GetProject(), query.GetPipeline())
}

func (handler *getLogsV2Handler) handleUserLogQuery(ctx context.Context, query *pps.UserLogQuery) {
  if query == nil {
    return
  }
  handler.lineMatchers.MatchProject(query.GetProject())

  handler.handlePipeline(ctx, query.GetPipeline())

  handler.lineMatchers.MatchDatum(query.GetDatum())
  handler.lineMatchers.MatchJob(query.GetJob())

  handler.handlePipelineJob(ctx, query.GetPipelineJob())
}

func (handler *getLogsV2Handler) handleAdminLogQuery(ctx context.Context, query *pps.AdminLogQuery) error {
  if query == nil {
    return nil
  }
  if pod := query.GetPod(); pod != "" {
    streamers, err := handler.kubeLogSource.GetPodStreamers(ctx, pod, "")
    if err != nil {
      return err
    }
    handler.getLogs.Streamers = append(handler.getLogs.Streamers, streamers...)
    return nil
  }
  if pod := query.GetPodContainer(); pod != nil {
    streamers, err := handler.kubeLogSource.GetPodStreamers(ctx, pod.GetPod(), pod.GetContainer())
    if err != nil {
      return err
    }
    handler.getLogs.Streamers = append(handler.getLogs.Streamers, streamers...)
    return nil
  }
  if master := query.GetMaster(); master != nil {
    streamers, err := handler.kubeLogSource.GetPodStreamers(ctx, "", "")
    if err != nil {
      return err
    }
    handler.getLogs.Streamers = append(handler.getLogs.Streamers, streamers...)
    handler.lineMatchers.MatchMaster(master.GetProject(), master.GetPipeline())
    return nil
  }
  if storage := query.GetStorage(); storage != nil {
    streamers, err := handler.kubeLogSource.GetPodStreamers(ctx, "", "")
    if err != nil {
      return err
    }
    handler.getLogs.Streamers = append(handler.getLogs.Streamers, streamers...)
    handler.lineMatchers.MatchStorage(storage.GetProject(), storage.GetPipeline())
    return nil
  }
  handler.handleUserLogQuery(ctx, query.GetUser())

  return nil
}

func (handler *getLogsV2Handler) handleQuery(ctx context.Context, query *pps.LogQuery) error {

  if query == nil {
    streamers, err := handler.kubeLogSource.GetPodStreamers(ctx, "", "")
    if err != nil {
      return err
    }
    handler.getLogs.Streamers = append(handler.getLogs.Streamers, streamers...)
    return nil
  }
  handler.handleUserLogQuery(ctx, query.GetUser())
  handler.handleAdminLogQuery(ctx, query.GetAdmin())

  return nil

}

func (handler *getLogsV2Handler) handleFilters(filters *pps.LogFilter) error {
  if filters == nil {
    return nil
  }
  handler.maxLines = filters.Limit
  handler.lineMatchers.MatchLogLevel(uint32(filters.Level))

  if filters.Regex != nil {
    if err := handler.lineMatchers.MatchRegex(filters.Regex.Pattern, filters.Regex.Negate); err != nil {
      return err
    }
  }

  if filters.TimeRange != nil {
    var fromTime time.Time
    untilTime := time.Now()
    if filters.TimeRange.From != nil {
      fromTime = filters.TimeRange.From.AsTime()
      handler.kubeLogSource.FromTime = fromTime
    }
    if filters.TimeRange.Until != nil {
      untilTime = filters.TimeRange.Until.AsTime()
    }
    if err := handler.lineMatchers.MatchTime(fromTime, untilTime); err != nil {
      return err
    }
  }

  return nil

}

// handles GetLogsV2 request
func (handler *getLogsV2Handler) handleRequest(ctx context.Context, request *pps.GetLogsV2Request) (<-chan []byte, error) {

  // initialize GetLogsV2 backend data structures to fill and use in the handlers for the different parts of the request
  handler.lineMatchers = new(getlogs.MatchLines)
  handler.kubeLogSource = &getlogs.KubeLogSource{
    Kube: handler.apiServer.env.KubeClient,
    Namespace: handler.apiServer.namespace,
    Follow: request.Tail,
  }

  handler.getLogs = &getlogs.GetLogs{
    Streamers: []getlogs.LogSource{},
    Matchers: handler.lineMatchers,
    LockStreams: request.Tail,
  }

  err := handler.handleFilters(request.Filter)
  if err != nil {
    return nil, err
  }

  err = handler.handleQuery(ctx, request.Query)
  if err != nil {
    return nil, err
  }

  maxLines := handler.maxLines
	outChannel := make(chan []byte)
  var numLines uint64

	var eg errgroup.Group
  handler.getLogs.StartMatching(ctx, eg, func(line []byte) error {

    select {
    case outChannel <- line:
      numLines++
      if numLines == maxLines {
        close(outChannel)
        return errutil.ErrBreak
      }
      return nil
    case <- ctx.Done():
      return errutil.ErrBreak
    }
  })

	return outChannel, nil

}

