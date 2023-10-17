package server

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"go.uber.org/zap"
)

type tokenResp struct {
	IdToken string `json:"id_token"`
}

type idToken struct {
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Groups        []string `json:"groups"`
}

func (w *dexWeb) interceptToken(next http.Handler) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ps, err := w.provisioners(ctx)
		if err != nil {
			log.Error(ctx, "failed to collect user provisioners", zap.Error(err))
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(ps) == 0 {
			return
		}
		srw := &saveResponseWriter{
			rw: &rw,
			b:  &bytes.Buffer{},
		}
		next.ServeHTTP(srw, r)
		token, err := idTokenFromBytes(ctx, srw.b.Bytes())
		if err != nil {
			// TODO: this doesn't work
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		if token.Email == "" {
			log.Error(ctx, "failed to find email claim in ID token", zap.Error(err))
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		for _, p := range ps {
			defer p.Close()
			if _, err := p.FindUser(ctx, token.Email); err != nil {
				if !errors.As(err, &errNotFound{}) {
					log.Error(ctx, "failed to find user", zap.Error(err),
						zap.String("user", token.Email))
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}
				if _, err := p.CreateUser(ctx, &User{name: token.Email}); err != nil {
					log.Error(ctx, "failed to provision user", zap.Error(err),
						zap.String("user", token.Email))
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
			var gs []*Group
			for _, g := range token.Groups {
				gs = append(gs, &Group{name: g})
			}
			for _, g := range token.Groups {
				if _, err := p.FindGroup(ctx, g); err != nil {
					if !errors.As(err, &errNotFound{}) {
						log.Error(ctx, "failed to find group", zap.Error(err),
							zap.String("group", g))
						rw.WriteHeader(http.StatusInternalServerError)
						return
					}
					if _, err := p.CreateGroup(ctx, &Group{name: g}); err != nil {
						log.Error(ctx, "failed to provision group", zap.Error(err),
							zap.String("group", g))
						rw.WriteHeader(http.StatusInternalServerError)
						return
					}
				}
				if err := p.SetUserGroups(ctx, &User{name: token.Email}, gs); err != nil {
					log.Error(ctx, "failed to set user groups", zap.Error(err),
						zap.String("user", token.Email),
						zap.Any("groups", token.Groups))
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}
	}
}

func idTokenFromBytes(ctx context.Context, b []byte) (*idToken, error) {
	resp := &tokenResp{}
	if err := json.Unmarshal(b, resp); err != nil {
		log.Error(ctx, "failed to unmarshal /token response", zap.Error(err))
		return nil, err
	}
	tok, err := b64.RawStdEncoding.DecodeString(jwtPayload(resp.IdToken))
	if err != nil {
		log.Error(ctx, "failed to base64 decode ID token", zap.Error(err),
			zap.String("token", resp.IdToken))
		return nil, err
	}
	token := &idToken{}
	if err := json.Unmarshal(tok, token); err != nil {
		log.Error(ctx, "failed to unmarshal /token response", zap.Error(err))
		return nil, err
	}
	return token, nil
}

type saveResponseWriter struct {
	b  *bytes.Buffer
	rw *http.ResponseWriter
}

func (sr *saveResponseWriter) Write(b []byte) (int, error) {
	if _, err := sr.b.Write(b); err != nil {
		return 0, err
	}
	w := *sr.rw
	return w.Write(b)
}

func (sr *saveResponseWriter) Header() http.Header {
	w := *sr.rw
	return w.Header()
}

func (sr *saveResponseWriter) WriteHeader(statusCode int) {
	w := *sr.rw
	w.WriteHeader(statusCode)
}

func (w *dexWeb) provisioners(ctx context.Context) ([]provisioner, error) {
	var ps []provisioner
	if w.env.Config.DeterminedURL != "" {
		d, err := NewDeterminedProvisioner(ctx, DeterminedConfig{
			MasterURL: w.env.Config.DeterminedURL,
			Username:  w.env.Config.DeterminedUsername,
			Password:  w.env.Config.DeterminedPassword,
			TLS:       w.env.Config.DeterminedTLS,
		})
		if err != nil {
			log.Error(ctx, "failed to find instantiate determined user provisioner", zap.Error(err))
			return nil, err
		}
		ps = append(ps, d)
	}
	return ps, nil
}

func jwtPayload(jwt string) string {
	parts := strings.Split(jwt, ".")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}
