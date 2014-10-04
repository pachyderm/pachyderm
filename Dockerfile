FROM golang
RUN go get github.com/coreos/go-etcd/etcd
ADD . /go/src/pfs
RUN go install pfs/services/master
RUN go install pfs/services/slave
RUN go install pfs/services/router
EXPOSE 80
