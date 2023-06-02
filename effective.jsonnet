local cluster = import 'cluster.jsonnet';
local project = import 'project.jsonnet';
local spec = import 'pipeline.json';

std.mergePatch(std.mergePatch(cluster, project), spec)
