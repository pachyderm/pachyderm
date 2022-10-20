// A proxy to the real S3 server underlying pachyderm, for cases when stronger
// S3 compatibility & performance is required for "s3_out" feature.

package server

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/sirupsen/logrus"
	awsauth "github.com/smartystreets/go-aws-auth"
)

type RawS3Proxy struct {
}

// TODO: These shouldn't be global variables, it's just expedient for this
// experimental feature.

// CurrentBucket is the real backend bucket. Set in s3g_sidecar.go
var CurrentBucket string = "out"

// This is like pipeline_name-<job-id>
var CurrentTargetPath string = ""

// Autodetected based on proxy traffic, will be like "example-data-24" or
// whatever path the user specified after "s3a://out/<path>"
var LastSeenPathToReplace string = ""
var CurrentJobInfo *pps.JobInfo
var CurrentPachClient *client.APIClient

func (r *RawS3Proxy) ListenAndServe(port uint16) error {

	// Note: os.GetEnv("STORAGE_BACKEND") ==> "MINIO" or, presumably, "AWS"...
	// MINIO_SECRET, MINIO_ID, MINIO_ENDPOINT (e.g.
	// minio.default.svc.cluster.local:9000), MINIO_SECURE, MINIO_BUCKET also
	// set.

	if os.Getenv("STORAGE_BACKEND") != "MINIO" || os.Getenv("STORAGE_BACKEND") != "AMAZON" {
		panic("only MINIO or AMAZON STORAGE_BACKEND supported by proxy to real backend s3_out feature right now")
	}

	proxyRouter := mux.NewRouter()
	proxyRouter.PathPrefix("/finish").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Worker is telling us the job is finished. If it's a raw_s3_out
			// job, upload any resulting data written to the job path in the
			// backend bucket recursively into the s3_out bucket

			// We do need to be careful to be idempotent! We are somewhat, in
			// that we still call RemoveRouter and don't block this function
			// even if the recursive copy doesn't succeed (e.g. the location
			// we're copying from has already been deleted).
			outputCommit := CurrentJobInfo.OutputCommit
			logrus.Infof(
				"PROXY Uploading result of job %s from job-scoped bucket path to output commit %s",
				CurrentJobInfo.Job.ID, outputCommit,
			)
			var url string
			if os.Getenv("STORAGE_BACKEND") == "MINIO" {
				url = fmt.Sprintf("test-minio://minio:9000/%s/%s", CurrentBucket, CurrentTargetPath)
			} else if os.Getenv("STORAGE_BACKEND") == "AMAZON" {
				url = fmt.Sprintf("s3://%s/%s", CurrentBucket, CurrentTargetPath)
			}
			logrus.Infof("PROXY Starting PutFileURL to / in output commit from %s", url)

			err := CurrentPachClient.WithModifyFileClient(outputCommit, func(mf client.ModifyFile) error {
				// See src/internal/obj/factory.go NewClientFromURLAndSecret
				err := mf.PutFileURL("/", url, true)
				if err != nil {
					logrus.Infof("PROXY ERROR while copying %s: %s", os.Getenv("STORAGE_BACKEND"), err)
					return err
				}
				return nil
			})
			if err != nil {
				// TODO: maybe we want to retry a certain number of times?
				logrus.Infof("PROXY error copying, skipping cleanup: %s", err)
				return
			}
			logrus.Infof("PROXY finished copying!")
			// Clean up the backend bucket!
			err = func() error {
				u, err := obj.ParseURL(url)
				if err != nil {
					return errors.Wrapf(err, "error parsing url %v", url)
				}
				objClient, err := obj.NewClientFromURLAndSecret(u, false)
				if err != nil {
					return err
				}
				logrus.Infof("PROXY starting cleanup!")
				err = objClient.Walk(context.Background(), CurrentTargetPath, func(path string) error {
					logrus.Infof("PROXY Deleting %s", path)
					return objClient.Delete(context.Background(), path)
				})
				return err
			}()
			if err != nil {
				logrus.Infof("PROXY failed cleanup with %s, continuing", err)
			} else {
				logrus.Infof("PROXY finished cleanup!")
			}
		},
	)

	proxyRouter.PathPrefix("/").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			if LastSeenPathToReplace == "" {
				// match the first Spark request per job...
				// HEAD /out/example-data-24 HTTP/1.1
				logrus.Debugf("PROXY r.Method='%s', r.RequestURI='%s'", r.Method, r.RequestURI)
				if r.Method == "HEAD" && strings.HasPrefix(r.RequestURI, "/out/") {
					LastSeenPathToReplace = r.RequestURI[len("/out/"):]
					logrus.Debugf("PROXY LastSeenPathToReplace: '%s'", LastSeenPathToReplace)
				}
			}

			transform := func(s string) string {
				// TODO: think about whether this might transform too much in some cases
				initial := s
				ret := strings.Replace(s, "/out", "/"+CurrentBucket, -1)
				// XML stylee as well
				ret = strings.Replace(ret, ">out", ">"+CurrentBucket, -1)
				if LastSeenPathToReplace != "" {
					// XXX SECURITY: Think about how inferring
					// LastSeenPathToReplace based on user generated traffic may
					// allow access to the whole bucket
					ret = strings.Replace(ret, LastSeenPathToReplace, CurrentTargetPath+"/"+LastSeenPathToReplace, -1)
				} else {
					if len(s) <= 100 {
						logrus.Debugf("PROXY Warning! LastSeenPathToReplace was empty when transforming up '%s'", s)
					}
				}
				// logrus.Debugf("transform '%s' ==> '%s'", s, ret)
				if initial != ret {
					logrus.Debugf("PROXY transformed '%s' to '%s'", initial, ret)
				}
				return ret
			}

			untransform := func(s string) string {
				// TODO: think about whether this might transform too much in some cases
				initial := s
				ret := strings.Replace(s, "/"+CurrentBucket, "/out", -1)
				// XML stylee as well
				ret = strings.Replace(ret, ">"+CurrentBucket, ">out", -1)
				if LastSeenPathToReplace != "" {
					// XXX SECURITY: Think about how inferring
					// LastSeenPathToReplace based on user generated traffic may
					// allow access to the whole bucket
					ret = strings.Replace(ret, CurrentTargetPath+"/"+LastSeenPathToReplace, LastSeenPathToReplace, -1)
				} else {
					if len(s) <= 100 {
						logrus.Debugf("PROXY Warning! LastSeenPathToReplace was empty when untransforming '%s'", s)
					}
				}
				if initial != ret {
					logrus.Debugf("PROXY reverse transformed '%s' to '%s'", initial, ret)
				}
				return ret
			}

			proxy := &httputil.ReverseProxy{
				Director: func(req *http.Request) {
					var u string
					if os.Getenv("STORAGE_BACKEND") == "MINIO" {
						u = fmt.Sprintf("http://%s", os.Getenv("MINIO_ENDPOINT"))
					} else if os.Getenv("STORAGE_BACKEND") == "AMAZON" {
						// https://docs.aws.amazon.com/AmazonS3/latest/userguide/access-bucket-intro.html
						u = fmt.Sprintf("https://s3.%s.amazonaws.com", os.Getenv("AWS_REGION"))
					} else {
						panic("only MINIO or AMAZON STORAGE_BACKEND supported by proxy to real backend s3_out feature right now")
					}
					logrus.Debugf("PROXY: r.URL.Path=%s", r.URL.Path)
					target, err := url.Parse(u)
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
					req.URL.Path = transform(req.URL.Path)
					req.URL.RawPath = transform(req.URL.RawPath)
					req.URL.RawQuery = transform(req.URL.RawQuery)

					// transform each of the request headers
					for k, v := range req.Header {
						for i, vv := range v {
							v[i] = transform(vv)
						}
						req.Header[k] = v
					}

					if req.Body != nil && req.Header.Get("Content-Type") != "application/octet-stream" {
						bodyBytes, err := ioutil.ReadAll(req.Body)
						if err != nil {
							logrus.Fatal(err)
						}
						transformed := transform(string(bodyBytes))
						modifiedBodyBytes := new(bytes.Buffer)
						modifiedBodyBytes.WriteString(transformed)
						req.Body = ioutil.NopCloser(modifiedBodyBytes)
						req.ContentLength = int64(len(transformed))
					}
					// TODO: check whether awsauth.Sign4 reads the whole request
					// body into memory, if it does that's bad for large writes
					// and we should figure out how we can stream it to disk
					// first...
					if os.Getenv("STORAGE_BACKEND") == "AMAZON" {
						awsauth.Sign4(req)
						// Maybe awsauth.Sign4 is magic enough to know to look
						// in all the right places as an aws client? Hopefully.
						// If not, we can do something like this:
						/*
							AmazonRegionEnvVar       = "AMAZON_REGION"
							AmazonBucketEnvVar       = "AMAZON_BUCKET"
							AmazonIDEnvVar           = "AMAZON_ID"
							AmazonSecretEnvVar       = "AMAZON_SECRET"
							AmazonTokenEnvVar        = "AMAZON_TOKEN"
							AmazonDistributionEnvVar = "AMAZON_DISTRIBUTION"
							CustomEndpointEnvVar     = "CUSTOM_ENDPOINT"
						*/
						// awsauth.Sign4(req, awsauth.Credentials{
						// 	AccessKeyID:     os.Getenv("AMAZON_ID"),
						// 	SecretAccessKey: os.Getenv("AMAZON_SECRET"),
						// 	SessionToken:    os.Getenv("AMAZON_TOKEN"),
						// }
						// Copilot suggested this too, dunno if it's useful...
						// ..., awsauth.SigningOptions{
						// 	   Region: os.Getenv(AmazonRegionEnvVar),
						// 	   Service: "s3",
						//     Expires: time.Now().Add(15 * time.Minute),
						// 	   Headers: []string{"host", "x-amz-date"},
						// })

					} else if os.Getenv("STORAGE_BACKEND") == "MINIO" {
						awsauth.Sign4(req, awsauth.Credentials{
							AccessKeyID:     os.Getenv("MINIO_ID"),
							SecretAccessKey: os.Getenv("MINIO_SECRET"),
						})
					}
				},
				ModifyResponse: func(resp *http.Response) error {
					// transform each of the response headers
					for k, v := range resp.Header {
						for i, vv := range v {
							v[i] = transform(vv)
						}
						resp.Header[k] = v
					}

					// find and replace - only if not a byte stream (large data!)
					if resp.Body != nil && resp.Header.Get("Content-Type") != "application/octet-stream" {
						bodyBytes, err := ioutil.ReadAll(resp.Body)
						if err != nil {
							return err
						}

						transformed := untransform(string(bodyBytes))
						modifiedBodyBytes := new(bytes.Buffer)
						modifiedBodyBytes.WriteString(transformed)
						prev := resp.ContentLength

						resp.Body = ioutil.NopCloser(modifiedBodyBytes)
						resp.ContentLength = int64(len(transformed))
						logrus.Debugf("PROXY Setting content length from %d to %d", prev, resp.ContentLength)

						// not sure why we need to do this as well as setting
						// resp.ContentLength, but we do (only in the response
						// case seemingly)
						resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(transformed)))
					}
					return nil
				},
				ErrorHandler: func(resp http.ResponseWriter, r *http.Request, err error) {
					log.Printf("Error from proxy connection for %s %s: %s, stack=%s", r.Host, r.URL.Path, err, debug.Stack())
					http.Error(w, err.Error(), http.StatusInternalServerError)
				},
			}

			proxy.ServeHTTP(w, r)
		},
	)
	// log where we're listening
	logrus.Debugf("PROXY: Listening on port %d", port)
	srv := &http.Server{
		Handler: proxyRouter,
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 600 * time.Second,
		ReadTimeout:  600 * time.Second,
	}
	return srv.ListenAndServe()
}

// copied from standard library
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
