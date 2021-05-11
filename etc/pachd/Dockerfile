FROM scratch

MAINTAINER jdoliner@pachyerm.io

LABEL name="Pachyderm" \
      vendor="Pachyderm"

COPY --from=pachyderm_build /licenses /
COPY --from=pachyderm_build /tmp/to-copy /

COPY --from=pachyderm_build /app/pachd /pachd
COPY --from=pachyderm_build /app/dex-assets /dex-assets
COPY --from=pachyderm_build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/pachd"]
