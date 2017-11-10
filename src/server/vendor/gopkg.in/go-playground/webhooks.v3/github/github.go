package github

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"gopkg.in/go-playground/webhooks.v3"
)

// Webhook instance contains all methods needed to process events
type Webhook struct {
	provider   webhooks.Provider
	secret     string
	eventFuncs map[Event]webhooks.ProcessPayloadFunc
}

// Config defines the configuration to create a new GitHub Webhook instance
type Config struct {
	Secret string
}

// Event defines a GitHub hook event type
type Event string

// GitHub hook types
const (
	CommitCommentEvent            Event = "commit_comment"
	CreateEvent                   Event = "create"
	DeleteEvent                   Event = "delete"
	DeploymentEvent               Event = "deployment"
	DeploymentStatusEvent         Event = "deployment_status"
	ForkEvent                     Event = "fork"
	GollumEvent                   Event = "gollum"
	IssueCommentEvent             Event = "issue_comment"
	IssuesEvent                   Event = "issues"
	LabelEvent                    Event = "label"
	MemberEvent                   Event = "member"
	MembershipEvent               Event = "membership"
	MilestoneEvent                Event = "milestone"
	OrganizationEvent             Event = "organization"
	OrgBlockEvent                 Event = "org_block"
	PageBuildEvent                Event = "page_build"
	ProjectCardEvent              Event = "project_card"
	ProjectColumnEvent            Event = "project_column"
	ProjectEvent                  Event = "project"
	PublicEvent                   Event = "public"
	PullRequestEvent              Event = "pull_request"
	PullRequestReviewEvent        Event = "pull_request_review"
	PullRequestReviewCommentEvent Event = "pull_request_review_comment"
	PushEvent                     Event = "push"
	ReleaseEvent                  Event = "release"
	RepositoryEvent               Event = "repository"
	StatusEvent                   Event = "status"
	TeamEvent                     Event = "team"
	TeamAddEvent                  Event = "team_add"
	WatchEvent                    Event = "watch"
)

// EventSubtype defines a GitHub Hook Event subtype
type EventSubtype string

// GitHub hook event subtypes
const (
	NoSubtype     EventSubtype = ""
	BranchSubtype EventSubtype = "branch"
	TagSubtype    EventSubtype = "tag"
	PullSubtype   EventSubtype = "pull"
	IssueSubtype  EventSubtype = "issues"
)

// New creates and returns a WebHook instance denoted by the Provider type
func New(config *Config) *Webhook {
	return &Webhook{
		provider:   webhooks.GitHub,
		secret:     config.Secret,
		eventFuncs: map[Event]webhooks.ProcessPayloadFunc{},
	}
}

// Provider returns the current hooks provider ID
func (hook Webhook) Provider() webhooks.Provider {
	return hook.provider
}

// RegisterEvents registers the function to call when the specified event(s) are encountered
func (hook Webhook) RegisterEvents(fn webhooks.ProcessPayloadFunc, events ...Event) {

	for _, event := range events {
		hook.eventFuncs[event] = fn
	}
}

