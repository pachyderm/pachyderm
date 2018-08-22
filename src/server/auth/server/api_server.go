package server

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc/metadata"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/crewjam/saml"
	"github.com/gogo/protobuf/types"
	"github.com/google/go-github/github"
	logrus "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/src/client"
	authclient "github.com/pachyderm/pachyderm/src/client/auth"
	enterpriseclient "github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsconsts"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"
)

const (
	// DisableAuthenticationEnvVar specifies an environment variable that, if set, causes
	// Pachyderm authentication to ignore github and authmatically generate a
	// pachyderm token for any username in the AuthenticateRequest.GitHubToken field
	DisableAuthenticationEnvVar = "PACHYDERM_AUTHENTICATION_DISABLED_FOR_TESTING"

	tokensPrefix              = "/tokens"
	authenticationCodesPrefix = "/auth-codes"
	aclsPrefix                = "/acls"
	adminsPrefix              = "/admins"
	membersPrefix             = "/members"
	groupsPrefix              = "/groups"
	configPrefix              = "/config"

	defaultTokenTTLSecs = 30 * 24 * 60 * 60 // 30 days

	// defaultAuthCodeTTLSecs is the lifetime of an Authentication Code from
	// GetAuthenticationCode
	defaultAuthCodeTTLSecs = 60

	// magicUser is a special, unrevokable cluster administrator. It's not
	// possible to log in as magicUser, but pipelines with no owner are run as
	// magicUser when auth is activated. This string is not secret, but is long
	// and random to avoid collisions with real usernames
	magicUser = `magic:GZD4jKDGcirJyWQt6HtK4hhRD6faOofP1mng34xNZsI`

	// configKey is a key (in etcd, in the config collection) that maps to the
	// auth configuration. This is the only key in that collection (due to
	// implemenation details of our config library, we can't use an empty key)
	configKey = "x"

	// SamlPort is the port where SAML ID Providers can send auth assertions
	SamlPort = 654
)

// githubTokenRegex is used when pachd is deployed in "dev mode" (i.e. when
// pachd is deployed with "pachctl deploy local") to guess whether a call to
// Authenticate or Authorize contains a real GitHub access token.
//
// If the field GitHubToken matches this regex, it's assumed to be a GitHub
// token and pachd retrieves the corresponding user's GitHub username. If not,
// pachd automatically authenticates the the caller as the GitHub user whose
// username is the string in "GitHubToken".
var githubTokenRegex = regexp.MustCompile("^[0-9a-f]{40}$")

// epsilon is small, nonempty protobuf to use as an etcd value (the etcd client
// library can't distinguish between empty values and missing values, even
// though empty values are still stored in etcd)
var epsilon = &types.BoolValue{Value: true}

type apiServer struct {
	pachLogger log.Logger
	etcdClient *etcd.Client

	address        string            // address of a Pachd server
	pachClient     *client.APIClient // pachd client
	pachClientOnce sync.Once         // used to initialize pachClient

	adminCache map[string]struct{} // cache of current cluster admins
	adminMu    sync.Mutex          // guard 'adminCache'

	configCache authclient.AuthConfig // cache of auth config in etcd
	configMu    sync.Mutex            // guard 'configCache'

	samlSP          *saml.ServiceProvider // object for parsing saml responses
	redirectAddress *url.URL              // address where users will be redirected after authenticating
	samlSPMu        sync.Mutex            // guard 'samlSP'

	// tokens is a collection of hashedToken -> TokenInfo mappings. These tokens are
	// returned to users by Authenticate()
	tokens col.Collection
	// authenticationCodes is a collection of hash(code) -> TokenInfo mappings.
	// These codes are generated internally, and converted to regular tokens by
	// Authenticate()
	authenticationCodes col.Collection
	// acls is a collection of repoName -> ACL mappings.
	acls col.Collection
	// admins is a collection of username -> Empty mappings (keys indicate which
	// github users are cluster admins)
	admins col.Collection
	// members is a collection of username -> groups mappings.
	members col.Collection
	// groups is a collection of group -> usernames mappings.
	groups col.Collection
	// collection containing the auth config (under the key configKey)
	authConfig col.Collection

	// This is a cache of the PPS master token. It's set once on startup and then
	// never updated
	ppsToken string

	// public addresses the fact that pachd in full mode initializes two auth
	// servers: one that exposes a public API, possibly over TLS, and one that
	// exposes a private API, for internal services. Only the public-facing auth
	// service should export the SAML ACS and Metadata services, so if public
	// is true and auth is active, this may export those SAML services
	public bool
}

// LogReq is like log.Logger.Log(), but it assumes that it's being called from
// the top level of a GRPC method implementation, and correspondingly extracts
// the method name from the parent stack frame
func (a *apiServer) LogReq(request interface{}) {
	a.pachLogger.Log(request, nil, nil, 0)
}

// LogResp is like log.Logger.Log(). However,
// 1) It assumes that it's being called from a defer() statement in a GRPC
//    method , and correspondingly extracts the method name from the grandparent
//    stack frame
// 2) It logs NotActivatedError at DebugLevel instead of ErrorLevel, as, in most
//    cases, this error is expected, and logging it frequently may confuse users
func (a *apiServer) LogResp(request interface{}, response interface{}, err error, duration time.Duration) {
	if err == nil {
		a.pachLogger.LogAtLevelFromDepth(request, response, err, duration, logrus.InfoLevel, 4)
	} else if authclient.IsErrNotActivated(err) {
		a.pachLogger.LogAtLevelFromDepth(request, response, err, duration, logrus.DebugLevel, 4)
	} else {
		a.pachLogger.LogAtLevelFromDepth(request, response, err, duration, logrus.ErrorLevel, 4)
	}
}

func (a *apiServer) getPachClient() *client.APIClient {
	a.pachClientOnce.Do(func() {
		var err error
		a.pachClient, err = client.NewFromAddress(a.address)
		if err != nil {
			panic(err)
		}
	})
	return a.pachClient
}

// NewAuthServer returns an implementation of authclient.APIServer.
func NewAuthServer(pachdAddress string, etcdAddress string, etcdPrefix string, public bool) (authclient.APIServer, error) {
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{etcdAddress},
		DialOptions: client.DefaultDialOptions(),
	})
	if err != nil {
		return nil, fmt.Errorf("error constructing etcdClient: %v", err)
	}

	s := &apiServer{
		pachLogger: log.NewLogger("authclient.API"),
		etcdClient: etcdClient,
		address:    pachdAddress,
		adminCache: make(map[string]struct{}),
		tokens: col.NewCollection(
			etcdClient,
			path.Join(etcdPrefix, tokensPrefix),
			nil,
			&authclient.TokenInfo{},
			nil,
			nil,
		),
		authenticationCodes: col.NewCollection(
			etcdClient,
			path.Join(etcdPrefix, authenticationCodesPrefix),
			nil,
			&authclient.TokenInfo{},
			nil,
			nil,
		),
		acls: col.NewCollection(
			etcdClient,
			path.Join(etcdPrefix, aclsPrefix),
			nil,
			&authclient.ACL{},
			nil,
			nil,
		),
		admins: col.NewCollection(
			etcdClient,
			path.Join(etcdPrefix, adminsPrefix),
			nil,
			&types.BoolValue{}, // smallest value that etcd actually stores
			nil,
			nil,
		),
		members: col.NewCollection(
			etcdClient,
			path.Join(etcdPrefix, membersPrefix),
			nil,
			&authclient.Groups{},
			nil,
			nil,
		),
		groups: col.NewCollection(
			etcdClient,
			path.Join(etcdPrefix, groupsPrefix),
			nil,
			&authclient.Users{},
			nil,
			nil,
		),
		authConfig: col.NewCollection(
			etcdClient,
			path.Join(etcdPrefix, configKey),
			nil,
			&authclient.AuthConfig{},
			nil,
			nil,
		),
		public: public,
	}
	go s.retrieveOrGeneratePPSToken()
	go s.getPachClient() // initialize connection to Pachd
	go s.watchAdmins(path.Join(etcdPrefix, adminsPrefix))

	// Watch config for new SAML options, and start SAML service (won't respond to
	// anything until config is set)
	go s.watchConfig()
	go s.serveSAML()
	return s, nil
}

type activationState int

const (
	none activationState = iota
	partial
	full
)

// activationState returns 'none' if auth is totally inactive, 'partial' if
// auth.Activate has been called, but hasn't finished or failed, and full
// if auth.Activate has been called and succeeded.
//
// When the activation state is 'partial' users cannot authenticate; the only
// functioning auth API calls are Activate() (to retry activation) and
// Deactivate() (to give up and revert to the 'none' state)
func (a *apiServer) activationState() activationState {
	a.adminMu.Lock()
	defer a.adminMu.Unlock()
	if len(a.adminCache) == 0 {
		return none
	}
	if _, magicUserIsAdmin := a.adminCache[magicUser]; magicUserIsAdmin {
		return partial
	}
	return full
}

