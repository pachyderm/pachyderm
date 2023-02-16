package server

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func propagateMetadata(ctx context.Context) context.Context {
	in, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	var pairs []string
	for k, vs := range in {
		for _, v := range vs {
			pairs = append(pairs, k, v)
		}
	}
	return metadata.AppendToOutgoingContext(ctx, pairs...)
}

func (s *debugServer) SetLogLevel(ctx context.Context, req *debug.SetLogLevelRequest) (*debug.SetLogLevelResponse, error) {
	result := new(debug.SetLogLevelResponse)
	d, err := types.DurationFromProto(req.GetDuration())
	if err != nil {
		return result, status.Errorf(codes.InvalidArgument, "invalid duration: %v", err)
	}
	switch x := req.GetLevel().(type) {
	case nil:
		return result, status.Error(codes.InvalidArgument, "no level provided")
	case *debug.SetLogLevelRequest_Grpc:
		switch x.Grpc {
		case debug.SetLogLevelRequest_DEBUG:
			log.SetGRPCLogLevelFor(zap.DebugLevel, d)
			log.Info(ctx, "set grpc log level to debug", zap.Duration("revert_after", d))
		case debug.SetLogLevelRequest_INFO:
			log.SetGRPCLogLevelFor(zap.InfoLevel, d)
			log.Info(ctx, "set grpc log level to info", zap.Duration("revert_after", d))
		case debug.SetLogLevelRequest_ERROR:
			log.SetGRPCLogLevelFor(zap.ErrorLevel, d)
			log.Info(ctx, "set grpc log level to error", zap.Duration("revert_after", d))
		case debug.SetLogLevelRequest_OFF:
			log.SetGRPCLogLevelFor(zap.FatalLevel, d)
			log.Info(ctx, "set grpc log level to fatal", zap.Duration("revert_after", d))
		default:
			return result, status.Errorf(codes.InvalidArgument, "cannot set grpc log level to %v", x.Grpc.String())
		}
	case *debug.SetLogLevelRequest_Pachyderm:
		switch x.Pachyderm {
		case debug.SetLogLevelRequest_DEBUG:
			log.SetLevelFor(log.DebugLevel, d)
			log.Debug(ctx, "set log level to debug", zap.Duration("revert_after", d))
		case debug.SetLogLevelRequest_INFO:
			log.SetLevelFor(log.InfoLevel, d)
			log.Info(ctx, "set log level to info", zap.Duration("revert_after", d))
		case debug.SetLogLevelRequest_ERROR:
			log.SetLevelFor(log.ErrorLevel, d)
			log.Error(ctx, "set log level to error", zap.Duration("revert_after", d))
		default:
			return result, status.Errorf(codes.InvalidArgument, "cannot set log level to %v", x.Pachyderm.String())
		}
	}

	if s.sidecarClient == nil {
		result.AffectedPods = append(result.AffectedPods, s.name)
	} else {
		result.AffectedPods = append(result.AffectedPods, s.name+".user")
	}

	// If this is the worker server, also adjust the storage sidecar.
	if cc := s.sidecarClient; cc != nil {
		tctx, c := context.WithTimeout(ctx, 5*time.Second)
		if _, err := cc.DebugClient.SetLogLevel(propagateMetadata(tctx), req); err != nil {
			result.ErroredPods = append(result.ErroredPods, fmt.Sprintf("%s.storage(%v)", s.name, err))
		} else {
			result.AffectedPods = append(result.AffectedPods, s.name+".storage")
		}
		c()
		return result, nil
	}

	if !req.GetRecurse() {
		// If not recursive mode, return now.
		return result, nil
	}
	req.Recurse = false

	// Recurse to other pachyderm processes.
	pods := map[string]string{}
	apps := map[string]string{
		"pach-enterprise": strconv.Itoa(int(s.env.Config().Port)),
		//"pachw": strconv.Itoa(int(s.env.Config().PeerPort)),
		"pachd":    strconv.Itoa(int(s.env.Config().Port)),
		"pipeline": os.Getenv(client.PPSWorkerPortEnv),
	}
	var enumerateErrs error
	for app, port := range apps {
		tctx, c := context.WithTimeout(ctx, 30*time.Second)
		podList, err := s.env.GetKubeClient().CoreV1().Pods(s.env.Config().Namespace).List(
			tctx,
			metav1.ListOptions{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ListOptions",
					APIVersion: "v1",
				},
				LabelSelector: metav1.FormatLabelSelector(
					metav1.SetAsLabelSelector(
						map[string]string{
							"suite": "pachyderm",
							"app":   app,
						},
					),
				),
			},
		)
		c()
		if err != nil {
			multierr.AppendInto(&enumerateErrs, errors.Wrapf(err, "ListPods(%v)", app))
			continue
		}
		for _, pod := range podList.Items {
			pods[pod.Name] = fmt.Sprintf("%s:%s", pod.Status.PodIP, port)
		}
	}
	for pod, addr := range pods {
		if pod == s.name {
			continue // skip self
		}
		res, err := propagateLogLevel(ctx, req, pod, addr)
		if err != nil {
			result.ErroredPods = append(result.ErroredPods, fmt.Sprintf("%v@%v(%v)", pod, addr, err))
			continue
		}
		result.AffectedPods = append(result.AffectedPods, res.GetAffectedPods()...)
		// This picks up recursive failures from workers; user container succeeded, storage
		// container failed.
		result.ErroredPods = append(result.ErroredPods, res.GetErroredPods()...)
	}
	if enumerateErrs != nil {
		return result, status.Errorf(codes.Unavailable, "some pods could not be enumerated: %v", enumerateErrs)
	}
	return result, nil
}

func propagateLogLevel(ctx context.Context, req *debug.SetLogLevelRequest, pod, addr string) (_ *debug.SetLogLevelResponse, retErr error) {
	ctx, c := context.WithTimeout(ctx, 5*time.Second)
	defer c()
	defer log.Span(ctx, fmt.Sprintf("propagateLogLevel(%s)", pod))(log.Errorp(&retErr))
	opts := client.DefaultDialOptions()
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	cc, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "dial")
	}
	defer cc.Close() //nolint:errcheck (Only returns an error if the cc is already closing.)
	client := debug.NewDebugClient(cc)
	res, err := client.SetLogLevel(propagateMetadata(ctx), req)
	if err != nil {
		return nil, errors.Wrap(err, "SetLogLevel")
	}
	return res, nil
}