// ParsePayload parses and verifies the payload and fires off the mapped function, if it exists.
func (hook Webhook) ParsePayload(w http.ResponseWriter, r *http.Request) {
	webhooks.DefaultLog.Info("Parsing Payload...")

	event := r.Header.Get("X-GitHub-Event")
	if len(event) == 0 {
		webhooks.DefaultLog.Error("Missing X-GitHub-Event Header")
		http.Error(w, "400 Bad Request - Missing X-GitHub-Event Header", http.StatusBadRequest)
		return
	}
	webhooks.DefaultLog.Debug(fmt.Sprintf("X-GitHub-Event:%s", event))

	gitHubEvent := Event(event)

	fn, ok := hook.eventFuncs[gitHubEvent]
	// if no event registered
	if !ok {
		webhooks.DefaultLog.Info(fmt.Sprintf("Webhook Event %s not registered, it is recommended to setup only events in github that will be registered in the webhook to avoid unnecessary traffic and reduce potential attack vectors.", event))
		return
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil || len(payload) == 0 {
		webhooks.DefaultLog.Error("Issue reading Payload")
		http.Error(w, "Issue reading Payload", http.StatusInternalServerError)
		return
	}
	webhooks.DefaultLog.Debug(fmt.Sprintf("Payload:%s", string(payload)))

	// If we have a Secret set, we should check the MAC
	if len(hook.secret) > 0 {
		webhooks.DefaultLog.Info("Checking secret")
		signature := r.Header.Get("X-Hub-Signature")
		if len(signature) == 0 {
			webhooks.DefaultLog.Error("Missing X-Hub-Signature required for HMAC verification")
			http.Error(w, "403 Forbidden - Missing X-Hub-Signature required for HMAC verification", http.StatusForbidden)
			return
		}
		webhooks.DefaultLog.Debug(fmt.Sprintf("X-Hub-Signature:%s", signature))

		mac := hmac.New(sha1.New, []byte(hook.secret))
		mac.Write(payload)

		expectedMAC := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(signature[5:]), []byte(expectedMAC)) {
			webhooks.DefaultLog.Error("HMAC verification failed")
			http.Error(w, "403 Forbidden - HMAC verification failed", http.StatusForbidden)
			return
		}
	}

	// Make headers available to ProcessPayloadFunc as a webhooks type
	hd := webhooks.Header(r.Header)

	switch gitHubEvent {
	case CommitCommentEvent:
		var cc CommitCommentPayload
		json.Unmarshal([]byte(payload), &cc)
		hook.runProcessPayloadFunc(fn, cc, hd)
	case CreateEvent:
		var c CreatePayload
		json.Unmarshal([]byte(payload), &c)
		hook.runProcessPayloadFunc(fn, c, hd)
	case DeleteEvent:
		var d DeletePayload
		json.Unmarshal([]byte(payload), &d)
		hook.runProcessPayloadFunc(fn, d, hd)
	case DeploymentEvent:
		var d DeploymentPayload
		json.Unmarshal([]byte(payload), &d)
		hook.runProcessPayloadFunc(fn, d, hd)
	case DeploymentStatusEvent:
		var d DeploymentStatusPayload
		json.Unmarshal([]byte(payload), &d)
		hook.runProcessPayloadFunc(fn, d, hd)
	case ForkEvent:
		var f ForkPayload
		json.Unmarshal([]byte(payload), &f)
		hook.runProcessPayloadFunc(fn, f, hd)
	case GollumEvent:
		var g GollumPayload
		json.Unmarshal([]byte(payload), &g)
		hook.runProcessPayloadFunc(fn, g, hd)
	case IssueCommentEvent:
		var i IssueCommentPayload
		json.Unmarshal([]byte(payload), &i)
		hook.runProcessPayloadFunc(fn, i, hd)
	case IssuesEvent:
		var i IssuesPayload
		json.Unmarshal([]byte(payload), &i)
		hook.runProcessPayloadFunc(fn, i, hd)
	case LabelEvent:
		var l LabelPayload
		json.Unmarshal([]byte(payload), &l)
		hook.runProcessPayloadFunc(fn, l, hd)
	case MemberEvent:
		var m MemberPayload
		json.Unmarshal([]byte(payload), &m)
		hook.runProcessPayloadFunc(fn, m, hd)
	case MembershipEvent:
		var m MembershipPayload
		json.Unmarshal([]byte(payload), &m)
		hook.runProcessPayloadFunc(fn, m, hd)
	case MilestoneEvent:
		var m MilestonePayload
		json.Unmarshal([]byte(payload), &m)
		hook.runProcessPayloadFunc(fn, m, hd)
	case OrganizationEvent:
		var o OrganizationPayload
		json.Unmarshal([]byte(payload), &o)
		hook.runProcessPayloadFunc(fn, o, hd)
	case OrgBlockEvent:
		var o OrgBlockPayload
		json.Unmarshal([]byte(payload), &o)
		hook.runProcessPayloadFunc(fn, o, hd)
	case PageBuildEvent:
		var p PageBuildPayload
		json.Unmarshal([]byte(payload), &p)
		hook.runProcessPayloadFunc(fn, p, hd)
	case ProjectCardEvent:
		var p ProjectCardPayload
		json.Unmarshal([]byte(payload), &p)
		hook.runProcessPayloadFunc(fn, p, hd)
	case ProjectColumnEvent:
		var p ProjectColumnPayload
		json.Unmarshal([]byte(payload), &p)
		hook.runProcessPayloadFunc(fn, p, hd)
	case ProjectEvent:
		var p ProjectPayload
		json.Unmarshal([]byte(payload), &p)
		hook.runProcessPayloadFunc(fn, p, hd)
	case PublicEvent:
		var p PublicPayload
		json.Unmarshal([]byte(payload), &p)
		hook.runProcessPayloadFunc(fn, p, hd)
	case PullRequestEvent:
		var p PullRequestPayload
		json.Unmarshal([]byte(payload), &p)
		hook.runProcessPayloadFunc(fn, p, hd)
	case PullRequestReviewEvent:
		var p PullRequestReviewPayload
		json.Unmarshal([]byte(payload), &p)
		hook.runProcessPayloadFunc(fn, p, hd)
	case PullRequestReviewCommentEvent:
		var p PullRequestReviewCommentPayload
		json.Unmarshal([]byte(payload), &p)
		hook.runProcessPayloadFunc(fn, p, hd)
	case PushEvent:
		var p PushPayload
		json.Unmarshal([]byte(payload), &p)
		hook.runProcessPayloadFunc(fn, p, hd)
	case ReleaseEvent:
		var r ReleasePayload
		json.Unmarshal([]byte(payload), &r)
		hook.runProcessPayloadFunc(fn, r, hd)
	case RepositoryEvent:
		var r RepositoryPayload
		json.Unmarshal([]byte(payload), &r)
		hook.runProcessPayloadFunc(fn, r, hd)
	case StatusEvent:
		var s StatusPayload
		json.Unmarshal([]byte(payload), &s)
		hook.runProcessPayloadFunc(fn, s, hd)
	case TeamEvent:
		var t TeamPayload
		json.Unmarshal([]byte(payload), &t)
		hook.runProcessPayloadFunc(fn, t, hd)
	case TeamAddEvent:
		var t TeamAddPayload
		json.Unmarshal([]byte(payload), &t)
		hook.runProcessPayloadFunc(fn, t, hd)
	case WatchEvent:
		var w WatchPayload
		json.Unmarshal([]byte(payload), &w)
		hook.runProcessPayloadFunc(fn, w, hd)
	}
}

func (hook Webhook) runProcessPayloadFunc(fn webhooks.ProcessPayloadFunc, results interface{}, header webhooks.Header) {
	go func(fn webhooks.ProcessPayloadFunc, results interface{}, header webhooks.Header) {
		fn(results, header)
	}(fn, results, header)
}