// type statReadCloser struct {
// 	io.ReadCloser
// 	bytesRead int
// 	data      bytes.Buffer
// }
//
// func (s *statReadCloser) Read(p []byte) (n int, err error) {
// 	n, err = s.ReadCloser.Read(p)
// 	s.bytesRead += n
// 	s.data.Write(p)
// 	return n, err
// }

// func (a *apiServer) serveSaml() {
// 	fmt.Printf(">>> About to enter a.serveSamlOnce.Do()\n")
// 	a.serveSamlOnce.Do(func() {
// 		fmt.Printf(">>> (once) Start SAML ACS\n")
// 		fmt.Printf(">>> (once) SAML ACS should now be started\n")
// 		sp := saml.ServiceProvider{
//
// 			// These need to be set from the config that I'm adding
// 			AcsURL: mustParseURL(a.configCache.SAMLServiceOptions.ACSURL),
// 			// MetadataURL: /* config value */
//
// 			Logger: logrus.New(),
//
// 			// Not set:
// 			// Key: Private key for Pachyderm ACS. Unclear if needed
// 			// Certificate: Public key for Pachyderm ACS. Unclear if needed
// 			// ForceAuthn (whether users need to re-authenticate with the IdP, even if
// 			//            they already have a session—leaving this false)
// 			// AuthnNameIDFormat (format the ACS expects the AuthnName to be in)
// 			// MetadataValidDuration (how long the SP endpoints are valid? Returned by
// 			//                       the Metadata service)
// 		}
// 		samlMux := http.NewServeMux()
// 		samlMux.HandleFunc("/saml/acs", func(w http.ResponseWriter, req *http.Request) {
// 			// stat := statReadCloser{
// 			// 	ReadCloser: req.Body,
// 			// }
// 			// req.Body = &stat
// 			out := io.MultiWriter(w, os.Stdout)
// 			possibleRequestIDs := []string{""} // only IdP-initiated auth enabled for now
// 			fmt.Printf(">>> req.PostFormValue(\"SAMLResponse\"): %s\n", req.PostFormValue("SAMLResponse"))
// 			assertion, err := sp.ParseResponse(req, possibleRequestIDs)
// 			if err != nil {
// 				w.WriteHeader(http.StatusInternalServerError)
// 				out.Write([]byte("<html><head></head><body>"))
// 				out.Write([]byte("Error parsing SAML response: "))
// 				out.Write([]byte(err.Error()))
// 				ie := err.(*saml.InvalidResponseError)
// 				out.Write([]byte("\nPrivate error: " + ie.PrivateErr.Error()))
// 				out.Write([]byte("\n"))
// 			}
// 			if assertion == nil {
// 				w.WriteHeader(http.StatusInternalServerError)
// 				out.Write([]byte("<html><head></head><body>"))
// 				out.Write([]byte("Nil assertion\n"))
// 			}
// 			out.Write([]byte("Would've authenticated as:\n"))
// 			fmt.Printf(">>> This is a test\n")
// 			if assertion != nil {
// 				out.Write([]byte(fmt.Sprintf("assertion: %v\n", assertion)))
// 				if assertion.Subject != nil {
// 					out.Write([]byte(fmt.Sprintf("assertion.Subject: %v\n", assertion.Subject)))
// 					if assertion.Subject.NameID != nil {
// 						out.Write([]byte(fmt.Sprintf("assertion.Subject.NameID.Value: %v\n", assertion.Subject.NameID.Value)))
// 					}
// 				}
// 				if assertion.Element() != nil {
// 					xmlBytes, err := xml.MarshalIndent(assertion.Element(), "", "  ")
// 					if err != nil {
// 						out.Write([]byte(fmt.Sprintf("could not marshall assertion: %v\n", err)))
// 					} else {
// 						out.Write([]byte(fmt.Sprintf("<pre>\n%s\n</pre>", xmlBytes)))
// 					}
// 				} else {
// 					out.Write([]byte("assertion.Element() was nil"))
// 				}
// 			} else {
// 				out.Write([]byte("assertion was nil"))
// 			}
// 			out.Write([]byte("</body></html>"))
// 		})
// 		samlMux.HandleFunc("/*", func(w http.ResponseWriter, req *http.Request) {
// 			fmt.Printf(">>> received request to %s\n", req.URL.Path)
// 			w.WriteHeader(http.StatusTeapot)
// 		})
// 		http.ListenAndServe(fmt.Sprintf(":%d", SamlPort), samlMux)
// 	})
// }

