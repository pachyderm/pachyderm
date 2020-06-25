FROM scratch
COPY --from=pachyderm_build /app/pachd /pachd
COPY --from=pachyderm_build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/pachd"]
