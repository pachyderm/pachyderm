#!/bin/bash -x

IMAGES=(
    "bash:4"
    "debian:buster-slim"
    "dockermuenster/caddy:0.9.3"
    "gcr.io/google_containers/kube-state-metrics:v0.5.0"
    "gcr.io/k8s-minikube/storage-provisioner:v1.8.1"
    "giantswarm/tiny-tools:latest"
    "golang:1.12.1"
    "golang:1.13.8"
    "golang:1.14.2"
    "golang:1.15.4"
    "grafana/grafana:4.2.0"
    "grafana/loki:2.0.0"
    "grafana/promtail:2.0.0"
    "k8s.gcr.io/coredns:1.6.7"
    "k8s.gcr.io/etcd:3.4.3-0"
    "k8s.gcr.io/kube-apiserver:v1.18.3"
    "k8s.gcr.io/kube-controller-manager:v1.18.3"
    "k8s.gcr.io/kube-proxy:v1.18.3"
    "k8s.gcr.io/kube-scheduler:v1.18.3"
    "k8s.gcr.io/pause:3.2"
    "pachyderm/dash:0.5.48"
    "pachyderm/etcd:v3.3.5"
    "pachyderm/grpc-proxy:0.4.10"
    "pachyderm/opencv:latest"
    "pachyderm/ubuntuplusnetcat:latest"
    "postgres:13.0-alpine"
    "prom/node-exporter:v0.14.0"
    "prom/prometheus:v1.7.0"
    "python:latest"
    "quay.io/prometheus/alertmanager:v0.7.1"
    "ubuntu:16.04"
    "ubuntu:18.04"
    "ubuntu:latest"
    "v4tech/imagemagick:latest"
)

mkdir -p /tmp/cache

for image in $(ls /tmp/cache); do 
    docker load -i /tmp/cache/$image
done

for f in ${IMAGES[@]}; do
    docker pull $f 
    docker save $f -o /tmp/cache/$(echo $f | sed 's|/|-|g')
done
