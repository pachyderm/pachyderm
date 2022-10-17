// A proxy to the real S3 server underlying pachyderm, for cases when stronger
// S3 compatibility & performance is required for "s3_out" feature.

package server

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	awsauth "github.com/smartystreets/go-aws-auth"
)

type RawS3Proxy struct {
}

// XXX this shouldn't be a global variable, it's just convenient for this PoC..
// e.g. "bucketname"
var CurrentBucket string = "out"

// This is like pipeline_name-<job-id>
var CurrentTargetPath string = ""

// Autodetected based on proxy traffic, will be like "example-data-24" or
// whatever path the user specified after "s3a://out/<path>"
var LastSeenPathToReplace string = ""

func (r *RawS3Proxy) ListenAndServe(port uint16) error {
	// os.GetEnv("STORAGE_BACKEND") ==> "MINIO" or, presumably, "AWS"...
	// MINIO_SECRET, MINIO_ID, MINIO_ENDPOINT (e.g.
	// minio.default.svc.cluster.local:9000), MINIO_SECURE, MINIO_BUCKET also
	// set.

	// TODO!!! Need to mutate the bucket name in the proxied requests from
	// "out" to whatever the real backend bucket is called in the backend...
	// Before we do this, we'll just modify the user code to use the real
	// minio bucket name, just to prove the concept.

	if os.Getenv("STORAGE_BACKEND") != "MINIO" {
		panic("only minio supported by proxy to real backend s3_out feature right now - TODO: add real AWS!")
	}

	proxyRouter := mux.NewRouter()
	proxyRouter.PathPrefix("/").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			if LastSeenPathToReplace == "" {
				// match the first Spark request per job...
				// HEAD /out/example-data-24 HTTP/1.1
				logrus.Infof("PROXY r.Method='%s', r.RequestURI='%s'", r.Method, r.RequestURI)
				if r.Method == "HEAD" && strings.HasPrefix(r.RequestURI, "/out/") {
					LastSeenPathToReplace = r.RequestURI[len("/out/"):]
					logrus.Infof("PROXY LastSeenPathToReplace: '%s'", LastSeenPathToReplace)
				}
			}

			mashup := func(s string) string {
				// danger danger, this will probably mash too much in some cases
				ret := strings.Replace(s, "/out", "/"+CurrentBucket, -1)
				// XML stylee as well
				ret = strings.Replace(ret, ">out", ">"+CurrentBucket, -1)
				if LastSeenPathToReplace != "" {
					// XXX SECURITY: Think about how inferring
					// LastSeenPathToReplace based on user generated traffic may
					// allow access to the whole bucket
					p := ret
					ret = strings.Replace(ret, LastSeenPathToReplace, CurrentTargetPath+"/"+LastSeenPathToReplace, -1)
					if p != ret {
						logrus.Infof("PROXY transformed '%s' to '%s'", p, ret)
					}
				} else {
					if len(s) <= 100 {
						logrus.Infof("PROXY Warning! LastSeenPathToReplace was empty when mashing up '%s'", s)
					}
				}
				// logrus.Infof("MASHUP '%s' ==> '%s'", s, ret)
				return ret
			}

			unmashup := func(s string) string {
				// danger danger, this will probably mash too much in some cases
				ret := strings.Replace(s, "/"+CurrentBucket, "/out", -1)
				// XML stylee as well
				ret = strings.Replace(ret, ">"+CurrentBucket, ">out", -1)
				if LastSeenPathToReplace != "" {
					// XXX SECURITY: Think about how inferring
					// LastSeenPathToReplace based on user generated traffic may
					// allow access to the whole bucket
					p := ret
					ret = strings.Replace(ret, CurrentTargetPath+"/"+LastSeenPathToReplace, LastSeenPathToReplace, -1)
					if p != ret {
						logrus.Infof("PROXY reverse transformed '%s' to '%s'", p, ret)
					}
				} else {
					if len(s) <= 100 {
						logrus.Infof("PROXY Warning! LastSeenPathToReplace was empty when unmashing '%s'", s)
					}
				}
				// logrus.Infof("MASHUP '%s' ==> '%s'", s, ret)
				return ret
			}

			// logrus.Infof("===> IN PROXY ROUTER, w=%+v, r=%+v", w, r)

			proxy := &httputil.ReverseProxy{
				Director: func(req *http.Request) {
					u := os.Getenv("MINIO_ENDPOINT")
					// TODO: check MINIO_SECURE
					// TODO: use MINIO_SECRET, MINIO_ID, ... need to sign
					// (after unwrapping??) the request before we send it
					// onwards: see https://github.com/smarty-archives/go-aws-auth (deprecated!) etc
					// TODO: need to add r.URL.Path?
					logrus.Infof("PROXY: r.URL.Path=%s", r.URL.Path)
					target, err := url.Parse(fmt.Sprintf("http://%s", u))
					if err != nil {
						log.Fatal(err)
					}
					targetQuery := target.RawQuery
					req.URL.Scheme = target.Scheme
					req.URL.Host = target.Host
					req.URL.Path, req.URL.RawPath = joinURLPath(target, req.URL)
					if targetQuery == "" || req.URL.RawQuery == "" {
						req.URL.RawQuery = targetQuery + req.URL.RawQuery
					} else {
						req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
					}
					req.URL.Path = mashup(req.URL.Path)
					req.URL.RawPath = mashup(req.URL.RawPath)
					req.URL.RawQuery = mashup(req.URL.RawQuery)

					// mashup each of the request headers
					for k, v := range req.Header {
						for i, vv := range v {
							v[i] = mashup(vv)
						}
						req.Header[k] = v
					}

					if _, ok := req.Header["User-Agent"]; !ok {
						// explicitly disable User-Agent so it's not set to default value ???
						req.Header.Set("User-Agent", "")
					}
					// // If calling the istio ingress, we need to set the endpoint host in the header
					// if coords.Host != "" {
					// 	// req.Header.Set("Host", coords.Host)
					// 	req.Host = coords.Host
					// }

					if req.Body != nil {
						bodyBytes, err := ioutil.ReadAll(req.Body)
						if err != nil {
							logrus.Fatal(err)
						}
						// TODO find and replace
						mashed := mashup(string(bodyBytes))
						modifiedBodyBytes := new(bytes.Buffer)
						modifiedBodyBytes.WriteString(mashed)
						req.Body = ioutil.NopCloser(modifiedBodyBytes)
						req.ContentLength = int64(len(mashed))
						// prev := req.Header.Get("Content-Length")
						// req.Header.Set("Content-Length", fmt.Sprintf("%d", len(mashed)))
						// logrus.Infof("PROXY request switching Content-Length from %d to %d", prev, len(mashed))
					}
					// TODO: check whether awsauth.Sign4 reads the whole request
					// body into memory, if it does that's bad for large writes
					// and we should figure out how we can stream it to disk
					// first...
					awsauth.Sign4(req, awsauth.Credentials{
						AccessKeyID:     os.Getenv("MINIO_ID"),
						SecretAccessKey: os.Getenv("MINIO_SECRET"),
					})
				},
				ModifyResponse: func(resp *http.Response) error {
					// TODO: skip loading the response body into memory if we're
					// streaming data! i.e. anything other than xml... maybe we
					// can use the Content-Type to distinguish..? Or only
					// PUT/GETs with certain headers and certain patterns?
					bodyBytes, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						logrus.Info(err)
						return err
					}

					// mashup each of the response headers
					for k, v := range resp.Header {
						for i, vv := range v {
							v[i] = mashup(vv)
						}
						resp.Header[k] = v
					}

					// find and replace
					if resp.Body != nil {
						mashed := unmashup(string(bodyBytes))
						modifiedBodyBytes := new(bytes.Buffer)
						modifiedBodyBytes.WriteString(mashed)
						resp.Body = ioutil.NopCloser(modifiedBodyBytes)
						resp.ContentLength = int64(len(mashed))

						// prev := resp.Header.Get("Content-Length")
						// resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(mashed)))
						// logrus.Infof("PROXY response switching Content-Length from %d to %d", prev, len(mashed))
					}
					return nil
				},
				ErrorHandler: func(resp http.ResponseWriter, r *http.Request, err error) {
					panic(err)
					log.Printf("Error from proxy connection for %s %s: %s", r.Host, r.URL.Path, err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
				},
			}

			proxy.ServeHTTP(w, r)
		},
	)
	// log where we're listening
	logrus.Infof("PROXY: Listening on port %d", port)
	srv := &http.Server{
		Handler: proxyRouter,
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	return srv.ListenAndServe()
}

// TODO: factor this out into a separate file, at least...
// from https://gist.github.com/thezelus/d5ac9ec563b061c514dc ...
// func handler(target string, p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		r.URL.Host = url.Host
// 		r.URL.Scheme = url.Scheme
// 		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
// 		r.Host = url.Host

// 		r.URL.Path = mux.Vars(r)["rest"]
// 		p.ServeHTTP(w, r)
// 	}
// }

// tf we need all this cruft
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}
