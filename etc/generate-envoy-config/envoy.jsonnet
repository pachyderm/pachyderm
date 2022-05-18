local Envoy = import 'envoy.libsonnet';
local Services = import 'pachyderm-services.libsonnet';

Envoy.bootstrap(
  listeners=[
    Envoy.httpListener(
      port=8080,  // Not port 80, because we run as user 101 and not root.
      name='proxy-http',
      // Everything except the metrics service is served on the multiplexed route.  The order of
      // services' routes is important!
      routes=std.flatMap(function(name) Services[name].routes, ['pachd-grpc', 'pachd-s3', 'pachd-identity', 'pachd-oidc', 'console'])
    ),
  ] + [
    local svc = Services[name];
    Envoy.httpListener(
      port=svc.internal_port,
      name='direct-' + name,
      routes=svc.routes,
    )
    // Every service gets a direct route.  We sort the keys so that the output is
    // stable across regeneration.
    for name in std.sort(std.objectFields(Services), keyF=function(name) Services[name].internal_port)
  ],
  clusters=[
    local svc = Services[name];
    (if 'grpc' in svc && svc.grpc then Envoy.GRPCCluster else Envoy.defaultCluster) + {
      name: name,
      load_assignment: Envoy.loadAssignment(name=name, address=svc.service, port=svc.internal_port),
      health_checks: if 'health_check' in svc then [svc.health_check] else [],
    }
    for name in std.objectFields(Services)
  ],
)
