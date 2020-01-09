module github.com/pachyderm/pachyderm

go 1.13

require (
	cloud.google.com/go v0.40.0
	github.com/Azure/azure-sdk-for-go v32.4.0+incompatible
	github.com/Azure/go-autorest/autorest/to v0.3.0 // indirect
	github.com/LK4D4/joincontext v0.0.0-20171026170139-1724345da6d5
	github.com/Microsoft/hcsshim v0.8.7 // indirect
	github.com/OneOfOne/xxhash v1.2.5
	github.com/aws/aws-lambda-go v1.11.1
	github.com/aws/aws-sdk-go v1.20.3
	github.com/beevik/etree v1.1.0
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/brianvoe/gofakeit v3.18.0+incompatible
	github.com/c-bata/go-prompt v0.2.3
	github.com/cevaris/ordered_map v0.0.0-20190319150403-3adeae072e73
	github.com/chmduquesne/rollinghash v4.0.0+incompatible
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/containerd/containerd v1.3.2 // indirect
	github.com/coreos/bbolt v1.3.3
	github.com/coreos/etcd v3.3.13+incompatible
	github.com/coreos/go-etcd v2.0.0+incompatible
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f
	github.com/crewjam/saml v0.0.0-20190521120225-344d075952c9
	github.com/dlclark/regexp2 v1.2.0 // indirect
	github.com/dnaeon/go-vcr v1.0.1 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.4.2-0.20191213113251-3452f136aa68 // indirect
	github.com/docker/go-units v0.4.0
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/elazarl/goproxy v0.0.0-20191011121108-aa519ddbe484 // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/fatih/camelcase v1.0.0
	github.com/fatih/color v1.7.0
	github.com/fatih/structs v1.1.0 // indirect
	github.com/fsouza/go-dockerclient v1.4.1
	github.com/go-ini/ini v1.42.0 // indirect
	github.com/go-test/deep v1.0.1 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef
	github.com/golang/protobuf v1.3.2
	github.com/golang/snappy v0.0.1
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gophercloud/gophercloud v0.2.0 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.12.1 // indirect
	github.com/hanwen/go-fuse v0.0.0-20180522155540-291273cb8ce0
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/hashicorp/go-hclog v0.8.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.1.0 // indirect
	github.com/hashicorp/go-plugin v1.0.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.5.4 // indirect
	github.com/hashicorp/go-rootcerts v1.0.1 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/go-version v1.1.0 // indirect
	github.com/hashicorp/golang-lru v0.5.1
	github.com/hashicorp/vault v1.1.3
	github.com/hashicorp/yamux v0.0.0-20181012175058-2f1d1f20f75d // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/juju/ansiterm v0.0.0-20180109212912-720a0952cc2a
	github.com/julienschmidt/httprouter v1.2.0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/mattn/go-runewidth v0.0.7 // indirect
	github.com/mattn/go-tty v0.0.3 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/minio/minio-go v6.0.14+incompatible
	github.com/mitchellh/go-testing-interface v1.0.0 // indirect
	github.com/onsi/ginkgo v1.10.2 // indirect
	github.com/onsi/gomega v1.8.1 // indirect
	github.com/opentracing-contrib/go-grpc v0.0.0-20180928155321-4b5a12d3ff02
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pachyderm/ohmyglob v0.0.0-20190713004043-630e5c15d4e4
	github.com/pachyderm/s2 v0.0.0-20191119172829-5e460c076ab6
	github.com/pkg/term v0.0.0-20190109203006-aa71e9d9e942 // indirect
	github.com/prometheus/client_golang v1.2.1
	github.com/prometheus/common v0.7.0
	github.com/remyoudompheng/bigfft v0.0.0-20190728182440-6a916e37a237 // indirect
	github.com/robfig/cron v1.2.0
	github.com/russellhaering/goxmldsig v0.0.0-20180430223755-7acd5e4a6ef7 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/segmentio/analytics-go v0.0.0-20160426181448-2d840d861c32
	github.com/segmentio/backo-go v0.0.0-20160424052352-204274ad699c // indirect
	github.com/segmentio/kafka-go v0.2.4
	github.com/sirupsen/logrus v1.4.2
	github.com/smartystreets/goconvey v0.0.0-20190330032615-68dc04aab96a // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/spf13/cobra v0.0.6-0.20191202130430-b04b5bfc50cb
	github.com/spf13/pflag v1.0.5
	github.com/uber-go/atomic v1.4.0 // indirect
	github.com/uber/jaeger-client-go v2.16.0+incompatible
	github.com/uber/jaeger-lib v2.0.0+incompatible // indirect
	github.com/willf/bitset v1.1.10 // indirect
	github.com/willf/bloom v2.0.3+incompatible
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c // indirect
	go.etcd.io/bbolt v1.3.3 // indirect
	go.uber.org/multierr v1.4.0 // indirect
	go.uber.org/zap v1.12.0 // indirect
	golang.org/x/crypto v0.0.0-20191206172530-e9b2fee46413
	golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f // indirect
	golang.org/x/net v0.0.0-20191209160850-c0dbc17a3553
	golang.org/x/oauth2 v0.0.0-20191202225959-858c2ad4c8b6
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20191210023423-ac6580df4449 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	golang.org/x/tools v0.0.0-20200107184032-11e9d9cc0042 // indirect
	google.golang.org/api v0.6.0
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/grpc v1.24.0
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/go-playground/webhooks.v5 v5.11.0
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.42.0 // indirect
	gopkg.in/pachyderm/yaml.v3 v3.0.0-20191204210125-74de94ae3b7a
	gopkg.in/square/go-jose.v2 v2.3.1 // indirect
	gopkg.in/src-d/go-git.v4 v4.12.0
	gopkg.in/yaml.v2 v2.2.7 // indirect
	k8s.io/api v0.0.0-20190816222004-e3a6b8045b0b
	k8s.io/apimachinery v0.0.0-20190816221834-a9f1d8a9c101
	k8s.io/client-go v11.0.1-0.20190918222721-c0e3722d5cf0+incompatible
	k8s.io/utils v0.0.0-20191114200735-6ca3b61696b6 // indirect
	modernc.org/mathutil v1.0.0
	sigs.k8s.io/yaml v1.1.0 // indirect
)

// Holy shit, the docker library versions are a clusterfuck, see https://github.com/moby/moby/issues/39302
// For the moment, the windows build requires a fix that has not been tagged with an official release
replace github.com/docker/docker => github.com/docker/docker v1.4.2-0.20191213113251-3452f136aa68
