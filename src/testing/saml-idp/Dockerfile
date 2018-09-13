# Build and statically-link our IdP server in alpine, which uses musl instead
# of glibc, so that the IdP server can be built into a completely static binary
# even though it calls getaddrinfo() (via Go's "net" library)
FROM golang:1.10.2-alpine AS build
LABEL maintainer="msteffen@pachyderm.io"

RUN ["apk", "update"]
RUN ["apk", "add", "git", "gcc", "libc-dev"]
RUN ["mkdir", "-p", "/go/src/saml-idp"]
ADD ["./idp.go", "/go/src/saml-idp"]
RUN ["go", "get", "github.com/crewjam/saml"]
RUN ["go", "get", "github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"]
RUN ["go", "install", "-ldflags", "-linkmode external -extldflags -static", "-a", "saml-idp"]

# Copy statically-built IdP server into a new image (based on scratch)
FROM scratch
COPY --from=build ["/go/bin/saml-idp", "/"]
CMD [ "/saml-idp" ]