// Retrieve the PPS master token, or generate it and put it in etcd.
// TODO This is a hack. It avoids the need to return superuser tokens from
// GetAuthToken (essentially, PPS and Auth communicate through etcd instead of
// an API) but we should define an internal API and use that instead.
func (a *apiServer) retrieveOrGeneratePPSToken() {
	var tokenProto types.StringValue // will contain PPS token
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 60 * time.Second
	b.MaxInterval = 5 * time.Second
	if err := backoff.Retry(func() error {
		if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			superUserTokenCol := col.NewCollection(a.etcdClient, ppsconsts.PPSTokenKey, nil, &types.StringValue{}, nil, nil).ReadWrite(stm)
			// TODO(msteffen): Don't use an empty key, as it will not be erased by
			// superUserTokenCol.DeleteAll()
			err := superUserTokenCol.Get("", &tokenProto)
			if err == nil {
				return nil
			}
			if col.IsErrNotFound(err) {
				// no existing token yet -- generate token
				token := uuid.NewWithoutDashes()
				tokenProto.Value = token
				if err := superUserTokenCol.Create("", &tokenProto); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		a.ppsToken = tokenProto.Value
		return nil
	}, b); err != nil {
		panic(fmt.Sprintf("couldn't create/retrieve PPS superuser token within 60s of starting up: %v", err))
	}
}

func (a *apiServer) watchAdmins(fullAdminPrefix string) {
	b := backoff.NewExponentialBackOff()
	backoff.RetryNotify(func() error {
		// Watch for the addition/removal of new admins. Note that this will return
		// any existing admins, so if the auth service is already activated, it will
		// stay activated.
		watcher, err := a.admins.ReadOnly(context.Background()).Watch()
		if err != nil {
			return err
		}
		defer watcher.Close()
		// The auth service is activated if we have admins, and not
		// activated otherwise.
		for {
			ev, ok := <-watcher.Watch()
			if !ok {
				return errors.New("admin watch closed unexpectedly")
			}
			b.Reset() // event successfully received

			if err := func() error {
				// Lock a.adminMu in case we need to modify a.adminCache
				a.adminMu.Lock()
				defer a.adminMu.Unlock()

				// Parse event data and potentially update adminCache
				var key string
				var boolProto types.BoolValue
				ev.Unmarshal(&key, &boolProto)
				username := strings.TrimPrefix(key, fullAdminPrefix+"/")
				switch ev.Type {
				case watch.EventPut:
					a.adminCache[username] = struct{}{}
				case watch.EventDelete:
					delete(a.adminCache, username)
				case watch.EventError:
					return ev.Err
				}
				return nil // unlock mu
			}(); err != nil {
				return err
			}
		}
	}, b, func(err error, d time.Duration) error {
		logrus.Printf("error watching admin collection: %v; retrying in %v", err, d)
		return nil
	})
}

func (a *apiServer) watchConfig() {
	b := backoff.NewExponentialBackOff()
	backoff.RetryNotify(func() error {
		// Watch for the addition/removal of new admins. Note that this will return
		// any existing admins, so if the auth service is already activated, it will
		// stay activated.
		watcher, err := a.authConfig.ReadOnly(context.Background()).Watch()
		if err != nil {
			return err
		}
		defer watcher.Close()
		// Wait for new config events to arrive
		for {
			ev, ok := <-watcher.Watch()
			if !ok {
				return errors.New("admin watch closed unexpectedly")
			}
			b.Reset() // event successfully received
			dbgLog.Printf("(apiServer.watchConfig) new auth config event received\n")

			if a.activationState() != full {
				return fmt.Errorf("received config event while auth not fully " +
					"activated (should be impossible), restarting")
			}
			if err := func() error {
				dbgLog.Printf("(apiServer.watchConfig) about to update config cache\n")
				// Lock a.configMu in case we need to modify a.configCache
				a.configMu.Lock()
				defer a.configMu.Unlock()

				// Parse event data and potentially update configCache
				var key string // always configKey, just need to put it somewhere
				var configProto authclient.AuthConfig
				ev.Unmarshal(&key, &configProto)
				switch ev.Type {
				case watch.EventPut:
					a.configCache = configProto
				case watch.EventDelete:
					a.configCache.Reset()
				case watch.EventError:
					return ev.Err
				}
				dbgLog.Printf("(apiServer.watchConfig) a.configCache = %v\n", a.configCache)
				if err := a.updateSAMLSP(); err != nil {
					logrus.Warnf("could not update SAML service with new config: %v", err)
				}
				return nil // unlock mu
			}(); err != nil {
				return err
			}
		}
	}, b, func(err error, d time.Duration) error {
		logrus.Printf("error watching auth config: %v; retrying in %v", err, d)
		return nil
	})
}

func (a *apiServer) getEnterpriseTokenState() (enterpriseclient.State, error) {
	pachClient := a.getPachClient()
	resp, err := pachClient.Enterprise.GetState(context.Background(),
		&enterpriseclient.GetStateRequest{})
	if err != nil {
		return 0, fmt.Errorf("could not get Enterprise status: %v", grpcutil.ScrubGRPC(err))
	}
	return resp.State, nil
}

func (a *apiServer) Activate(ctx context.Context, req *authclient.ActivateRequest) (resp *authclient.ActivateResponse, retErr error) {
	pachClient := a.getPachClient().WithCtx(ctx)
	ctx = a.pachClient.Ctx()
	// We don't want to actually log the request/response since they contain
	// credentials.
	defer func(start time.Time) { a.LogResp(nil, nil, retErr, time.Since(start)) }(time.Now())
	// If the cluster's Pachyderm Enterprise token isn't active, the auth system
	// cannot be activated
	state, err := a.getEnterpriseTokenState()
	if err != nil {
		return nil, fmt.Errorf("error confirming Pachyderm Enterprise token: %v", err)
	}
	if state != enterpriseclient.State_ACTIVE {
		return nil, fmt.Errorf("Pachyderm Enterprise is not active in this " +
			"cluster, and the Pachyderm auth API is an Enterprise-level feature")
	}

	// Activating an already activated auth service should fail, because
	// otherwise anyone can just activate the service again and set
	// themselves as an admin. If activation failed in PFS, calling auth.Activate
	// again should work (in this state, the only admin will be 'magicUser')
	if a.activationState() == full {
		return nil, fmt.Errorf("already activated")
	}

	// The Pachyderm token that Activate() returns will have the TTL
	// - 'defaultTokenTTLSecs' if the initial admin is a GitHub user (who can get
	//   a new token by re-authenticating via GitHub after this token expires)
	// - 0 (no TTL, indefinite lifetime) if the initial admin is a robot user
	//   (who has no way to acquire a new token once this token expires)
	ttlSecs := int64(defaultTokenTTLSecs)
	// Authenticate the caller (or generate a new auth token if req.Subject is a
	// robot user)
	if req.Subject != "" {
		req.Subject, err = a.lenientCanonicalizeSubject(ctx, req.Subject)
		if err != nil {
			return nil, err
		}
	}
	switch {
	case req.Subject == "":
		fallthrough
	case strings.HasPrefix(req.Subject, authclient.GitHubPrefix):
		username, err := GitHubTokenToUsername(ctx, req.GitHubToken)
		if err != nil {
			return nil, err
		}
		if req.Subject != "" && req.Subject != username {
			return nil, fmt.Errorf("asserted subject \"%s\" did not match owner of GitHub token \"%s\"", req.Subject, username)
		}
		req.Subject = username
	case strings.HasPrefix(req.Subject, authclient.RobotPrefix):
		// req.Subject will be used verbatim, and the resulting code will
		// authenticate the holder as the robot account therein
		ttlSecs = 0 // no expiration for robot tokens -- see above
	default:
		return nil, fmt.Errorf("invalid subject in request (must be a GitHub user or robot): \"%s\"", req.Subject)
	}

	// Hack: set the cluster admins to just {magicUser}. This puts the auth system
	// in the "partial" activation state. Users cannot authenticate, but auth
	// checks are now enforced, which means no pipelines or repos can be created
	// while ACLs are being added to every repo for the existing pipelines
	if _, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		return a.admins.ReadWrite(stm).Put(magicUser, epsilon)
	}); err != nil {
		return nil, err
	}
	// wait until watchAdmins has updated the local cache (changing the activation
	// state)
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 30 * time.Second
	if err := backoff.Retry(func() error {
		if a.activationState() != partial {
			return fmt.Errorf("auth never entered \"partial\" activation state")
		}
		return nil
	}, b); err != nil {
		return nil, err
	}
	time.Sleep(time.Second) // give other pachd nodes time to update their cache

	// Call PPS.ActivateAuth to set up all affected pipelines and repos
	superUserClient := pachClient.WithCtx(pachClient.Ctx()) // clone pachClient
	superUserClient.SetAuthToken(a.ppsToken)
	if _, err := superUserClient.ActivateAuth(superUserClient.Ctx(), &pps.ActivateAuthRequest{}); err != nil {
		return nil, err
	}

	// Generate a new Pachyderm token (as the caller is authenticating) and
	// initialize admins (watchAdmins() above will see the write)
	pachToken := uuid.NewWithoutDashes()
	if _, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		admins := a.admins.ReadWrite(stm)
		tokens := a.tokens.ReadWrite(stm)
		if err := admins.Delete(magicUser); err != nil {
			return err
		}
		if err := admins.Put(req.Subject, epsilon); err != nil {
			return err
		}
		return tokens.PutTTL(
			hashToken(pachToken),
			&authclient.TokenInfo{
				Subject: req.Subject,
				Source:  authclient.TokenInfo_AUTHENTICATE,
			},
			ttlSecs,
		)
	}); err != nil {
		return nil, err
	}

	// wait until watchAdmins has updated the local cache (changing the
	// activation state), so that Activate() is less likely to race with
	// subsequent calls that expect auth to be activated.
	// TODO this is a bit hacky (checking repeatedly in a spin loop) but
	// Activate() is rarely called, and it helps avoid races due to other pachd
	// pods being out of date.
	if err := backoff.Retry(func() error {
		if a.activationState() != full {
			return fmt.Errorf("auth never entered \"full\" activation state")
		}
		return nil
	}, b); err != nil {
		return nil, err
	}
	time.Sleep(time.Second) // give other pachd nodes time to update their cache
	return &authclient.ActivateResponse{PachToken: pachToken}, nil
}

func (a *apiServer) Deactivate(ctx context.Context, req *authclient.DeactivateRequest) (resp *authclient.DeactivateResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if a.activationState() == none {
		// users should be able to call "deactivate" from the "partial" activation
		// state, in case activation fails and they need to revert to "none"
		return nil, authclient.ErrNotActivated
	}

	// Check if the cluster is in a partially-activated state. If so, allow it to
	// be completely deactivated so that it returns to normal
	var magicUserIsAdmin bool
	func() {
		a.adminMu.Lock()
		defer a.adminMu.Unlock()
		_, magicUserIsAdmin = a.adminCache[magicUser]
	}()
	if magicUserIsAdmin {
		_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			a.admins.ReadWrite(stm).DeleteAll() // watchAdmins() will see the write
			return nil
		})
		if err != nil {
			return nil, err
		}
		return &authclient.DeactivateResponse{}, nil
	}

	// Get calling user. The user must be a cluster admin to disable auth for the
	// cluster
	callerInfo, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if !a.isAdmin(callerInfo.Subject) {
		return nil, &authclient.ErrNotAuthorized{
			Subject: callerInfo.Subject,
			AdminOp: "DeactivateAuth",
		}
	}
	_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		a.acls.ReadWrite(stm).DeleteAll()
		a.tokens.ReadWrite(stm).DeleteAll()
		a.admins.ReadWrite(stm).DeleteAll() // watchAdmins() will see the write
		a.members.ReadWrite(stm).DeleteAll()
		a.groups.ReadWrite(stm).DeleteAll()
		a.authConfig.ReadWrite(stm).DeleteAll()
		return nil
	})
	if err != nil {
		return nil, err
	}

	// wait until watchAdmins has deactivated auth, so that Deactivate() is less
	// likely to race with subsequent calls that expect auth to be deactivated.
	// TODO this is a bit hacky (checking repeatedly in a spin loop) but
	// Deactivate() is rarely called, and it helps avoid races due to other pachd
	// pods being out of date.
	if err := backoff.Retry(func() error {
		if a.activationState() != none {
			return fmt.Errorf("auth still activated")
		}
		return nil
	}, backoff.RetryEvery(time.Second)); err != nil {
		return nil, err
	}
	time.Sleep(time.Second) // give other pachd nodes time to update their cache
	return &authclient.DeactivateResponse{}, nil
}

