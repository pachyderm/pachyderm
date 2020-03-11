local params = std.extVar("__ksonnet/params").components["workflows"];

local k = import 'k.libsonnet';
local workflows = import 'workflows.libsonnet';
local namespace = params.namespace;

local name = params.name;

local prowEnv = workflows.parseEnv(params.prow_env);
local bucket = params.bucket;

std.prune(k.core.v1.list.new([workflows.parts(namespace, name, params).e2e(prowEnv, bucket)]))
