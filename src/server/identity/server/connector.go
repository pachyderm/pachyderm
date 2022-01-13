package server

import (
	"context"
	"net/http"
	"net/url"

	"github.com/dexidp/dex/connector"
	"github.com/dexidp/dex/pkg/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

var (
	_ connector.CallbackConnector = &placeholder{}
)

// placeholder is a fake Dex connector which redirects the user to a static page with instructions.
// This is necessary because Dex won't start the web server unless a connector is configured.
type placeholderConfig struct{}

func (placeholderConfig) Open(id string, logger log.Logger) (connector.Connector, error) {
	return &placeholder{}, nil
}

type placeholder struct{}

func (*placeholder) LoginURL(s connector.Scopes, callbackURL, state string, _ url.Values) (string, error) {
	return "/static/not-configured.html", nil
}

// HandleCallback parses the request and returns the user's identity
func (*placeholder) HandleCallback(s connector.Scopes, r *http.Request) (connector.Identity, error) {
	return connector.Identity{}, errors.Errorf("no connectors configured")
}

// Refresh updates the identity during a refresh token request.
func (*placeholder) Refresh(ctx context.Context, s connector.Scopes, identity connector.Identity) (connector.Identity, error) {
	return connector.Identity{}, errors.Errorf("no connectors configured")
}