// GitHubTokenToUsername takes a OAuth access token issued by GitHub and uses
// it discover the username of the user who obtained the code (or verify that
// the code belongs to githubUsername). This is how Pachyderm currently
// implements authorization in a production cluster
func GitHubTokenToUsername(ctx context.Context, oauthToken string) (string, error) {
	if !githubTokenRegex.MatchString(oauthToken) && os.Getenv(DisableAuthenticationEnvVar) == "true" {
		logrus.Warnf("Pachyderm is deployed in DEV mode. The provided auth token "+
			"will NOT be verified with GitHub; the caller is automatically "+
			"authenticated as the GitHub user \"%s\"", oauthToken)
		return authclient.GitHubPrefix + oauthToken, nil
	}

	// Initialize GitHub client with 'oauthToken'
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: oauthToken,
		},
	)
	tc := oauth2.NewClient(ctx, ts)
	gclient := github.NewClient(tc)

	// Retrieve the caller's GitHub Username (the empty string gets us the
	// authenticated user)
	user, _, err := gclient.Users.Get(ctx, "")
	if err != nil {
		return "", fmt.Errorf("error getting the authenticated user: %v", err)
	}
	verifiedUsername := user.GetLogin()
	return authclient.GitHubPrefix + verifiedUsername, nil
}

func (a *apiServer) GetAdmins(ctx context.Context, req *authclient.GetAdminsRequest) (resp *authclient.GetAdminsResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	switch a.activationState() {
	case none:
		return nil, authclient.ErrNotActivated
	case full:
		// Get calling user. There is no auth check to see the list of cluster
		// admins, other than that the user must log in. Otherwise how will users
		// know who to ask for admin privileges? Requiring the user to be logged in
		// mitigates phishing
		_, err := a.getAuthenticatedUser(ctx)
		if err != nil {
			return nil, err
		}
	}

	a.adminMu.Lock()
	defer a.adminMu.Unlock()
	resp = &authclient.GetAdminsResponse{
		Admins: make([]string, 0, len(a.adminCache)),
	}
	for admin := range a.adminCache {
		resp.Admins = append(resp.Admins, admin)
	}
	return resp, nil
}

func (a *apiServer) validateModifyAdminsRequest(add []string, remove []string) error {
	// Check to make sure that req doesn't remove all cluster admins
	m := make(map[string]struct{})
	// copy existing admins into m
	func() {
		a.adminMu.Lock()
		defer a.adminMu.Unlock()
		for u := range a.adminCache {
			m[u] = struct{}{}
		}
	}()
	for _, u := range add {
		m[u] = struct{}{}
	}
	for _, u := range remove {
		delete(m, u)
	}

	// Confirm that there will be at least one admin.
	//
	// This is required so that the admin can get the cluster out of any broken
	// state that it may enter.
	if len(m) == 0 {
		return fmt.Errorf("invalid request: cannot remove all cluster administrators while auth is active, to avoid unfixable cluster states")
	}
	return nil
}

func (a *apiServer) ModifyAdmins(ctx context.Context, req *authclient.ModifyAdminsRequest) (resp *authclient.ModifyAdminsResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	switch a.activationState() {
	case none:
		return nil, authclient.ErrNotActivated
	case partial:
		return nil, authclient.ErrPartiallyActivated
	}

	// Get calling user. The user must be an admin to change the list of admins
	callerInfo, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if !a.isAdmin(callerInfo.Subject) {
		return nil, &authclient.ErrNotAuthorized{
			Subject: callerInfo.Subject,
			AdminOp: "ModifyAdmins",
		}
	}

	// Canonicalize GitHub usernames in request (must canonicalize before we can
	// validate, so we know who is actually being added/removed & can confirm
	// that not all admins are being removed)
	eg := &errgroup.Group{}
	canonicalizedToAdd := make([]string, len(req.Add))
	for i, user := range req.Add {
		i, user := i, user
		eg.Go(func() error {
			user, err = a.lenientCanonicalizeSubject(ctx, user)
			if err != nil {
				return err
			}
			canonicalizedToAdd[i] = user
			return nil
		})
	}
	canonicalizedToRemove := make([]string, len(req.Remove))
	for i, user := range req.Remove {
		i, user := i, user
		eg.Go(func() error {
			user, err = a.lenientCanonicalizeSubject(ctx, user)
			if err != nil {
				return err
			}
			canonicalizedToRemove[i] = user
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	err = a.validateModifyAdminsRequest(canonicalizedToAdd, canonicalizedToRemove)
	if err != nil {
		return nil, err
	}

	// Update "admins" list (watchAdmins() will update admins cache)
	for _, user := range canonicalizedToAdd {
		if _, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			return a.admins.ReadWrite(stm).Put(user, epsilon)
		}); err != nil && retErr == nil {
			retErr = err
		}
	}
	for _, user := range canonicalizedToRemove {
		if _, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			return a.admins.ReadWrite(stm).Delete(user)
		}); err != nil && retErr == nil {
			retErr = err
		}
	}
	if retErr != nil {
		return nil, retErr
	}
	return &authclient.ModifyAdminsResponse{}, nil
}

// expiredClusterAdminCheck enforces that if the cluster's enterprise token is
// expired, only admins may log in.
func (a *apiServer) expiredClusterAdminCheck(username string) error {
	state, err := a.getEnterpriseTokenState()
	if err != nil {
		return fmt.Errorf("error confirming Pachyderm Enterprise token: %v", err)
	}
	if state != enterpriseclient.State_ACTIVE && !a.isAdmin(username) {
		return errors.New("Pachyderm Enterprise is not active in this " +
			"cluster (until Pachyderm Enterprise is re-activated or Pachyderm " +
			"auth is deactivated, only cluster admins can perform any operations)")
	}
	return nil
}

func (a *apiServer) Authenticate(ctx context.Context, req *authclient.AuthenticateRequest) (resp *authclient.AuthenticateResponse, retErr error) {
	switch a.activationState() {
	case none:
		// PPS is authenticated by a token read from etcd. It never calls or needs
		// to call authenticate, even while the cluster is partway through the
		// activation process
		return nil, authclient.ErrNotActivated
	case partial:
		return nil, authclient.ErrPartiallyActivated
	}

	// We don't want to actually log the request/response since they contain
	// credentials.
	defer func(start time.Time) { a.LogResp(nil, nil, retErr, time.Since(start)) }(time.Now())

	// verify whatever credential the user has presented, and write a new
	// Pachyderm token for the user that their credential belongs to
	var pachToken string
	switch {
	case req.GitHubToken != "":
		// Determine caller's Pachyderm/GitHub username
		username, err := GitHubTokenToUsername(ctx, req.GitHubToken)
		if err != nil {
			return nil, err
		}

		// If the cluster's enterprise token is expired, only admins may log in
		if err := a.expiredClusterAdminCheck(username); err != nil {
			return nil, err
		}

		// Generate a new Pachyderm token and write it
		pachToken = uuid.NewWithoutDashes()
		if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			tokens := a.tokens.ReadWrite(stm)
			return tokens.PutTTL(hashToken(pachToken),
				&authclient.TokenInfo{
					Subject: username,
					Source:  authclient.TokenInfo_AUTHENTICATE,
				},
				defaultTokenTTLSecs)
		}); err != nil {
			return nil, fmt.Errorf("error storing auth token for user \"%s\": %v", username, err)
		}

	case req.PachAuthenticationCode != "":
		if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			// read short-lived authentication code (and delete it if found)
			codes := a.authenticationCodes.ReadWrite(stm)
			key := hashToken(req.PachAuthenticationCode)
			var tokenInfo authclient.TokenInfo
			if err := codes.Get(key, &tokenInfo); err != nil {
				return err
			}
			codes.Delete(key)

			// If the cluster's enterprise token is expired, only admins may log in
			if err := a.expiredClusterAdminCheck(tokenInfo.Subject); err != nil {
				return err
			}

			// write long-lived pachyderm token
			pachToken = uuid.NewWithoutDashes()
			return a.tokens.ReadWrite(stm).PutTTL(hashToken(pachToken), &tokenInfo,
				defaultTokenTTLSecs)
		}); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unrecognized authentication mechanism (old pachd?)")
	}

	// Return new pachyderm token to caller
	return &authclient.AuthenticateResponse{
		PachToken: pachToken,
	}, nil
}

