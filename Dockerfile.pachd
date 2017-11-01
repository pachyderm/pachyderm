FROM scratch
LABEL maintainer="jdoliner@pachyderm.io"

ADD ./pachd /
ADD ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/pachd"]
