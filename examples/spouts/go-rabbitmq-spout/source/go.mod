module github.com/pachyderm/pachyderm/examples/spouts/go-rabbitmq-spout/source

go 1.16

require (
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/pachyderm/pachyderm/v2 v2.0.0-alpha.23
	github.com/streadway/amqp v1.0.0
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43 // indirect
	golang.org/x/text v0.3.4 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20201111145450-ac7456db90a6 // indirect
	google.golang.org/grpc v1.33.2 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
)

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190718183610-8e956561bbf5

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible
