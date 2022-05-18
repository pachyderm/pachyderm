../helm/pachyderm/envoy.json: envoy.jsonnet envoy.libsonnet
	jsonnet envoy.jsonnet > ../helm/pachyderm/envoy.json

format: envoy.jsonnet envoy.libsonnet
	jsonnetfmt -i *.jsonnet *.libsonnet

# This requires that you install Envoy, which is a big pain.  I just "docker cp" it out of the
# docker release, but there is a Homebrew package and a Debian package:
# https://www.envoyproxy.io/docs/envoy/latest/start/install
check-envoy-config: ../helm/pachyderm/envoy.json
	envoy -c ../helm/pachyderm/envoy.json --mode validate

# So that CI can fail when you forget to run jsonnetfmt.
check-formatting:
	jsonnetfmt --test *.jsonnet *.libsonnet
	echo "jsonnetfmt ok"

# So that CI can fail if you modify envoy.jsonnet but don't regenerate envoy.json.
check-generated:
	jsonnet envoy.jsonnet > /tmp/test-formatting.envoy.json
	diff -u ../helm/pachyderm/envoy.json /tmp/test-formatting.envoy.json  --label in-repo --label would-generate
	echo "codegen ok"

test: check-formatting check-generated check-envoy-config

.PHONY: format check-envoy-config check-formatting check-generated test
