// Services that we want to proxy; where they are and how to route them.
local services = {
  'pachd-grpc': {
    internal_port: 1650,
    external_port: 30650,
    grpc: true,
    service: 'pachd-proxy-backend',
    routes: [
      {
        match: {
          grpc: {},
          prefix: '/',
        },
        route: {
          cluster: 'pachd-grpc',
          timeout: '604800s',
        },
      },
    ],
    health_check: {
      grpc_health_check: {},
      healthy_threshold: 1,
      interval: '10s',
      timeout: '10s',
      unhealthy_threshold: 2,
      no_traffic_interval: '10s',
      no_traffic_healthy_interval: '10s',
    },
  },
  'pachd-s3': {
    internal_port: 1600,
    external_port: 30600,
    service: 'pachd-proxy-backend',
    routes: local base = {
      route: {
        cluster: 'pachd-s3',
        idle_timeout: '600s',
        timeout: '604800s',
      },
    }; [
      base {  // S3v4
        match: {
          prefix: '/',
          headers: [
            {
              name: 'authorization',
              string_match: {
                prefix: 'AWS4-HMAC-SHA256',
              },
            },
          ],
        },
      },
      base {  // S3v2
        match: {
          prefix: '/',
          headers: [
            {
              name: 'authorization',
              string_match: {
                prefix: 'AWS ',
              },
            },
          ],
        },
      },
    ],
  },
  'pachd-identity': {
    internal_port: 1658,
    external_port: 30658,
    service: 'pachd-proxy-backend',
    routes: [
      {
        match: {
          prefix: '/dex',
        },
        route: {
          cluster: 'pachd-identity',
          idle_timeout: '60s',
          timeout: '60s',
        },
      },
    ],
    health_check: {
      healthy_threshold: 1,
      http_health_check: {
        host: 'localhost',
        path: '/dex/.well-known/openid-configuration',
      },
      interval: '30s',
      timeout: '10s',
      unhealthy_threshold: 2,
      no_traffic_interval: '10s',
      no_traffic_healthy_interval: '10s',
    },
  },
  'pachd-oidc': {
    internal_port: 1657,
    external_port: 30657,
    service: 'pachd-proxy-backend',
    routes: [
      {
        match: {
          prefix: '/authorization-code/callback',
        },
        route: {
          cluster: 'pachd-oidc',
          idle_timeout: '60s',
          timeout: '60s',
        },
      },
    ],
  },
  console: {
    internal_port: 4000,
    external_port: 4000,
    service: 'console-proxy-backend',
    routes: [
      {
        match: {
          prefix: '/',
        },
        route: {
          cluster: 'console',
          idle_timeout: '600s',
          timeout: '3600s',
          upgrade_configs: [
            {
              enabled: true,
              upgrade_type: 'websocket',
            },
          ],
        },
      },
    ],
    health_check: {
      healthy_threshold: 1,
      http_health_check: {
        host: 'localhost',  // This is just the value of the Host: header, not something to connect to.
        path: '/',
      },
      interval: '30s',
      timeout: '10s',
      unhealthy_threshold: 2,
      no_traffic_interval: '10s',
      no_traffic_healthy_interval: '10s',
    },
  },
  'pachd-metrics': {
    internal_port: 1656,
    external_port: 30656,
    service: 'pachd-proxy-backend',
    routes: [
      {
        match: {
          prefix: '/',
        },
        route: {
          cluster: 'pachd-metrics',
          idle_timeout: '60s',
          timeout: '60s',
        },
      },
    ],
  },
};

// This is the generation code.
local Envoy = import 'envoy.libsonnet';
Envoy.bootstrap(
  listeners=[
    Envoy.httpListener(
      port=80,
      name='proxy-http',
      routes=std.flatMap(function(name) services[name].routes, ['pachd-grpc', 'pachd-s3', 'pachd-identity', 'pachd-oidc', 'console'])
    ),
  ] + [
    local svc = services[name];
    Envoy.httpListener(
      port=svc.internal_port,
      name='direct-' + name,
      routes=svc.routes,
    )
    for name in ['pachd-grpc', 'pachd-s3', 'pachd-identity', 'pachd-oidc', 'console', 'pachd-metrics']
  ],
  clusters=[
    local svc = services[name];
    (if 'grpc' in svc && svc.grpc then Envoy.GRPCCluster else Envoy.defaultCluster) + {
      name: name,
      load_assignment: Envoy.loadAssignment(name=name, address=svc.service, port=svc.internal_port),
      health_checks: if 'health_check' in svc then [svc.health_check] else [],
    }
    for name in std.objectFields(services)
  ],
)
