local Envoy = import 'envoy.libsonnet';
local Services = import 'pachyderm-services.libsonnet';

Envoy.bootstrap(listeners=[], clusters=[])
