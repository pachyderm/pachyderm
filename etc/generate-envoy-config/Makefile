all: ../helm/pachyderm/envoy.json ../helm/pachyderm/envoy-tls.json

../helm/pachyderm/envoy.json: envoy.jsonnet envoy.libsonnet pachyderm-services.libsonnet
	jsonnet envoy.jsonnet > ../helm/pachyderm/envoy.json

../helm/pachyderm/envoy-tls.json: envoy-tls.jsonnet envoy.libsonnet pachyderm-services.libsonnet
	jsonnet envoy-tls.jsonnet > ../helm/pachyderm/envoy-tls.json

format: *.jsonnet *.libsonnet
	jsonnetfmt -i *.jsonnet *.libsonnet

# This requires that you install Envoy, which is a big pain.  I just "docker cp" it out of the
# docker release, but there is a Homebrew package and a Debian package:
# https://www.envoyproxy.io/docs/envoy/latest/start/install
check-envoy-config: ../helm/pachyderm/envoy.json
	envoy -c ../helm/pachyderm/envoy.json --mode validate

# This requires that the CI machine or your workstation look a little like the Kubernetes pod we run
# in; ../helm/pachyderm/envoy-sds.yaml needs to be copied to /etc/envoy/sds.yaml and you need to
# have TLS certificates (empty files ok) in /tls.  If you want to make requests to Envoy (that you
# start up like the validation command below, but without `--mode validate`), they need to be real
# certs.  I got mine out of here: https://go.dev/src/net/http/internal/testcert/testcert.go.
check-envoy-config-tls: ../helm/pachyderm/envoy-tls.json
	mkdir -p /etc/envoy
	mkdir -p /tls
	test -e /tls/tls.key || touch /tls/tls.key
	test -e /tls/tls.crt || touch /tls/tls.crt
	test -e /etc/envoy/sds.yaml || cp ../helm/pachyderm/envoy-sds.yaml /etc/envoy/sds.yaml
	envoy -c ../helm/pachyderm/envoy-tls.json --mode validate --service-node localhost --service-zone localhost --service-cluster envoy

# So that CI can fail when you forget to run jsonnetfmt.
check-formatting:
	jsonnetfmt --test *.jsonnet *.libsonnet
	echo "jsonnetfmt ok"

# So that CI can fail if you modify envoy.jsonnet but don't regenerate envoy.json.
check-generated:
	jsonnet envoy.jsonnet > /tmp/test-formatting.envoy.json
	diff -u ../helm/pachyderm/envoy.json /tmp/test-formatting.envoy.json  --label in-repo --label would-generate
	echo "codegen ok"

check-generated-tls:
	jsonnet envoy-tls.jsonnet > /tmp/test-formatting.envoy.json
	diff -u ../helm/pachyderm/envoy-tls.json /tmp/test-formatting.envoy.json  --label in-repo --label would-generate
	echo "tls codegen ok"

test: check-formatting check-generated check-generated-tls check-envoy-config check-envoy-config-tls

.PHONY: all format check-envoy-config check-envoy-config-tls check-formatting check-generated check-generated-tls test
