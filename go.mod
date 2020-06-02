module github.com/pachyderm/pachyderm

go 1.13

require (
	cloud.google.com/go v0.49.0
	cloud.google.com/go/storage v1.3.0
	github.com/Azure/azure-sdk-for-go v36.1.0+incompatible
	github.com/OneOfOne/xxhash v1.2.6
	github.com/aws/aws-lambda-go v1.13.3
	github.com/aws/aws-sdk-go v1.27.0
	github.com/beevik/etree v1.1.0
	github.com/brianvoe/gofakeit v3.18.0+incompatible
	github.com/c-bata/go-prompt v0.2.3
	github.com/cevaris/ordered_map v0.0.0-20190319150403-3adeae072e73
	github.com/cheggaaa/pb/v3 v3.0.4
	github.com/chmduquesne/rollinghash v4.0.0+incompatible
	github.com/coreos/bbolt v1.3.3
	github.com/coreos/etcd v3.3.13+incompatible
	github.com/coreos/go-etcd v2.0.0+incompatible
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f
	github.com/crewjam/saml v0.0.0-20190521120225-344d075952c9
	github.com/dlclark/regexp2 v1.2.0 // indirect
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
	github.com/gofrs/uuid v3.3.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9
	github.com/golang/protobuf v1.3.3
	github.com/golang/snappy v0.0.1
	github.com/google/go-github v17.0.0+incompatible
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/grafana/loki v1.5.0
	github.com/hanwen/go-fuse/v2 v2.0.2
	github.com/hashicorp/go-hclog v0.8.0 // indirect
	github.com/hashicorp/go-plugin v1.0.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.5.4 // indirect
	github.com/hashicorp/golang-lru v0.5.3
	github.com/hashicorp/vault v1.1.3
	github.com/hashicorp/yamux v0.0.0-20181012175058-2f1d1f20f75d // indirect
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/jinzhu/gorm v1.9.12
	github.com/juju/ansiterm v0.0.0-20180109212912-720a0952cc2a
	github.com/julienschmidt/httprouter v1.3.0
	github.com/lib/pq v1.3.0
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/mattn/go-tty v0.0.3 // indirect
	github.com/minio/minio-go v6.0.14+incompatible
	github.com/onsi/gomega v1.8.1 // indirect
	github.com/opentracing-contrib/go-grpc v0.0.0-20180928155321-4b5a12d3ff02
	github.com/opentracing/opentracing-go v1.1.1-0.20200124165624-2876d2018785
	github.com/pachyderm/ohmyglob v0.0.0-20190713004043-630e5c15d4e4
	github.com/pachyderm/s2 v0.0.0-20200528231500-590b33e3c716
	github.com/pkg/errors v0.9.1
	github.com/pkg/term v0.0.0-20190109203006-aa71e9d9e942 // indirect
	github.com/prometheus/client_golang v1.5.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.9.1
	github.com/robfig/cron v1.2.0
	github.com/russellhaering/goxmldsig v0.0.0-20180430223755-7acd5e4a6ef7 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/segmentio/analytics-go v0.0.0-20160426181448-2d840d861c32
	github.com/segmentio/backo-go v0.0.0-20160424052352-204274ad699c // indirect
	github.com/segmentio/kafka-go v0.2.4
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v0.0.6-0.20191202130430-b04b5bfc50cb
	github.com/spf13/pflag v1.0.5
	github.com/uber-go/atomic v1.4.0 // indirect
	github.com/uber/jaeger-client-go v2.20.1+incompatible
	github.com/willf/bitset v1.1.10 // indirect
	github.com/willf/bloom v2.0.3+incompatible
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c // indirect
	golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b
	golang.org/x/oauth2 v0.0.0-20191202225959-858c2ad4c8b6
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20200523222454-059865788121
	golang.org/x/tools v0.0.0-20200601175630-2caf76543d99 // indirect
	google.golang.org/api v0.14.0
	google.golang.org/grpc v1.27.0
	gopkg.in/go-playground/webhooks.v5 v5.11.0
	gopkg.in/pachyderm/yaml.v3 v3.0.0-20200130061037-1dd3d7bd0850
	gopkg.in/square/go-jose.v2 v2.3.1 // indirect
	gopkg.in/src-d/go-git.v4 v4.12.0
	gopkg.in/urfave/cli.v1 v1.20.0 // indirect
	helm.sh/helm/v3 v3.1.2
	honnef.co/go/tools v0.0.1-2020.1.4 // indirect
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/cli-runtime v0.17.4
	k8s.io/client-go v12.0.0+incompatible
	modernc.org/mathutil v1.0.0
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/yaml v1.1.0
)

// Holy shit, the docker library versions are a clusterfuck, see https://github.com/moby/moby/issues/39302
// For the moment, the windows build requires a fix that has not been tagged with an official release
replace github.com/docker/docker => github.com/docker/docker v1.4.2-0.20191213113251-3452f136aa68

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0