func (a *apiServer) GetAuthenticationCode(ctx context.Context, req *authclient.GetAuthenticationCodeRequest) (resp *authclient.GetAuthenticationCodeResponse, retErr error) {
	// We don't want to actually log the request/response since they contain
	// credentials.
	defer func(start time.Time) { a.LogResp(nil, nil, retErr, time.Since(start)) }(time.Now())
	switch a.activationState() {
	case none:
		// PPS is authenticated by a token read from etcd. It never calls or needs
		// to call authenticate, even while the cluster is partway through the
		// activation process
		return nil, authclient.ErrNotActivated
	case partial:
		return nil, authclient.ErrPartiallyActivated
	}

	callerInfo, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	code, err := a.getAuthenticationCode(ctx, callerInfo.Subject)
	if err != nil {
		return nil, err
	}
	return &authclient.GetAuthenticationCodeResponse{
		Code: code,
	}, nil
}

// getAuthenticationCode contains the implementation of GetAuthenticationCode,
// but is also called directly by handleSAMLREsponse. It generates a
// short-lived authentication code for 'username', writes it to
// a.authenticationCodes, and returns it
func (a *apiServer) getAuthenticationCode(ctx context.Context, username string) (code string, err error) {
	code = "auth_code:" + uuid.NewWithoutDashes()
	if _, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		return a.authenticationCodes.ReadWrite(stm).PutTTL(hashToken(code),
			&authclient.TokenInfo{
				Subject: username,
				Source:  authclient.TokenInfo_AUTHENTICATE,
			},
			defaultAuthCodeTTLSecs)
	}); err != nil {
		return "", err
	}
	return code, nil
}

func (a *apiServer) Authorize(ctx context.Context, req *authclient.AuthorizeRequest) (resp *authclient.AuthorizeResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if a.activationState() == none {
		return nil, authclient.ErrNotActivated
	}

	callerInfo, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	// admins are always authorized
	if a.isAdmin(callerInfo.Subject) {
		return &authclient.AuthorizeResponse{Authorized: true}, nil
	}

	if req.Repo == ppsconsts.SpecRepo {
		// All users are authorized to read from the spec repo, but only admins are
		// authorized to write to it (writing to it may break your cluster)
		return &authclient.AuthorizeResponse{
			Authorized: req.Scope == authclient.Scope_READER,
		}, nil
	}

	// If the cluster's enterprise token is expired, only admins and pipelines may
	// authorize (and admins are already handled)
	state, err := a.getEnterpriseTokenState()
	if err != nil {
		return nil, fmt.Errorf("error confirming Pachyderm Enterprise token: %v", err)
	}
	if state != enterpriseclient.State_ACTIVE &&
		!strings.HasPrefix(callerInfo.Subject, authclient.PipelinePrefix) {
		return nil, errors.New("Pachyderm Enterprise is not active in this " +
			"cluster (until Pachyderm Enterprise is re-activated or Pachyderm " +
			"auth is deactivated, only cluster admins can perform any operations)")
	}

	// Get ACL to check
	var acl authclient.ACL
	if err := a.acls.ReadOnly(ctx).Get(req.Repo, &acl); err != nil && !col.IsErrNotFound(err) {
		return nil, fmt.Errorf("error getting ACL for repo \"%s\": %v", req.Repo, err)
	}

	if req.Scope <= acl.Entries[callerInfo.Subject] {
		return &authclient.AuthorizeResponse{
			Authorized: true,
		}, nil
	}

	groupsResp, err := a.GetGroups(ctx, &authclient.GetGroupsRequest{
		Username: callerInfo.Subject,
	})
	if err != nil {
		return nil, fmt.Errorf("could not retrieve caller's group memberships: %v", err)
	}
	for _, g := range groupsResp.Groups {
		if req.Scope <= acl.Entries[g] {
			return &authclient.AuthorizeResponse{
				Authorized: true,
			}, nil
		}
	}

	return &authclient.AuthorizeResponse{
		Authorized: false,
	}, nil
}

func (a *apiServer) WhoAmI(ctx context.Context, req *authclient.WhoAmIRequest) (resp *authclient.WhoAmIResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if a.activationState() == none {
		return nil, authclient.ErrNotActivated
	}

	token, err := getAuthToken(ctx)
	if err != nil {
		return nil, err
	}
	callerInfo, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	ttl := int64(-1) // value returned by etcd for keys w/ no lease (no TTL)
	if callerInfo.Subject != magicUser {
		ttl, err = a.tokens.ReadOnly(ctx).TTL(hashToken(token)) // lookup token TTL
		if err != nil {
			return nil, fmt.Errorf("error looking up TTL for token: %v", err)
		}
	}
	return &authclient.WhoAmIResponse{
		Username: callerInfo.Subject,
		IsAdmin:  a.isAdmin(callerInfo.Subject),
		TTL:      ttl,
	}, nil
}

func validateSetScopeRequest(ctx context.Context, req *authclient.SetScopeRequest) error {
	if req.Username == "" {
		return fmt.Errorf("invalid request: must set username")
	}
	if req.Repo == "" {
		return fmt.Errorf("invalid request: must set repo")
	}
	return nil
}

func (a *apiServer) isAdmin(user string) bool {
	if user == magicUser {
		return true
	}
	a.adminMu.Lock()
	defer a.adminMu.Unlock()
	_, ok := a.adminCache[user]
	return ok
}

func (a *apiServer) SetScope(ctx context.Context, req *authclient.SetScopeRequest) (resp *authclient.SetScopeResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if a.activationState() == none {
		return nil, authclient.ErrNotActivated
	}
	pachClient := a.getPachClient().WithCtx(ctx)

	// validate request & authenticate user
	if err := validateSetScopeRequest(ctx, req); err != nil {
		return nil, err
	}
	callerInfo, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		acls := a.acls.ReadWrite(stm)
		var acl authclient.ACL
		if err := acls.Get(req.Repo, &acl); err != nil {
			if !col.IsErrNotFound(err) {
				return err
			}
			// ACL not found. Check that repo exists (return error if not). Note that
			// this is not consistent with the rest of the transaction, but users
			// should not be able to send a SetScope request that races with
			// CreateRepo
			_, err := pachClient.InspectRepo(req.Repo)
			if err != nil {
				return err
			}

			// Repo exists, but has no ACL. Create default (empty) ACL
			acl.Entries = make(map[string]authclient.Scope)
		}

		// Check if the caller is authorized
		authorized, err := func() (bool, error) {
			if a.isAdmin(callerInfo.Subject) {
				// admins are automatically authorized
				return true, nil
			}

			// Check if the cluster's enterprise token is expired (fail if so)
			state, err := a.getEnterpriseTokenState()
			if err != nil {
				return false, fmt.Errorf("error confirming Pachyderm Enterprise token: %v", err)
			}
			if state != enterpriseclient.State_ACTIVE {
				return false, fmt.Errorf("Pachyderm Enterprise is not active in this " +
					"cluster (only a cluster admin can set a scope)")
			}

			// Check if the user is on the ACL directly
			if acl.Entries[callerInfo.Subject] == authclient.Scope_OWNER {
				return true, nil
			}
			return false, nil
		}()
		if err != nil {
			return err
		}
		if !authorized {
			return &authclient.ErrNotAuthorized{
				Subject:  callerInfo.Subject,
				Repo:     req.Repo,
				Required: authclient.Scope_OWNER,
			}
		}

		// Scope change is authorized. Make the change
		principal, err := a.lenientCanonicalizeSubject(ctx, req.Username)
		if err != nil {
			return err
		}
		if req.Scope != authclient.Scope_NONE {
			acl.Entries[principal] = req.Scope
		} else {
			delete(acl.Entries, principal)
		}
		if len(acl.Entries) == 0 {
			return acls.Delete(req.Repo)
		}
		return acls.Put(req.Repo, &acl)
	})
	if err != nil {
		return nil, err
	}

	return &authclient.SetScopeResponse{}, nil
}

