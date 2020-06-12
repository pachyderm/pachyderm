FROM scratch
WORKDIR /app
COPY --from=pachyderm_build /app/pachd .
COPY --from=pachyderm_build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/app/pachd"]
