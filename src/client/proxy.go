package client

import (
	"sync"

	"github.com/lib/pq"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/proxy"
	"golang.org/x/net/context"
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
	client       proxy.APIClient
	mu           sync.Mutex
	channelInfos map[string]*channelInfo
}

func NewProxyPostgresListener(client proxy.APIClient) col.PostgresListener {
	return &proxyPostgresListener{
		client:       client,
		channelInfos: make(map[string]*channelInfo),
	}
}

func (ppl *proxyPostgresListener) Register(notifier col.Notifier) error {
	ppl.mu.Lock()
	defer ppl.mu.Unlock()
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
	go func() {
		if err := func() error {
			listenClient, err := ppl.client.Listen(ctx, &proxy.ListenRequest{
				Channel: channel,
			})
			if err != nil {
				return err
			}
			for {
				resp, err := listenClient.Recv()
				if err != nil {
					return err
				}
				ppl.mu.Lock()
				if ci, ok := ppl.channelInfos[channel]; ok {
					for _, notifier := range ci.notifiers {
						notifier.Notify(&pq.Notification{
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
