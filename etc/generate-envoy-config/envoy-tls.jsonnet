local Envoy = import 'envoy.libsonnet';
local Services = import 'pachyderm-services.libsonnet';

Envoy.bootstrap(
  listeners=[
    // The HTTP port just redirects to HTTPS.
    Envoy.httpListener(
      port=8080,
      name='https-redirect',
      routes=[Envoy.redirectToHttpsRoute],
    ),

    // The HTTPS port serves the multiplexed route table with TLS.  (Additionally, if a cleartext
    // request reaches this port, we generate a redirect to the https protocol.)  Note: the order of
    // the routes we select below is crucial to the multiplexing working.
    Envoy.httpsListener(
      port=8443,
      name='proxy-https',
      routes=std.flatMap(function(name) Services[name].routes, ['pachd-grpc', 'pachd-s3', 'pachd-identity', 'pachd-oidc', 'pachd-http', 'console'])
    ),
  ],
  clusters=[
    Envoy.serviceAsCluster(name, Services[name])
    for name in std.objectFields(Services)
  ],
)
