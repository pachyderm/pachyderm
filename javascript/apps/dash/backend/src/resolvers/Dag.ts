import {RepoInfo} from '@pachyderm/proto/pb/pfs/pfs_pb';
import {Input, PipelineInfo} from '@pachyderm/proto/pb/pps/pps_pb';
import ELK from 'elkjs/lib/elk.bundled.js';
import flatten from 'lodash/flatten';
import flattenDeep from 'lodash/flattenDeep';
import keyBy from 'lodash/keyBy';

import client from '@dash-backend/grpc/client';
import disconnectedComponents from '@dash-backend/lib/disconnectedComponents';
import {LinkInputData, NodeInputData, Vertex} from '@dash-backend/lib/types';
import {Link, Node, NodeType, QueryResolvers} from '@graphqlTypes';

interface DagResolver {
  Query: {
    dag: QueryResolvers['dag'];
    dags: QueryResolvers['dags'];
  };
}

const elk = new ELK();

const flattenPipelineInput = (input: Input.AsObject): string[] => {
  const result = [];

  if (input.pfs) {
    result.push(`${input.pfs.repo}_repo`);
  }

  if (input.cron) {
    result.push(`${input.cron.repo}_repo`);
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

const normalizeDAGData = async (
  vertices: Vertex[],
  nodeWidth: number,
  nodeHeight: number,
) => {
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

  // create elk edges
  const elkEdges = flatten(
    vertices.map((node) =>
      node.parents.reduce<LinkInputData[]>((acc, parentName) => {
        const source = correspondingIndex[parentName || ''];
        if (Number.isInteger(source)) {
          return [
            ...acc,
            {
              id: `${parentName}-${node.name}`,
              sources: [parentName],
              targets: [node.name],
              state: vertices[source].jobState,
              sourceState: vertices[source].state,
              targetstate: node.state,
              sections: [],
            },
          ];
        }
        return acc;
      }, []),
    ),
  );

  // create elk children
  const elkChildren = vertices.map<NodeInputData>((node) => ({
    id: node.name,
    name: node.name,
    type: node.type,
    state: node.state,
    access: true,
    x: 0,
    y: 0,
    width: nodeWidth,
    height: nodeHeight,
  }));

  await elk.layout(
    {id: 'root', children: elkChildren, edges: elkEdges},
    {
      layoutOptions: {
        'org.eclipse.elk.algorithm': 'layered',
        'org.eclipse.elk.mergeEdges': 'false',
        'org.eclipse.elk.direction': 'RIGHT',
        'org.eclipse.elk.layered.layering.strategy': 'INTERACTIVE',
        'org.eclipse.elk.edgeRouting': 'ORTHOGONAL',
        'org.eclipse.elk.layered.spacing.edgeNodeBetweenLayers': '20',
        'org.eclipse.elk.layered.spacing.nodeNodeBetweenLayers': '40',
      },
    },
  );

  // convert elk edges to graphl Link
  const links = elkEdges.map<Link>((edge) => {
    const {
      sections,
      sources,
      targets,
      junctionPoints,
      labels,
      layoutOptions,
      ...rest
    } = edge;
    return {
      ...rest,
      source: sources[0],
      target: targets[0],
      startPoint: sections[0].startPoint,
      endPoint: sections[0].endPoint,
      bendPoints: sections[0].bendPoints || [],
    };
  });

  //convert elk children to graphl Node
  const nodes = elkChildren.map<Node>((node) => {
    return {
      id: node.id,
      x: node.x || 0,
      y: node.y || 0,
      name: node.name,
      type: node.type,
      state: node.state,
      access: node.access,
    };
  });

  return {
    nodes,
    links,
  };
};

const dagResolver: DagResolver = {
  Query: {
    dag: async (
      _field,
      {args: {projectId, nodeHeight, nodeWidth}},
      {pachdAddress = '', authToken = '', log},
    ) => {
      const pachClient = client({pachdAddress, authToken, projectId, log});

      // TODO: Error handling
      const [repos, pipelines] = await Promise.all([
        pachClient.pfs().listRepo(),
        pachClient.pps().listPipeline(),
      ]);

      log.info({
        eventSource: 'dag resolver',
        event: 'deriving vertices for dag',
        meta: {projectId},
      });
      const allVertices = deriveVertices(repos, pipelines);

      return normalizeDAGData(allVertices, nodeWidth, nodeHeight);
    },
    dags: async (
      _field,
      {args: {projectId, nodeHeight, nodeWidth}},
      {pachdAddress = '', authToken = '', log},
    ) => {
      const pachClient = client({pachdAddress, authToken, projectId, log});

      // TODO: Error handling
      const [repos, pipelines] = await Promise.all([
        pachClient.pfs().listRepo(),
        pachClient.pps().listPipeline(),
      ]);

      log.info({
        eventSource: 'dag resolver',
        event: 'deriving vertices for dag',
        meta: {projectId},
      });
      const allVertices = deriveVertices(repos, pipelines);

      log.info({
        eventSource: 'dag resolver',
        event: 'discovering disconnected components',
        meta: {projectId},
      });
      const components = disconnectedComponents(allVertices);

      return components.map(async (component) =>
        normalizeDAGData(component, nodeWidth, nodeHeight),
      );
    },
  },
};

export default dagResolver;
