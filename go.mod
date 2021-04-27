module github.com/pachyderm/pachyderm/v2

go 1.15

require (
	cloud.google.com/go/bigtable v1.1.0 // indirect
	cloud.google.com/go/storage v1.6.0
	github.com/Azure/azure-sdk-for-go v36.1.0+incompatible
	github.com/Masterminds/squirrel v0.0.0-20161115235646-20f192218cf5 // indirect
	github.com/aws/aws-lambda-go v1.13.3
	github.com/aws/aws-sdk-go v1.27.0
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/bmatcuk/doublestar v1.2.2 // indirect
	github.com/c-bata/go-prompt v0.2.3
	github.com/c2h5oh/datasize v0.0.0-20200112174442-28bbd4740fee // indirect
	github.com/cevaris/ordered_map v0.0.0-20190319150403-3adeae072e73
	github.com/chmduquesne/rollinghash v4.0.0+incompatible
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/etcd v3.3.13+incompatible
	github.com/coreos/go-etcd v2.0.0+incompatible
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f
	github.com/crewjam/saml v0.4.4-0.20201214083806-0dd2422c212e
	github.com/dexidp/dex v0.0.0-20201118094123-6ca0cbc85759
	github.com/dexidp/dex/api/v2 v2.0.0
	github.com/dlclark/regexp2 v1.2.0 // indirect
	github.com/docker/go-metrics v0.0.0-20181218153428-b84716841b82 // indirect
	github.com/docker/go-plugins-helpers v0.0.0-20181025120712-1e6269c305b8 // indirect
	github.com/docker/go-units v0.4.0
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/elazarl/goproxy v0.0.0-20191011121108-aa519ddbe484 // indirect
	github.com/evanphx/json-patch v4.9.0+incompatible
	github.com/fatih/camelcase v1.0.0
	github.com/fatih/color v1.9.0
	github.com/fluent/fluent-bit-go v0.0.0-20190925192703-ea13c021720c // indirect
	github.com/frankban/quicktest v1.7.2 // indirect
	github.com/fsouza/go-dockerclient v1.4.1
	github.com/gocql/gocql v0.0.0-20200121121104-95d072f1b5bb // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/golang-migrate/migrate/v4 v4.7.0 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e
	github.com/golang/protobuf v1.4.3
	github.com/gomodule/redigo v2.0.0+incompatible // indirect
	github.com/gorilla/mux v1.7.4
	github.com/grafana/loki v1.0.2
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.1-0.20191002090509-6af20e3a5340 // indirect
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645 // indirect
	github.com/hanwen/go-fuse/v2 v2.0.3
	github.com/hashicorp/go-retryablehttp v0.5.4 // indirect
	github.com/hashicorp/golang-lru v0.5.3
	github.com/hashicorp/vault v1.1.3
	github.com/influxdata/go-syslog/v2 v2.0.1 // indirect
	github.com/itchyny/gojq v0.11.2
	github.com/jackc/pgerrcode v0.0.0-20201024163028-a0d42d470451
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/jmoiron/sqlx v1.2.0
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/juju/ansiterm v0.0.0-20180109212912-720a0952cc2a
	github.com/julienschmidt/httprouter v1.3.0
	github.com/klauspost/compress v1.9.4 // indirect
	github.com/lann/builder v0.0.0-20150808151131-f22ce00fd939 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/lib/pq v1.3.0
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/mattn/go-isatty v0.0.12
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/mattn/go-tty v0.0.3 // indirect
	github.com/minio/minio-go/v6 v6.0.55
	github.com/ncw/swift v1.0.50 // indirect
	github.com/onsi/gomega v1.8.1 // indirect
	github.com/opentracing-contrib/go-grpc v0.0.0-20180928155321-4b5a12d3ff02
	github.com/opentracing/opentracing-go v1.1.1-0.20200124165624-2876d2018785
	github.com/pachyderm/ohmyglob v0.0.0-20210308211843-d5b47775fc36
	github.com/pachyderm/s2 v0.0.0-20200609183354-d52f35094520
	github.com/pierrec/lz4 v2.5.3-0.20200429092203-e876bbd321b3+incompatible // indirect
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/pkg/errors v0.9.1
	github.com/pkg/term v0.0.0-20190109203006-aa71e9d9e942 // indirect
	github.com/prometheus/client_golang v1.5.0
	github.com/prometheus/common v0.9.1
	github.com/prometheus/prometheus v1.8.2-0.20200213233353-b90be6f32a33 // indirect
	github.com/rafaeljusto/redigomock v0.0.0-20190202135759-257e089e14a1 // indirect
	github.com/robfig/cron v1.2.0
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/segmentio/analytics-go v0.0.0-20160426181448-2d840d861c32
	github.com/segmentio/backo-go v0.0.0-20160424052352-204274ad699c // indirect
	github.com/segmentio/fasthash v0.0.0-20180216231524-a72b379d632e // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v0.0.6-0.20191202130430-b04b5bfc50cb
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	github.com/thanos-io/thanos v0.11.0 // indirect
	github.com/tonistiigi/fifo v0.0.0-20190226154929-a9fb20d87448 // indirect
	github.com/uber/jaeger-client-go v2.20.1+incompatible
	github.com/ugorji/go v1.1.7 // indirect
	github.com/vbauerster/mpb/v6 v6.0.2
	github.com/weaveworks/common v0.0.0-20200429090833-ac38719f57dd // indirect
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c // indirect
	go.etcd.io/etcd v0.0.0-20200401174654-e694b7bb0875 // indirect
	go.uber.org/automaxprocs v1.4.0
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5 // indirect
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/net v0.0.0-20210224082022-3d97a244fca7
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	golang.org/x/sys v0.0.0-20210421221651-33663a62ff08
	golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d
	google.golang.org/api v0.20.0
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/grpc v1.29.1
	gopkg.in/go-playground/webhooks.v5 v5.11.0
	gopkg.in/pachyderm/yaml.v3 v3.0.0-20200130061037-1dd3d7bd0850
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	gopkg.in/src-d/go-git.v4 v4.12.0
	gopkg.in/yaml.v2 v2.4.0
	helm.sh/helm/v3 v3.1.2
	honnef.co/go/tools v0.1.3 // indirect
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/cli-runtime v0.17.4
	k8s.io/client-go v0.21.0
	modernc.org/mathutil v1.0.0
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/yaml v1.2.0
)

// Docker library versioning is not straightforward, see https://github.com/moby/moby/issues/39302
// For the moment, the windows build requires a fix that has not been tagged with an official release
replace github.com/docker/docker => github.com/docker/docker v1.4.2-0.20191213113251-3452f136aa68

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible

//replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190718183610-8e956561bbf5

replace github.com/sercand/kuberesolver => github.com/sercand/kuberesolver v1.0.1-0.20200204133151-f60278fd3dac