func (a *apiServer) GetScope(ctx context.Context, req *authclient.GetScopeRequest) (resp *authclient.GetScopeResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if a.activationState() == none {
		return nil, authclient.ErrNotActivated
	}

	callerInfo, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	// Check if the cluster's enterprise token is expired (fail if so)
	state, err := a.getEnterpriseTokenState()
	if err != nil {
		return nil, fmt.Errorf("error confirming Pachyderm Enterprise token: %v", err)
	}
	if state != enterpriseclient.State_ACTIVE && !a.isAdmin(callerInfo.Subject) {
		return nil, fmt.Errorf("Pachyderm Enterprise is not active in this " +
			"cluster (only a cluster admin can perform any operations)")
	}

	// For now, we don't return OWNER if the user is an admin, even though that's
	// their effective access scope for all repos--the caller may want to know
	// what will happen if the user's admin privileges are revoked

	// Read repo ACL from etcd
	acls := a.acls.ReadOnly(ctx)
	resp = new(authclient.GetScopeResponse)

	for _, repo := range req.Repos {
		var acl authclient.ACL
		if err := acls.Get(repo, &acl); err != nil && !col.IsErrNotFound(err) {
			return nil, err
		}

		// ACL read is authorized
		if req.Username == "" {
			resp.Scopes = append(resp.Scopes, acl.Entries[callerInfo.Subject])
		} else {
			// Caller is getting another user's scopes. Check if the caller is
			// authorized to view this repo's ACL
			if !a.isAdmin(callerInfo.Subject) && acl.Entries[callerInfo.Subject] < authclient.Scope_READER {
				return nil, &authclient.ErrNotAuthorized{
					Subject:  callerInfo.Subject,
					Repo:     repo,
					Required: authclient.Scope_READER,
				}
			}
			principal, err := a.lenientCanonicalizeSubject(ctx, req.Username)
			if err != nil {
				return nil, err
			}
			resp.Scopes = append(resp.Scopes, acl.Entries[principal])
		}
	}
	return resp, nil
}

func (a *apiServer) GetACL(ctx context.Context, req *authclient.GetACLRequest) (resp *authclient.GetACLResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if a.activationState() == none {
		return nil, authclient.ErrNotActivated
	}

	// Validate request
	if req.Repo == "" {
		return nil, fmt.Errorf("invalid request: must provide name of repo to get that repo's ACL")
	}

	// Get calling user
	callerInfo, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	// Check if the cluster's enterprise token is expired (fail if so)
	state, err := a.getEnterpriseTokenState()
	if err != nil {
		return nil, fmt.Errorf("error confirming Pachyderm Enterprise token: %v", err)
	}
	if state != enterpriseclient.State_ACTIVE && !a.isAdmin(callerInfo.Subject) {
		return nil, fmt.Errorf("Pachyderm Enterprise is not active in this " +
			"cluster (only a cluster admin can perform any operations)")
	}

	// Read repo ACL from etcd
	acl := &authclient.ACL{}
	if err = a.acls.ReadOnly(ctx).Get(req.Repo, acl); err != nil && !col.IsErrNotFound(err) {
		return nil, err
	}
	resp = &authclient.GetACLResponse{
		Entries: make([]*authclient.ACLEntry, 0),
	}
	for user, scope := range acl.Entries {
		resp.Entries = append(resp.Entries, &authclient.ACLEntry{
			Username: user,
			Scope:    scope,
		})
	}
	// For now, no access is require to read a repo's ACL
	// https://github.com/pachyderm/pachyderm/issues/2353
	return resp, nil
}

func (a *apiServer) SetACL(ctx context.Context, req *authclient.SetACLRequest) (resp *authclient.SetACLResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if a.activationState() == none {
		return nil, authclient.ErrNotActivated
	}

	// Validate request
	if req.Repo == "" {
		return nil, fmt.Errorf("invalid request: must provide name of repo you want to modify")
	}

	// Get calling user
	callerInfo, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	// Canonicalize entries in the request (must have canonical request before we
	// can authorize, as we inspect the ACL contents during authorization in the
	// case where we're creating a new repo)
	eg := &errgroup.Group{}
	var aclMu sync.Mutex
	newACL := new(authclient.ACL)
	if len(req.Entries) > 0 {
		newACL.Entries = make(map[string]authclient.Scope)
	}
	for _, entry := range req.Entries {
		user, scope := entry.Username, entry.Scope
		eg.Go(func() error {
			principal, err := a.lenientCanonicalizeSubject(ctx, user)
			if err != nil {
				return err
			}
			aclMu.Lock()
			defer aclMu.Unlock()
			newACL.Entries[principal] = scope
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// Read repo ACL from etcd
	_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		acls := a.acls.ReadWrite(stm)

		// determine if the caller is authorized to set this repo's ACL
		authorized, err := func() (bool, error) {
			if a.isAdmin(callerInfo.Subject) {
				// admins are automatically authorized
				return true, nil
			}

			// Check if the cluster's enterprise token is expired (fail if so)
			state, err := a.getEnterpriseTokenState()
			if err != nil {
				return false, fmt.Errorf("error confirming Pachyderm Enterprise token: %v", err)
			}
			if state != enterpriseclient.State_ACTIVE {
				return false, fmt.Errorf("Pachyderm Enterprise is not active in this " +
					"cluster (only a cluster admin can modify an ACL)")
			}

			// Check if there is an existing ACL, and if the user is on it
			var acl authclient.ACL
			if err := acls.Get(req.Repo, &acl); err != nil {
				// ACL not found -- construct empty ACL proto
				acl.Entries = make(map[string]authclient.Scope)
			}
			if len(acl.Entries) > 0 {
				// ACL is present; caller must be authorized directly
				if acl.Entries[callerInfo.Subject] == authclient.Scope_OWNER {
					return true, nil
				}
				return false, nil
			}

			// No ACL -- check if the repo being modified exists
			_, err = a.getPachClient().WithCtx(ctx).InspectRepo(req.Repo)
			err = grpcutil.ScrubGRPC(err)
			if err == nil {
				// Repo exists -- user isn't authorized
				return false, nil
			} else if !strings.HasSuffix(err.Error(), "not found") {
				// Unclear if repo exists -- return error
				return false, fmt.Errorf("could not inspect \"%s\": %v", req.Repo, err)
			} else if len(newACL.Entries) == 1 &&
				newACL.Entries[callerInfo.Subject] == authclient.Scope_OWNER {
				// Special case: Repo doesn't exist, but user is creating a new Repo, and
				// making themself the owner, e.g. for CreateRepo or CreatePipeline, then
				// the request is authorized
				return true, nil
			}
			return false, err
		}()
		if err != nil {
			return err
		}
		if !authorized {
			return &authclient.ErrNotAuthorized{
				Subject:  callerInfo.Subject,
				Repo:     req.Repo,
				Required: authclient.Scope_OWNER,
			}
		}

		// Set new ACL
		if len(newACL.Entries) == 0 {
			return acls.Delete(req.Repo)
		}
		return acls.Put(req.Repo, newACL)
	})
	if err != nil {
		return nil, fmt.Errorf("could not put new ACL: %v", err)
	}
	return &authclient.SetACLResponse{}, nil
}

func (a *apiServer) GetAuthToken(ctx context.Context, req *authclient.GetAuthTokenRequest) (resp *authclient.GetAuthTokenResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if a.activationState() == none {
		return nil, authclient.ErrNotActivated
	}
	if req.Subject == "" {
		return nil, fmt.Errorf("must set GetAuthTokenRequest.Subject")
	}
	if req.Subject == magicUser {
		return nil, fmt.Errorf("GetAuthTokenRequest.Subject is invalid")
	}

	// Authorize caller
	callerInfo, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if req.Subject != callerInfo.Subject && !a.isAdmin(callerInfo.Subject) {
		return nil, &authclient.ErrNotAuthorized{
			Subject: callerInfo.Subject,
			AdminOp: "GetAuthToken on behalf of another user",
		}
	}
	subject, err := a.lenientCanonicalizeSubject(ctx, req.Subject)
	if err != nil {
		return nil, err
	}
	tokenInfo := authclient.TokenInfo{
		Source:  authclient.TokenInfo_GET_TOKEN,
		Subject: subject,
	}

	// generate new token, and write to etcd
	token := uuid.NewWithoutDashes()
	if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		return a.tokens.ReadWrite(stm).PutTTL(hashToken(token), &tokenInfo, req.TTL)
	}); err != nil {
		if tokenInfo.Subject != magicUser {
			return nil, fmt.Errorf("error storing token for user \"%s\": %v", tokenInfo.Subject, err)
		}
		return nil, fmt.Errorf("error storing token: %v", err)
	}
	return &authclient.GetAuthTokenResponse{
		Subject: subject,
		Token:   token,
	}, nil
}

