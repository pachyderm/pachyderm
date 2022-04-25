{
  bootstrap(listeners, clusters): {
    admin: {
      access_log: [
        {
          name: 'envoy.access_loggers.stderr',
          typed_config: {
            '@type': 'type.googleapis.com/envoy.extensions.access_loggers.stream.v3.StderrAccessLog',
          },
        },
      ],
      address: {
        socket_address: {
          address: '0.0.0.0',
          port_value: 9901,
        },
      },
    },
    static_resources: {
      clusters: clusters,
      listeners: listeners,
    },
    overload_manager: {
      actions: [
        {
          name: 'envoy.overload_actions.shrink_heap',
          triggers: [
            {
              name: 'envoy.resource_monitors.fixed_heap',
              threshold: {
                value: 0.95,
              },
            },
          ],
        },
        {
          name: 'envoy.overload_actions.stop_accepting_requests',
          triggers: [
            {
              name: 'envoy.resource_monitors.fixed_heap',
              threshold: {
                value: 0.98,
              },
            },
          ],
        },
      ],
      refresh_interval: '0.25s',
      resource_monitors: [
        {
          name: 'envoy.resource_monitors.fixed_heap',
          typed_config: {
            '@type': 'type.googleapis.com/envoy.extensions.resource_monitors.fixed_heap.v3.FixedHeapConfig',
            max_heap_size_bytes: 500000000,
          },
        },
      ],
    },
    layered_runtime: {
      layers: [
        {
          name: 'static_layer_0',
          static_layer: {
            overload: {
              global_downstream_max_connections: 50000,
            },
          },
        },
      ],
    },
  },

  loadAssignment(name, address, port): {
    cluster_name: name,
    endpoints: [
      {
        lb_endpoints: [
          {
            endpoint: {
              address: {
                socket_address: {
                  address: address,
                  port_value: port,
                },
              },
            },
          },
        ],
      },
    ],
  },

  httpListener(port, name, routes): {
    name: name,
    per_connection_buffer_limit_bytes: 32768,
    traffic_direction: 'INBOUND',
    address: {
      socket_address: {
        address: '0.0.0.0',
        port_value: port,
      },
    },
    filter_chains: [
      {
        filters: [
          {
            name: 'envoy.http_connection_manager',
            typed_config: {
              '@type': 'type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager',
              access_log: [
                {
                  name: 'envoy.access_loggers.stdout',
                  typed_config: {
                    '@type': 'type.googleapis.com/envoy.extensions.access_loggers.stream.v3.StdoutAccessLog',
                  },
                },
              ],
              codec_type: 'auto',
              common_http_protocol_options: {
                headers_with_underscores_action: 'REJECT_REQUEST',
                idle_timeout: '60s',
              },
              http2_protocol_options: {
                initial_connection_window_size: 1048576,
                initial_stream_window_size: 65536,
                max_concurrent_streams: 100,
              },
              http_filters: std.prune([
                // If there's a GRPC router, enable grpc_stats.
                if std.foldl(function(last, route) last || 'grpc' in route.match, routes, false) then
                  {
                    name: 'envoy.filters.http.grpc_stats',
                    typed_config: {
                      '@type': 'type.googleapis.com/envoy.extensions.filters.http.grpc_stats.v3.FilterConfig',
                      enable_upstream_stats: true,
                      // Stats for all methods = true would let anyone on the Internet fill up the
                      // statistics buffer with bogus RPC method names, so we have to leave it off.
                      stats_for_all_methods: false,
                    },
                  } else {},
                {
                  name: 'envoy.filters.http.router',
                  typed_config: {
                    '@type': 'type.googleapis.com/envoy.extensions.filters.http.router.v3.Router',
                  },
                },
              ]),
              http_protocol_options: {
                accept_http_10: false,
              },
              request_timeout: '60s',
              route_config: {
                virtual_hosts: [
                  {
                    domains: [
                      '*',
                    ],
                    name: 'any',
                    routes: routes,
                  },
                ],
              },
              stat_prefix: name,
              stream_idle_timeout: '0s',  // disable stream idle timeouts, for waiting on jobs, tailing logs, etc.
              use_remote_address: true,
            },
          },
        ],
      },
    ],
  },
  defaultCluster: {
    connect_timeout: '10s',
    dns_lookup_family: 'V4_ONLY',
    lb_policy: 'round_robin',
    type: 'strict_dns',
    upstream_connection_options: {
      tcp_keepalive: {},
    },
  },
  GRPCCluster: self.defaultCluster {
    typed_extension_protocol_options: {
      'envoy.extensions.upstreams.http.v3.HttpProtocolOptions': {
        '@type': 'type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions',
        explicit_http_config: {
          http2_protocol_options: {},
        },
      },
    },
  },
}
