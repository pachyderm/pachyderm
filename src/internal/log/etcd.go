package log

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// AddLoggerToEtcdServer adds the context's logger to an in-memory etcd server.
func AddLoggerToEtcdServer(ctx context.Context, config *embed.Config) zap.AtomicLevel {
	lvl := zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	config.ZapLoggerBuilder = embed.NewZapLoggerBuilder(extractLogger(ctx).WithOptions(zap.IncreaseLevel(lvl), zap.AddCallerSkip(-1)).Named("etcd-server"))
	return lvl
}

// GetEtcdClientConfig returns an etcd client configuration object with the logger set to this context's logger.
func GetEtcdClientConfig(ctx context.Context) clientv3.Config {
	l := extractLogger(ctx).Named("etcd-client")
	if lvl := zap.InfoLevel; l.Level() <= lvl {
		l = l.WithOptions(zap.IncreaseLevel(lvl))
	}
	l = l.WithOptions(zap.AddCallerSkip(-1)) // This logger won't be using our logging methods.
	return clientv3.Config{
		Context: withLogger(ctx, l),
		Logger:  l,
	}
}