func (a *apiServer) ExtendAuthToken(ctx context.Context, req *authclient.ExtendAuthTokenRequest) (resp *authclient.ExtendAuthTokenResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if a.activationState() != full {
		return nil, authclient.ErrNotActivated
	}
	if req.TTL == 0 {
		return nil, fmt.Errorf("invalid request: ExtendAuthTokenRequest.TTL must be > 0")
	}

	// Only admins can extend auth tokens (for now)
	callerInfo, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if !a.isAdmin(callerInfo.Subject) {
		return nil, &authclient.ErrNotAuthorized{
			Subject: callerInfo.Subject,
			AdminOp: "ExtendAuthToken",
		}
	}

	// Only let people extend tokens by up to 30 days (the equivalent of logging
	// in again)
	if req.TTL > defaultTokenTTLSecs {
		return nil, fmt.Errorf("can only extend tokens by at most %d seconds", defaultTokenTTLSecs)
	}

	// The token must already exist. If a token has been revoked, it can't be
	// extended
	if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		tokens := a.tokens.ReadWrite(stm)

		// Actually look up the request token in the relevant collections
		var tokenInfo authclient.TokenInfo
		if err := tokens.Get(hashToken(req.Token), &tokenInfo); err != nil && !col.IsErrNotFound(err) {
			return err
		}
		if tokenInfo.Subject == "" {
			return authclient.ErrBadToken
		}

		ttl, err := tokens.TTL(hashToken(req.Token))
		if err != nil {
			return fmt.Errorf("Error looking up TTL for token: %v", err)
		}
		// TODO(msteffen): ttl may be -1 if the token has no TTL. We deliberately do
		// not check this case so that admins can put TTLs on tokens that don't have
		// them (otherwise any attempt to do so would get ErrTooShortTTL), but that
		// decision may be revised
		if req.TTL < ttl {
			return authclient.ErrTooShortTTL{
				RequestTTL:  req.TTL,
				ExistingTTL: ttl,
			}
		}
		return tokens.PutTTL(hashToken(req.Token), &tokenInfo, req.TTL)
	}); err != nil {
		return nil, err
	}
	return &authclient.ExtendAuthTokenResponse{}, nil
}

func (a *apiServer) RevokeAuthToken(ctx context.Context, req *authclient.RevokeAuthTokenRequest) (resp *authclient.RevokeAuthTokenResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if a.activationState() != full {
		return nil, authclient.ErrNotActivated
	}

	// Get the caller. Users can revoke their own tokens, and admins can revoke
	// any user's token
	callerInfo, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		tokens := a.tokens.ReadWrite(stm)
		var tokenInfo authclient.TokenInfo
		if err := tokens.Get(hashToken(req.Token), &tokenInfo); err != nil {
			if col.IsErrNotFound(err) {
				return nil
			}
			return err
		}
		if !a.isAdmin(callerInfo.Subject) && tokenInfo.Subject != callerInfo.Subject {
			return &authclient.ErrNotAuthorized{
				Subject: callerInfo.Subject,
				AdminOp: "RevokeAuthToken on another user's token",
			}
		}
		return tokens.Delete(hashToken(req.Token))
	}); err != nil {
		return nil, err
	}
	return &authclient.RevokeAuthTokenResponse{}, nil
}

func (a *apiServer) setGroupsForUser(ctx context.Context, username string, groups []string) error {
	_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		members := a.members.ReadWrite(stm)

		// Get groups to remove/add user from/to
		var removeGroups authclient.Groups
		addGroups := addToSet(nil, groups...)
		if err := members.Get(username, &removeGroups); err == nil {
			for _, group := range groups {
				if removeGroups.Groups[group] {
					removeGroups.Groups = removeFromSet(removeGroups.Groups, group)
					addGroups = removeFromSet(addGroups, group)
				}
			}
		}

		// Set groups for user
		if err := members.Put(username, &authclient.Groups{
			Groups: addToSet(nil, groups...),
		}); err != nil {
			return err
		}

		// Remove user from previous groups
		groups := a.groups.ReadWrite(stm)
		var membersProto authclient.Users
		for group := range removeGroups.Groups {
			if err := groups.Upsert(group, &membersProto, func() error {
				membersProto.Usernames = removeFromSet(membersProto.Usernames, username)
				return nil
			}); err != nil {
				return err
			}
		}

		// Add user to new groups
		for group := range addGroups {
			if err := groups.Upsert(group, &membersProto, func() error {
				membersProto.Usernames = addToSet(membersProto.Usernames, username)
				return nil
			}); err != nil {
				return err
			}
		}

		return nil
	})
	return err
}

func (a *apiServer) SetGroupsForUser(ctx context.Context, req *authclient.SetGroupsForUserRequest) (resp *authclient.SetGroupsForUserResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if a.activationState() != full {
		return nil, authclient.ErrNotActivated
	}

	callerInfo, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	if !a.isAdmin(callerInfo.Subject) {
		return nil, &authclient.ErrNotAuthorized{
			Subject: callerInfo.Subject,
			AdminOp: "SetGroupsForUser",
		}
	}

	username, err := a.lenientCanonicalizeSubject(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	if err := a.setGroupsForUser(ctx, username, req.Groups); err != nil {
		return nil, err
	}
	return &authclient.SetGroupsForUserResponse{}, nil
}

func (a *apiServer) ModifyMembers(ctx context.Context, req *authclient.ModifyMembersRequest) (resp *authclient.ModifyMembersResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if a.activationState() != full {
		return nil, authclient.ErrNotActivated
	}

	callerInfo, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	if !a.isAdmin(callerInfo.Subject) {
		return nil, &authclient.ErrNotAuthorized{
			Subject: callerInfo.Subject,
			AdminOp: "ModifyMembers",
		}
	}

	add, err := a.lenientCanonicalizeSubjects(ctx, req.Add)
	if err != nil {
		return nil, err
	}
	// TODO(bryce) Skip canonicalization if the users can be found.
	remove, err := a.lenientCanonicalizeSubjects(ctx, req.Remove)
	if err != nil {
		return nil, err
	}

	if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		members := a.members.ReadWrite(stm)
		var groupsProto authclient.Groups
		for _, username := range add {
			if err := members.Upsert(username, &groupsProto, func() error {
				groupsProto.Groups = addToSet(groupsProto.Groups, req.Group)
				return nil
			}); err != nil {
				return err
			}
		}
		for _, username := range remove {
			if err := members.Upsert(username, &groupsProto, func() error {
				groupsProto.Groups = removeFromSet(groupsProto.Groups, req.Group)
				return nil
			}); err != nil {
				return err
			}
		}

		groups := a.groups.ReadWrite(stm)
		var membersProto authclient.Users
		if err := groups.Upsert(req.Group, &membersProto, func() error {
			membersProto.Usernames = addToSet(membersProto.Usernames, add...)
			membersProto.Usernames = removeFromSet(membersProto.Usernames, remove...)
			return nil
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &authclient.ModifyMembersResponse{}, nil
}

func addToSet(set map[string]bool, elems ...string) map[string]bool {
	if set == nil {
		set = map[string]bool{}
	}

	for _, elem := range elems {
		set[elem] = true
	}
	return set
}

func removeFromSet(set map[string]bool, elems ...string) map[string]bool {
	if set != nil {
		for _, elem := range elems {
			delete(set, elem)
		}
	}

	return set
}

func (a *apiServer) GetGroups(ctx context.Context, req *authclient.GetGroupsRequest) (resp *authclient.GetGroupsResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if a.activationState() != full {
		return nil, authclient.ErrNotActivated
	}

	callerInfo, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	// must be admin or user getting your own groups
	if req.Username != callerInfo.Subject {
		if !a.isAdmin(callerInfo.Subject) {
			return nil, &authclient.ErrNotAuthorized{
				Subject: callerInfo.Subject,
				AdminOp: "GetGroups",
			}
		}
	}

	// Filter by username
	if req.Username != "" {
		username, err := a.lenientCanonicalizeSubject(ctx, req.Username)
		if err != nil {
			return nil, err
		}

		var groupsProto authclient.Groups
		if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			members := a.members.ReadWrite(stm)
			if err := members.Get(username, &groupsProto); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return nil, err
		}

		return &authclient.GetGroupsResponse{Groups: setToList(groupsProto.Groups)}, nil
	}

	groupsCol := a.groups.ReadOnly(ctx)
	var groups []string
	users := &authclient.Users{}
	if err := groupsCol.List(users, col.DefaultOptions, func(group string) error {
		groups = append(groups, group)
		return nil
	}); err != nil {
		return nil, err
	}
	return &authclient.GetGroupsResponse{Groups: groups}, nil
}

func (a *apiServer) GetUsers(ctx context.Context, req *authclient.GetUsersRequest) (resp *authclient.GetUsersResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if a.activationState() != full {
		return nil, authclient.ErrNotActivated
	}

	callerInfo, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	if !a.isAdmin(callerInfo.Subject) {
		return nil, &authclient.ErrNotAuthorized{
			Subject: callerInfo.Subject,
			AdminOp: "GetUsers",
		}
	}

	// Filter by group
	if req.Group != "" {
		var membersProto authclient.Users
		if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			groups := a.groups.ReadWrite(stm)
			if err := groups.Get(req.Group, &membersProto); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return nil, err
		}

		return &authclient.GetUsersResponse{Usernames: setToList(membersProto.Usernames)}, nil
	}

	membersCol := a.members.ReadOnly(ctx)
	groups := &authclient.Groups{}
	var users []string
	if err := membersCol.List(groups, col.DefaultOptions, func(user string) error {
		users = append(users, user)
		return nil
	}); err != nil {
		return nil, err
	}
	return &authclient.GetUsersResponse{Usernames: users}, nil
}

func setToList(set map[string]bool) []string {
	if set == nil {
		return []string{}
	}

	list := []string{}
	for elem := range set {
		list = append(list, elem)
	}
	return list
}

// hashToken converts a token to a cryptographic hash.
// We don't want to store tokens verbatim in the database, as then whoever
// that has access to the database has access to all tokens.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", sum)
}

