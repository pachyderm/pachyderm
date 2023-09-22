local Envoy = import 'envoy.libsonnet';
local Services = import 'pachyderm-services.libsonnet';

Envoy.bootstrap(
  listeners=[
    // Serve the multiplexed route table on port 8080.
    Envoy.httpListener(
      port=8080,  // Not port 80, because we run as user 101 and not root.
      name='proxy-http',
      // Everything except the metrics service is served on the multiplexed route.  The order of
      // services' routes is important!
      routes=std.flatMap(function(name) Services[name].routes, ['pachd-grpc', 'pachd-s3', 'pachd-identity', 'pachd-oidc', 'pachd-http', 'pachd-restgateway', 'console'])
    ),

    // In case someone port-forwards port 8443 because they were expecting TLS to be working, this
    // will tell them why it isn't.
    Envoy.httpListener(
      port=8443,
      name='https-warning',
      routes=[
        Envoy.messageRoute(500, 'This is the cleartext installation of pachyderm-proxy; enable TLS in the helm chart and reinstall.'),
      ],
    ),
  ] + [
    local svc = Services[name];
    Envoy.httpListener(
      port=svc.internal_port,
      name='direct-' + name,
      routes=svc.routes,
    )
    // Every service gets a direct route.  Whether or not traffic from the Internet can reach these
    // is controlled by helm values (the service won't bind the ports if the user doesn't want to
    // use them).  We sort the keys here so that the output is stable across regeneration.
    for name in std.sort(std.objectFields(Services), keyF=function(name) Services[name].internal_port)
  ],
  clusters=[
    Envoy.serviceAsCluster(name, Services[name])
    for name in std.objectFields(Services)
  ],
)
