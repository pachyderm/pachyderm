package client

import (
	"context"
	"sync"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/proxy"
)

type channelInfo struct {
	cancel    context.CancelFunc
	notifiers map[string]col.Notifier
}

func newChannelInfo(cancel context.CancelFunc, notifier col.Notifier) *channelInfo {
	c := &channelInfo{
		cancel:    cancel,
		notifiers: make(map[string]col.Notifier),
	}
	c.notifiers[notifier.ID()] = notifier
	return c
}

type proxyPostgresListener struct {
	clientFactory func() (proxy.APIClient, error)
	client        proxy.APIClient
	mu            sync.Mutex
	channelInfos  map[string]*channelInfo
}

func NewProxyPostgresListener(clientFactory func() (proxy.APIClient, error)) col.PostgresListener {
	return &proxyPostgresListener{
		clientFactory: clientFactory,
		channelInfos:  make(map[string]*channelInfo),
	}
}

func (ppl *proxyPostgresListener) setup() error {
	if ppl.client != nil {
		return nil
	}
	var err error
	ppl.client, err = ppl.clientFactory()
	return err
}

// TODO: for now, calls to Register() will block on the Listener being established
// on the main pachd instance. This could become a bottleneck if many Watchers are
// instantiated at once. Consider finer grained locking at the channel level using
// a mechanism such as the "WorkDeduper" (src/internal/miscutil/work_deduper.go).
func (ppl *proxyPostgresListener) Register(notifier col.Notifier) error {
	ppl.mu.Lock()
	defer ppl.mu.Unlock()
	if err := ppl.setup(); err != nil {
		return err
	}
	if ci, ok := ppl.channelInfos[notifier.Channel()]; ok {
		ci.notifiers[notifier.ID()] = notifier
		return nil
	}
	ppl.listen(notifier)
	return nil
}

func (ppl *proxyPostgresListener) listen(notifier col.Notifier) {
	channel := notifier.Channel()
	ctx, cancel := context.WithCancel(context.Background())
	ppl.channelInfos[channel] = newChannelInfo(cancel, notifier)

	var listenClient proxy.API_ListenClient
	if err := func() error {
		var err error
		listenClient, err = ppl.client.Listen(ctx, &proxy.ListenRequest{
			Channel: channel,
		})
		if err != nil {
			return err
		}
		// wait for the first init event to be returned by the server
		_, err = listenClient.Recv()
		return err
	}(); err != nil {
		notifier.Error(err)
		cancel()
		delete(ppl.channelInfos, channel)
		return
	}

	go func() {
		if err := func() error {
			for {
				resp, err := listenClient.Recv()
				if err != nil {
					return err
				}
				ppl.mu.Lock()
				if ci, ok := ppl.channelInfos[channel]; ok {
					for _, notifier := range ci.notifiers {
						notifier.Notify(&col.Notification{
							Channel: channel,
							Extra:   resp.Extra,
						})
					}
				}
				ppl.mu.Unlock()
			}
		}(); err != nil {
			ppl.mu.Lock()
			ci, ok := ppl.channelInfos[channel]
			if ok {
				for _, notifier := range ci.notifiers {
					notifier.Error(err)
				}
				ci.cancel()
				delete(ppl.channelInfos, channel)
			}
			ppl.mu.Unlock()
		}
	}()
}

func (ppl *proxyPostgresListener) Unregister(notifier col.Notifier) error {
	ppl.mu.Lock()
	defer ppl.mu.Unlock()
	id := notifier.ID()
	channel := notifier.Channel()
	if ci, ok := ppl.channelInfos[channel]; ok {
		delete(ci.notifiers, id)
		if len(ci.notifiers) == 0 {
			ci.cancel()
			delete(ppl.channelInfos, channel)
		}
	}
	return nil
}

func (ppl *proxyPostgresListener) Close() error {
	ppl.mu.Lock()
	defer ppl.mu.Unlock()
	for channel, ci := range ppl.channelInfos {
		for _, notifier := range ci.notifiers {
			notifier.Error(errors.New("proxy postgres listener has been closed"))
		}
		ci.cancel()
		delete(ppl.channelInfos, channel)
	}
	return nil
}
