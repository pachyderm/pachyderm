package server

import (
	"errors"

	"github.com/lib/pq"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/proxy"
)

type APIServer struct {
	env serviceenv.ServiceEnv
}

func NewAPIServer(env serviceenv.ServiceEnv) *APIServer {
	return &APIServer{
		env: env,
	}
}

// TODO: Log?
func (a *APIServer) Listen(request *proxy.ListenRequest, server proxy.API_ListenServer) (retErr error) {
	listener := a.env.GetPostgresListener()
	notifier := newNotifier(server, request.Channel)
	if err := listener.Register(notifier); err != nil {
		return err
	}
	defer func() {
		if err := listener.Unregister(notifier); retErr == nil {
			retErr = err
		}
	}()
	select {
	case err := <-notifier.errChan:
		return err
	case <-server.Context().Done():
		return server.Context().Err()
	}
}

type notifier struct {
	server  proxy.API_ListenServer
	id      string
	channel string
	bufChan chan *pq.Notification
	errChan chan error
}

func newNotifier(server proxy.API_ListenServer, channel string) *notifier {
	n := &notifier{
		server:  server,
		id:      uuid.NewWithoutDashes(),
		channel: channel,
		bufChan: make(chan *pq.Notification, col.ChannelBufferSize),
		errChan: make(chan error),
	}
	go n.send()
	return n
}

func (n *notifier) ID() string {
	return n.id
}

func (n *notifier) Channel() string {
	return n.channel
}

func (n *notifier) Notify(notification *pq.Notification) {
	select {
	case n.bufChan <- notification:
	case <-n.server.Context().Done():
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
			}
		case <-n.server.Context().Done():
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
	case <-n.server.Context().Done():
	}
}
