FROM ubuntu:18.04

# Install Go
RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
		g++ \
		gcc \
		libc6-dev \
		make \
        curl \
	&& rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION 1.15.4
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOLANG_DOWNLOAD_SHA256 eb61005f0b932c93b424a3a4eaa67d72196c79129d9a3ea8578047683e2c80d5

RUN curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz \
	&& echo "$GOLANG_DOWNLOAD_SHA256  golang.tar.gz" | sha256sum -c - \
	&& tar -C /usr/local -xzf golang.tar.gz \
	&& rm golang.tar.gz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
WORKDIR $GOPATH

ADD map.go /
RUN go build -o /map /map.go 
ENTRYPOINT ["/map"]
