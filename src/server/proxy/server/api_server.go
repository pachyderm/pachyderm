package server

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/proxy"
	"go.uber.org/zap"
)

// Env is the set of dependencies required for APIServer
type Env struct {
	Listener collection.PostgresListener
}

type APIServer struct {
	proxy.UnimplementedAPIServer
	env Env
}

func NewAPIServer(env Env) *APIServer {
	return &APIServer{
		env: env,
	}
}

// Listen streams database events.
// It signals that it is internally set up by sending an initial empty ListenResponse.
func (a *APIServer) Listen(request *proxy.ListenRequest, server proxy.API_ListenServer) (retErr error) {
	listener := a.env.Listener
	notifier := newNotifier(server, request.Channel)
	if err := listener.Register(notifier); err != nil {
		return errors.EnsureStack(err)
	}

	// send initial empty event to indicate to client that the listener has been registered
	if err := server.Send(&proxy.ListenResponse{}); err != nil {
		return errors.EnsureStack(err)
	}

	go notifier.send()

	defer func() {
		if err := listener.Unregister(notifier); err != nil {
			log.Error(pctx.TODO(), "errored while unregistering notifier", zap.Error(err))
		}
	}()
	return <-notifier.errChan
}

type notifier struct {
	server  proxy.API_ListenServer
	id      string
	channel string
	bufChan chan *collection.Notification
	errChan chan error
}

func newNotifier(server proxy.API_ListenServer, channel string) *notifier {
	return &notifier{
		server:  server,
		id:      uuid.NewWithoutDashes(),
		channel: channel,
		bufChan: make(chan *collection.Notification, collection.ChannelBufferSize),
		errChan: make(chan error, 1),
	}
}

func (n *notifier) ID() string {
	return n.id
}

func (n *notifier) Channel() string {
	return n.channel
}

func (n *notifier) Notify(notification *collection.Notification) {
	select {
	case n.bufChan <- notification:
	case <-n.server.Context().Done():
		n.sendError(n.server.Context().Err())
	default:
		n.sendError(errors.New("listener buffer is full, aborting listen"))
	}
}

func (n *notifier) send() {
	for {
		select {
		case notification := <-n.bufChan:
			if err := n.server.Send(&proxy.ListenResponse{
				Extra: notification.Extra,
			}); err != nil {
				n.sendError(err)
				return
			}
		case <-n.server.Context().Done():
			n.sendError(n.server.Context().Err())
			return
		}
	}
}

func (n *notifier) Error(err error) {
	n.sendError(err)
}

func (n *notifier) sendError(err error) {
	select {
	case n.errChan <- err:
	default:
	}
}
