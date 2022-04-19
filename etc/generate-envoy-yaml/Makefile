../helm/pachyderm/envoy.json: envoy.jsonnet envoy.libsonnet
	jsonnet envoy.jsonnet > ../helm/pachyderm/envoy.json

format: envoy.jsonnet envoy.libsonnet
	jsonnetfmt -i *.jsonnet *.libsonnet

check: ../helm/pachyderm/envoy.json
	envoy -c ../helm/pachyderm/envoy.json --mode validate

.PHONY: check format
