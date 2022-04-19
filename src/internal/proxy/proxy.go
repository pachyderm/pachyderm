package proxy

import (
	"io"
	"net/http"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/randutil"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Aggregated struct {
	GRPCServer            *grpc.Server
	S3Handler, DexHandler http.Handler
}

func (p *Aggregated) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r == nil {
		http.Error(w, "nil request", http.StatusInternalServerError)
		return
	}
	l := logrus.WithContext(r.Context()).WithFields(logrus.Fields{
		"source":     "http-proxy",
		"request_id": randutil.UniqueString(""),
	})
	l.WithFields(logrus.Fields{
		"http.request.proto_major": r.ProtoMajor,
		"http.request.uri":         r.URL.String(),
		"http.request.method":      r.Method,
	}).Debug("proxy request")

	if r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("content-type"), "application/grpc") {
		l.Debug("grpc request")
		p.GRPCServer.ServeHTTP(w, r)
		l.Debug("served grpc")
		return
	}
	if strings.HasPrefix(r.Header.Get("authorization"), "AWS4-HMAC-SHA256") {
		l.Debug("s3 gateway request")
		p.S3Handler.ServeHTTP(w, r)
		l.Debug("served s3 gateway request")
		return
	}
	if strings.HasPrefix(r.URL.String(), "/dex") {
		l.Debug("dex request")
		p.DexHandler.ServeHTTP(w, r)
		l.Debug("served dex")
		return
	}
	if strings.HasPrefix(r.URL.String(), "/authorization-code/callback") {
		l.Debug("oidc request")
		// Bug: CORE-467
		http.DefaultServeMux.ServeHTTP(w, r)
		l.Debug("served oidc")
		return
	}
	l.Debug("console request")
	upreq, err := http.NewRequestWithContext(r.Context(), r.Method, "http://console.default.svc.cluster.local:4000"+r.URL.Path, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		l.WithError(err).Error("create console request")
		return
	}
	res, err := http.DefaultClient.Do(upreq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		l.WithError(err).Error("do console request")
		return
	}
	for k, v := range res.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(res.StatusCode)
	if _, err := io.Copy(w, res.Body); err != nil {
		l.WithError(err).Error("copy body")
	}
	res.Body.Close()
	l.Debug("served console")
}
