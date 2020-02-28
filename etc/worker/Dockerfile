FROM debian:buster
WORKDIR /pach
COPY --from=pachyderm_build /app/worker .
COPY --from=pachyderm_build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY worker.sh .
