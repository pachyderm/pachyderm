local cluster = import 'cluster.jsonnet';
local project = import 'project.jsonnet';
local spec = import 'pipeline.json';
local validator = import 'validator.jsonnetlib';

local effective = std.mergePatch(std.mergePatch(cluster, project), spec);

effective + validator
