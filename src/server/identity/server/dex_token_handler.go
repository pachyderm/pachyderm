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
		srw := &saveResponseWriter{
			rw: &rw,
			b:  &bytes.Buffer{},
		}
		next.ServeHTTP(srw, r)
		ps, err := w.provisioners(ctx)
		if err != nil {
			log.Error(ctx, "failed to collect user provisioners", zap.Error(err))
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(ps) == 0 {
			return
		}
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
			u := &User{name: token.Email}
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
			for _, grp := range token.Groups {
				g, err := p.FindGroup(ctx, grp)
				if err != nil {
					if !errors.As(err, &errNotFound{}) {
						log.Error(ctx, "failed to find group", zap.Error(err),
							zap.String("group", grp))
						rw.WriteHeader(http.StatusInternalServerError)
						return
					}
					if _, err := p.CreateGroup(ctx, &Group{name: grp}); err != nil {
						log.Error(ctx, "failed to provision group", zap.Error(err),
							zap.String("group", grp))
						rw.WriteHeader(http.StatusInternalServerError)
						return
					}
				}
				gs = append(gs, g)
			}
			if err := p.SetUserGroups(ctx, u, gs); err != nil {
				log.Error(ctx, "failed to set user groups", zap.Error(err),
					zap.String("user", token.Email),
					zap.Any("groups", token.Groups))
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}

func idTokenFromBytes(ctx context.Context, b []byte) (*idToken, error) {
	resp := &tokenResp{}
	if err := json.Unmarshal(b, resp); err != nil {
		msg := "failed to unmarshal /token response"
		log.Error(ctx, msg, zap.Error(err))
		return nil, errors.Wrap(err, msg)
	}
	tok, err := b64.RawStdEncoding.DecodeString(jwtPayload(resp.IdToken))
	if err != nil {
		msg := "failed to base64 decode ID token"
		log.Error(ctx, msg, zap.Error(err),
			zap.String("token", resp.IdToken))
		return nil, errors.Wrap(err, msg)
	}
	token := &idToken{}
	if err := json.Unmarshal(tok, token); err != nil {
		msg := "failed to unmarshal /token response"
		log.Error(ctx, msg, zap.Error(err))
		return nil, errors.Wrap(err, msg)
	}
	return token, nil
}

type saveResponseWriter struct {
	b  *bytes.Buffer
	rw *http.ResponseWriter
}

func (sr *saveResponseWriter) Write(b []byte) (int, error) {
	if _, err := sr.b.Write(b); err != nil {
		return 0, errors.Wrap(err, "write bytes to save response writer buffer")
	}
	w := *sr.rw
	i, err := w.Write(b)
	return i, errors.Wrap(err, "write bytes to response writer")
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
