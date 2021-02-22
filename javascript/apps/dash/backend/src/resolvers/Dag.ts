import {RepoInfo} from '@pachyderm/proto/pb/pfs/pfs_pb';
import {Input, PipelineInfo} from '@pachyderm/proto/pb/pps/pps_pb';
import flatten from 'lodash/flatten';
import flattenDeep from 'lodash/flattenDeep';
import keyBy from 'lodash/keyBy';

import {NodeType, QueryResolvers} from 'generated/types';
import client from 'grpc/client';
import disconnectedComponents from 'lib/disconnectedComponents';
import {LinkInputData, Vertex} from 'lib/types';

interface DagResolver {
  Query: {
    dag: QueryResolvers['dag'];
    dags: QueryResolvers['dags'];
  };
}

const flattenPipelineInput = (input: Input.AsObject): string[] => {
  const result = [];

  // TODO: Deal with git and cron input types
  if (input.pfs) {
    result.push(`${input.pfs.repo}_repo`);
  }

  // TODO: Update to indicate which elemets are crossed, unioned, joined, grouped, with.
  return flattenDeep([
    result,
    input.crossList.map((i) => flattenPipelineInput(i)),
    input.unionList.map((i) => flattenPipelineInput(i)),
    input.joinList.map((i) => flattenPipelineInput(i)),
    input.groupList.map((i) => flattenPipelineInput(i)),
  ]);
};

const deriveVertices = (
  repos: RepoInfo.AsObject[],
  pipelines: PipelineInfo.AsObject[],
) => {
  const pipelineMap = keyBy(pipelines, (p) => p.pipeline?.name);

  const repoNodes = repos.map((r) => ({
    name: `${r.repo?.name}_repo`,
    type: NodeType.Repo,
    state: null,
    access: true,
    // detect out output repos as those name matching a pipeline
    parents: r.repo && pipelineMap[r.repo.name] ? [r.repo.name] : [],
  }));

  const pipelineNodes = pipelines.map((p) => ({
    name: p.pipeline?.name || '',
    type: NodeType.Pipeline,
    state: p.state,
    access: true,
    jobState: p.lastJobState,
    parents: p.input ? flattenPipelineInput(p.input) : [],
  }));

  return [...repoNodes, ...pipelineNodes];
};

const normalizeDAGData = (vertices: Vertex[]) => {
  // Calculate the indicies of the nodes in the DAG, as
  // the "link" object requires the source & target attributes
  // to reference the index from the "nodes" list. This prevents
  // us from needing to run "indexOf" on each individual node
  // when building the list of links for the client.
  const correspondingIndex: {[key: string]: number} = vertices.reduce(
    (acc, val, index) => ({
      ...acc,
      [val.name]: index,
    }),
    {},
  );

  const allLinks = vertices.map((node) =>
    node.parents.reduce((acc, parentName) => {
      const source = correspondingIndex[parentName || ''];
      if (Number.isInteger(source)) {
        return [
          ...acc,
          {
            source: source,
            target: correspondingIndex[node.name],
            state: vertices[source].jobState,
          },
        ];
      }
      return acc;
    }, [] as LinkInputData[]),
  );

  const dataNodes = vertices.map((node) => ({
    name: node.name,
    type: node.type,
    state: node.state,
    access: true,
  }));

  return {
    nodes: dataNodes,
    links: flatten(allLinks),
  };
};

const dagResolver: DagResolver = {
  Query: {
    dag: async (
      _field,
      {args: {projectId}},
      {pachdAddress = '', authToken = ''},
    ) => {
      const pachClient = client(pachdAddress, authToken);

      // TODO: Error handling
      const [repos, pipelines] = await Promise.all([
        pachClient.pfs().listRepo(projectId),
        pachClient.pps().listPipeline(projectId),
      ]);

      const allVertices = deriveVertices(repos, pipelines);

      return normalizeDAGData(allVertices);
    },
    dags: async (
      _field,
      {args: {projectId}},
      {pachdAddress = '', authToken = ''},
    ) => {
      const pachClient = client(pachdAddress, authToken);

      // TODO: Error handling
      const [repos, pipelines] = await Promise.all([
        pachClient.pfs().listRepo(projectId),
        pachClient.pps().listPipeline(projectId),
      ]);

      const allVertices = deriveVertices(repos, pipelines);

      const components = disconnectedComponents(allVertices);

      return components.map((component) => normalizeDAGData(component));
    },
  },
};

export default dagResolver;
