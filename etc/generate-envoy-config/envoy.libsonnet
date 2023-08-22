{
  accessLogFormat:: {
    omit_empty_values: true,
    json_format: {
      timestamp: '%START_TIME%',
      severity: 'info',
      message: 'http response',
      method: '%REQ(:method)%',
      path: '%REQ(X-ENVOY-ORIGINAL-PATH?:PATH)%',
      protocol: '%PROTOCOL%',
      upstream_protocol: '%UPSTREAM_PROTOCOL%',
      response_code: '%RESPONSE_CODE%',
      response_code_details: '%RESPONSE_CODE_DETAILS%',
      connection_termination_details: '%CONNECTION_TERMINATION_DETAILS%',
      grpc_status_code: '%GRPC_STATUS(SNAKE_STRING)%',
      grpc_message: '%RESP(GRPC-MESSAGE):64%',
      response_flags: '%RESPONSE_FLAGS%',
      rx_bytes: '%BYTES_RECEIVED%',
      tx_bytes: '%BYTES_SENT%',
      request_headers_bytes: '%REQUEST_HEADERS_BYTES%',
      duration_ms: '%DURATION%',
      upstream_service_time_ms: '%RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)%',
      xff: '%REQ(X-FORWARDED-FOR)%',
      downstream_remote_address: '%DOWNSTREAM_DIRECT_REMOTE_ADDRESS_WITHOUT_PORT%',
      user_agent: '%REQ(USER-AGENT)%',
      connection_id: '%CONNECTION_ID%',
      'x-request-id': '%REQ(X-REQUEST-ID)%',
      stream_id: '%STREAM_ID%',
      authority: '%REQ(authority)%',
      upstream_host: '%UPSTREAM_HOST%',
      route: '%ROUTE_NAME%',
      upstream_local_address: '%UPSTREAM_LOCAL_ADDRESS%',
      pachctl_command: '%REQ(COMMAND):64%',
      request_attempt_count: '%UPSTREAM_REQUEST_ATTEMPT_COUNT%',
      upgrade: '%RESP(UPGRADE)%',
      upstream: '%UPSTREAM_CLUSTER%',
    },
  },

  bootstrap(listeners, clusters): {
    admin: {
      access_log: [
        {
          name: 'envoy.access_loggers.stdout',
          filter: {
            or_filter: {
              filters: [
                {
                  or_filter: {
                    filters: [
                      {
                        status_code_filter: {
                          comparison: {
                            op: 'le',
                            value: {
                              default_value: 199,
                              runtime_key: 'invalid.ignore_kubeprobe_response_code',  // Required by the API spec, but unused in this configuration.
                            },
                          },
                        },
                      },
                      {
                        status_code_filter: {
                          comparison: {
                            op: 'ge',
                            value: {
                              default_value: 300,
                              runtime_key: 'invalid.ignore_kubeprobe_response_code',  // Required by the API spec, but unused in this configuration.
                            },
                          },
                        },
                      },
                    ],
                  },
                },
                {
                  header_filter: {
                    header: {
                      name: 'user-agent',
                      string_match: {
                        prefix: 'kube-probe/',
                      },
                      invert_match: true,
                    },
                  },
                },
              ],
            },
          },
          typed_config: {
            '@type': 'type.googleapis.com/envoy.extensions.access_loggers.stream.v3.StderrAccessLog',
            log_format: $.accessLogFormat { json_format+: { message: 'admin response' } },
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

  loadAssignment(name, address, port):: {
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

  httpConnectionManager(name, routes=[], response_headers_to_add={}):: {
    name: 'envoy.http_connection_manager',
    typed_config: {
      '@type': 'type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager',
      access_log: [
        {
          name: 'envoy.access_loggers.stdout',
          typed_config: {
            '@type': 'type.googleapis.com/envoy.extensions.access_loggers.stream.v3.StdoutAccessLog',
            log_format: $.accessLogFormat,
          },
        },
      ],
      codec_type: 'auto',
      common_http_protocol_options: {
        headers_with_underscores_action: 'REJECT_REQUEST',
        idle_timeout: '86400s',
      },
      http2_protocol_options: {
        initial_connection_window_size: 1048576,
        initial_stream_window_size: 524288,
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
      request_timeout: '604800s',  // Necessary to allow long file uploads.
      stream_idle_timeout: '86400s',  // Only completely idle streams are dropped after this timeout.
      route_config: {
        [if std.length(response_headers_to_add) > 0 then 'response_headers_to_add' else null]: [
          {
            header: {
              key: key,
              value: response_headers_to_add[key],
            },
          }
          for key in std.objectFields(response_headers_to_add)
        ],
        virtual_hosts: [
          {
            domains: ['*'],
            name: 'any',
            routes: routes,
            retry_policy: {
              retry_on: 'connect-failure',
              num_retries: 4,
              host_selection_retry_max_attempts: 4,
            },
          },
        ],
      },
      stat_prefix: name,
      use_remote_address: true,
    },
  },

  redirectToHttpsRoute: {
    match: {
      prefix: '/',
    },
    redirect: {
      https_redirect: true,
    },
  },

  messageRoute(status, message): {
    match: {
      prefix: '/',
    },
    direct_response: {
      status: status,
      body: {
        inline_string: message,
      },
    },
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
        filters: [$.httpConnectionManager(name, routes)],
      },
    ],
    access_log: [
      {
        name: 'envoy.access_loggers.stdout',
        typed_config: {
          '@type': 'type.googleapis.com/envoy.extensions.access_loggers.stream.v3.StderrAccessLog',
          log_format: $.accessLogFormat { json_format+: { listener: name, message: 'listener log' } },
        },
      },
    ],
  },

  httpsListener(port, name, routes): {
    name: name,
    per_connection_buffer_limit_bytes: 32768,
    traffic_direction: 'INBOUND',
    address: {
      socket_address: {
        address: '0.0.0.0',
        port_value: port,
      },
    },
    listener_filters: [
      {
        name: 'envoy.filters.listener.tls_inspector',
        typed_config: {
          '@type': 'type.googleapis.com/envoy.extensions.filters.listener.tls_inspector.v3.TlsInspector',
        },
      },
    ],
    filter_chains: [
      {
        filter_chain_match: {
          transport_protocol: 'tls',
        },
        transport_socket: {
          name: 'envoy.transport_sockets.tls',
          typed_config: {
            '@type': 'type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.DownstreamTlsContext',
            common_tls_context: {
              alpn_protocols: ['h2', 'http/1.1'],
              tls_params: {
                tls_minimum_protocol_version: 'TLSv1_2',
                cipher_suites: [
                  '[ECDHE-ECDSA-AES128-GCM-SHA256|ECDHE-ECDSA-CHACHA20-POLY1305]',
                  '[ECDHE-RSA-AES128-GCM-SHA256|ECDHE-RSA-CHACHA20-POLY1305]',
                  'ECDHE-ECDSA-AES256-GCM-SHA384',
                  'ECDHE-RSA-AES256-GCM-SHA384',
                ],
              },
              tls_certificate_sds_secret_configs: [
                {
                  name: 'tls',
                  sds_config: {
                    resource_api_version: 'V3',
                    path_config_source: {
                      path: '/etc/envoy/sds.yaml',
                      watched_directory: {
                        path: '/etc/envoy',
                      },
                    },
                  },
                },
              ],
            },
          },
        },
        filters: [$.httpConnectionManager(name, routes, response_headers_to_add={ 'strict-transport-security': 'max-age=604800' })],
      },

      // Redirect to https if this request wasn't sent over TLS.
      {
        filter_chain_match: {
          transport_protocol: 'raw_buffer',
        },
        filters: [$.httpConnectionManager(name + '-cleartext', [$.redirectToHttpsRoute])],
      },
    ],
  },

  defaultCluster:: {
    connect_timeout: '10s',
    dns_lookup_family: 'V4_ONLY',
    dns_refresh_rate: '1s',
    dns_failure_refresh_rate: {
      base_interval: '0.1s',
      max_interval: '1s',
    },
    lb_policy: 'random',
    type: 'strict_dns',
    upstream_connection_options: {
      tcp_keepalive: {},
    },
    common_lb_config: {
      healthy_panic_threshold: {
        value: 0.0,
      },
    },
  },

  defaultGRPCCluster:: self.defaultCluster {
    typed_extension_protocol_options: {
      'envoy.extensions.upstreams.http.v3.HttpProtocolOptions': {
        '@type': 'type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions',
        explicit_http_config: {
          http2_protocol_options: {},
        },
      },
    },
  },

  serviceAsCluster(name, service):
    ((if 'grpc' in service && service.grpc then $.defaultGRPCCluster else $.defaultCluster) + {
       name: name,
       load_assignment: $.loadAssignment(name=name, address=service.service, port=service.internal_port),
       health_checks: if 'health_check' in service then [service.health_check] else [],
     }),
}