// getAuthToken extracts the auth token embedded in 'ctx', if there is on
func getAuthToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", authclient.ErrNoMetadata
	}
	if len(md[authclient.ContextTokenKey]) > 1 {
		return "", fmt.Errorf("multiple authentication token keys found in context")
	} else if len(md[authclient.ContextTokenKey]) == 0 {
		return "", authclient.ErrNotSignedIn
	}
	return md[authclient.ContextTokenKey][0], nil
}

func (a *apiServer) getAuthenticatedUser(ctx context.Context) (*authclient.TokenInfo, error) {
	// TODO(msteffen) cache these lookups, especially since users always authorize
	// themselves at the beginning of a request. Don't want to look up the same
	// token -> username entry twice.
	token, err := getAuthToken(ctx)
	if err != nil {
		return nil, err
	}
	if token == a.ppsToken {
		// TODO(msteffen): This is a hack. The idea is that there is a logical user
		// entry mapping ppsToken to magicUser. Soon, magicUser will go away and
		// this check should happen in authorize
		return &authclient.TokenInfo{
			Subject: magicUser,
			Source:  authclient.TokenInfo_GET_TOKEN,
		}, nil
	}

	// Lookup the token
	var tokenInfo authclient.TokenInfo
	if err := a.tokens.ReadOnly(ctx).Get(hashToken(token), &tokenInfo); err != nil {
		if col.IsErrNotFound(err) {
			return nil, authclient.ErrBadToken
		}
		return nil, err
	}
	return &tokenInfo, nil
}

// lenientCanonicalizeSubjects applies lenientCanonicalizeSubject to a list
func (a *apiServer) lenientCanonicalizeSubjects(ctx context.Context, subjects []string) ([]string, error) {
	if subjects == nil {
		return []string{}, nil
	}

	eg := &errgroup.Group{}
	canonicalizedSubjects := make([]string, len(subjects))
	for i, subject := range subjects {
		i, subject := i, subject
		eg.Go(func() error {
			subject, err := a.lenientCanonicalizeSubject(ctx, subject)
			if err != nil {
				return err
			}
			canonicalizedSubjects[i] = subject
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return canonicalizedSubjects, nil
}

func (a *apiServer) lenientValidateSubject(subject string) error {
	return nil
}

// lenientCanonicalizeSubject is like 'canonicalizeSubject', except that if
// 'subject' has no prefix, they are assumed to be a GitHub user.
func (a *apiServer) lenientCanonicalizeSubject(ctx context.Context, subject string) (string, error) {
	if strings.Index(subject, ":") < 0 {
		// assume default 'github:' prefix (then canonicalize the GitHub username in
		// canonicalizeSubject)
		subject = authclient.GitHubPrefix + subject
	}
	return a.canonicalizeSubject(ctx, subject)
}

// canonicalizeSubject establishes the type of 'subject' by looking for one of
// pachyderm's subject prefixes, and then canonicalizes the subject based on
// that.
func (a *apiServer) canonicalizeSubject(ctx context.Context, subject string) (string, error) {
	colonIdx := strings.Index(subject, ":")
	if colonIdx < 0 {
		return "", fmt.Errorf("subject must have a type prefix (e.g. \"github:\" or \"pachyderm_robot:\") but had none")
	}
	prefix := subject[:colonIdx+1]
	switch prefix {
	case authclient.GitHubPrefix:
		var err error
		subject, err = canonicalizeGitHubUsername(ctx, subject[len(authclient.GitHubPrefix):])
		if err != nil {
			return "", err
		}
	case authclient.PipelinePrefix, authclient.RobotPrefix, authclient.SAMLPrefix:
		break
	case path.Join("group", authclient.SAMLPrefix):
		break
	default:
		return "", fmt.Errorf("subject has unrecognized prefix: %s", subject[:colonIdx+1])
	}
	return subject, nil
}

// canonicalizeGitHubUsername corrects 'user' for case errors by looking
// up the corresponding user's GitHub profile and extracting their login ID
// from that. 'user' should not have any subject prefixes (as they are required
// to be a GitHub user).
func canonicalizeGitHubUsername(ctx context.Context, user string) (string, error) {
	if strings.Index(user, ":") >= 0 {
		return "", fmt.Errorf("invalid username has multiple prefixes: %s%s", authclient.GitHubPrefix, user)
	}
	if os.Getenv(DisableAuthenticationEnvVar) == "true" {
		// authentication is off -- user might not even be real
		return authclient.GitHubPrefix + user, nil
	}
	gclient := github.NewClient(http.DefaultClient)
	u, _, err := gclient.Users.Get(ctx, strings.ToLower(user))
	if err != nil {
		return "", fmt.Errorf("error canonicalizing \"%s\": %v", user, err)
	}
	return authclient.GitHubPrefix + u.GetLogin(), nil
}

func (a *apiServer) GetConfiguration(ctx context.Context, req *authclient.GetConfigurationRequest) (resp *authclient.GetConfigurationResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	switch a.activationState() {
	case none:
		return nil, authclient.ErrNotActivated
	case partial:
		return nil, authclient.ErrPartiallyActivated
	}

	// Get calling user. The user must be logged in to get the cluster config
	_, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	// Retrieve & return configuration
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	authConfigRO := a.authConfig.ReadOnly(ctx)

	var currentConfig authclient.AuthConfig
	if err := authConfigRO.Get(configKey, &currentConfig); err != nil && !col.IsErrNotFound(err) {
		return nil, err
	}
	a.configMu.Lock()
	defer a.configMu.Unlock()
	if a.configCache.LiveConfigVersion < currentConfig.LiveConfigVersion {
		logrus.Printf("current config (v.%d) is newer than cache (v.%d); updating cache",
			currentConfig.LiveConfigVersion, a.configCache.LiveConfigVersion)
		a.configCache = currentConfig
	} else if a.configCache.LiveConfigVersion > currentConfig.LiveConfigVersion {
		logrus.Warnln("config cache is NEWER than live config; this shouldn't happen")
	}
	return &authclient.GetConfigurationResponse{
		Configuration: &currentConfig,
	}, nil
}

func (a *apiServer) SetConfiguration(ctx context.Context, req *authclient.SetConfigurationRequest) (resp *authclient.SetConfigurationResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	switch a.activationState() {
	case none:
		return nil, authclient.ErrNotActivated
	case partial:
		return nil, authclient.ErrPartiallyActivated
	}

	// Get calling user. The user must be an admin to set the cluster config
	callerInfo, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if !a.isAdmin(callerInfo.Subject) {
		return nil, &authclient.ErrNotAuthorized{
			Subject: callerInfo.Subject,
			AdminOp: "SetConfiguration",
		}
	}

	// Validate new config
	for _, idp := range req.Configuration.IDProviders {
		if idp.Name == "" {
			return nil, errors.New("All ID providers must have a name specified " +
				"(for use during authorization)")
		}
		// TODO(msteffen): make sure we don't have to extend this every time we add
		// a new built-in backend.
		switch idp.Name {
		case authclient.GitHubPrefix:
			return nil, errors.New("cannot configure auth backend with reserved prefix " +
				authclient.GitHubPrefix)
		case authclient.RobotPrefix:
			return nil, errors.New("cannot configure auth backend with reserved prefix " +
				authclient.RobotPrefix)
		case authclient.PipelinePrefix:
			return nil, errors.New("cannot configure auth backend with reserved prefix " +
				authclient.PipelinePrefix)
		}
	}

	// upsert new config
	if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		var currentConfig authclient.AuthConfig
		return a.authConfig.ReadWrite(stm).Upsert(configKey, &currentConfig, func() error {
			if currentConfig.LiveConfigVersion != req.Configuration.LiveConfigVersion {
				return fmt.Errorf("expected config version %d, but live config has version %d",
					req.Configuration.LiveConfigVersion, currentConfig.LiveConfigVersion)
			}
			currentConfig.Reset()
			currentConfig = *req.Configuration
			currentConfig.LiveConfigVersion++
			return nil
		})
	}); err != nil {
		return nil, err
	}
	return &authclient.SetConfigurationResponse{}, nil
}
