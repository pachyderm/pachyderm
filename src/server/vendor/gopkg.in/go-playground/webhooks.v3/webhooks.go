package webhooks

import (
	"fmt"
	"net/http"
)

// Header provides http.Header to minimize imports
type Header http.Header

// Provider defines the type of webhook
type Provider int

func (p Provider) String() string {
	switch p {
	case GitHub:
		return "GitHub"
	case Bitbucket:
		return "Bitbucket"
	case GitLab:
		return "GitLab"
	default:
		return "Unknown"
	}
}

// webhooks available providers
const (
	GitHub Provider = iota
	Bitbucket
	GitLab
)

// Webhook interface defines a webhook to receive events
type Webhook interface {
	Provider() Provider
	ParsePayload(w http.ResponseWriter, r *http.Request)
}

type server struct {
	hook             Webhook
	path             string
	includePathCheck bool
}

// ProcessPayloadFunc is a common function for payload return values
type ProcessPayloadFunc func(payload interface{}, header Header)

// Handler returns the webhook http.Handler for use in your own Mux implementation
func Handler(hook Webhook) http.Handler {
	return &server{
		hook: hook,
	}
}

// Run runs a server
func Run(hook Webhook, addr string, path string) error {
	srv := &server{
		hook:             hook,
		path:             path,
		includePathCheck: true,
	}
	s := &http.Server{Addr: addr, Handler: srv}

	DefaultLog.Info(fmt.Sprintf("Listening on addr: %s path: %s", addr, path))
	return s.ListenAndServe()
}

// RunServer runs a custom server.
func RunServer(s *http.Server, hook Webhook, path string) error {

	srv := &server{
		hook:             hook,
		path:             path,
		includePathCheck: true,
	}

	s.Handler = srv
	DefaultLog.Info(fmt.Sprintf("Listening on addr: %s path: %s", s.Addr, path))
	return s.ListenAndServe()
}

// RunTLSServer runs a custom server with TLS configuration.
// NOTE: http.Server Handler will be overridden by this library, just set it to nil.
// Setting the Certificates can be done in the http.Server.TLSConfig.Certificates
// see example here: https://github.com/go-playground/webhooks/blob/v2/webhooks_test.go#L178
func RunTLSServer(s *http.Server, hook Webhook, path string) error {

	srv := &server{
		hook:             hook,
		path:             path,
		includePathCheck: true,
	}

	s.Handler = srv
	DefaultLog.Info(fmt.Sprintf("Listening on addr: %s path: %s", s.Addr, path))
	return s.ListenAndServeTLS("", "")
}

// ServeHTTP is the Handler for every posted WebHook Event
func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	DefaultLog.Info("Webhook received")

	if r.Method != "POST" {
		DefaultLog.Error(fmt.Sprintf("405 Method not allowed, attempt made using Method: %s", r.Method))
		http.Error(w, "405 Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	DefaultLog.Debug(fmt.Sprintf("Include path check: %t", s.includePathCheck))
	if s.includePathCheck {
		if r.URL.Path != s.path {
			DefaultLog.Error(fmt.Sprintf("404 Not found, POST made using path: %s, but expected %s", r.URL.Path, s.path))
			http.Error(w, "404 Not found", http.StatusNotFound)
			return
		}
	}

	s.hook.ParsePayload(w, r)
}
