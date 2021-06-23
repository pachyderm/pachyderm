import {RepoInfo} from '@pachyderm/proto/pb/pfs/pfs_pb';
import {
  CronInput,
  Input,
  PFSInput,
  PipelineInfo,
} from '@pachyderm/proto/pb/pps/pps_pb';
import ELK from 'elkjs/lib/elk.bundled.js';
import flatMap from 'lodash/flatMap';
import flatten from 'lodash/flatten';
import flattenDeep from 'lodash/flattenDeep';
import keyBy from 'lodash/keyBy';
import minBy from 'lodash/minBy';
import objectHash from 'object-hash';

import disconnectedComponents from '@dash-backend/lib/disconnectedComponents';
import {
  toGQLJobState,
  toGQLPipelineState,
} from '@dash-backend/lib/gqlEnumMappers';
import hasRepoReadPermissions from '@dash-backend/lib/hasRepoReadPermissions';
import {
  EgressVertex,
  LinkInputData,
  NodeInputData,
  PipelineVertex,
  RepoVertex,
} from '@dash-backend/lib/types';
import withSubscription from '@dash-backend/lib/withSubscription';
import {
  Dag,
  Link,
  Node,
  NodeType,
  QueryResolvers,
  SubscriptionResolvers,
  DagDirection,
  JobState,
} from '@graphqlTypes';

interface DagResolver {
  Query: {
    dag: QueryResolvers['dag'];
  };
  Subscription: {
    dags: SubscriptionResolvers['dags'];
  };
}

const elk = new ELK();

const flattenPipelineInput = (
  input: Input.AsObject,
): (
  | Omit<PFSInput.AsObject, 'commit'>
  | Omit<CronInput.AsObject, 'commit'>
)[] => {
  const result = [];

  if (input.pfs) {
    const {commit, ...rest} = input.pfs;
    result.push(rest);
  }

  if (input.cron) {
    const {commit, ...rest} = input.cron;

    result.push(rest);
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
  const pipelineMap = keyBy(pipelines, (p) => p.pipeline?.name || '');
  const repoMap = keyBy(repos, (r) => r.repo?.name || '');

  const repoNodes = repos.map<RepoVertex>((r) => ({
    name: `${r.repo?.name}_repo`,
    type:
      r.repo && pipelineMap[r.repo.name]
        ? NodeType.OUTPUT_REPO
        : NodeType.INPUT_REPO,
    state: null,
    access: hasRepoReadPermissions(r.authInfo?.permissionsList),
    created: r.created,
    // detect out output repos as those name matching a pipeline
    parents: r.repo && pipelineMap[r.repo.name] ? [r.repo.name] : [],
  }));

  const pipelineNodes = flatMap(pipelines, (p) => {
    const pipelineName = p.pipeline?.name || '';
    const state = toGQLPipelineState(p.state);
    const jobState = toGQLJobState(p.lastJobState);

    const nodes: (PipelineVertex | EgressVertex)[] = [
      {
        name: pipelineName,
        type: NodeType.PIPELINE,
        state,
        access: p.pipeline
          ? hasRepoReadPermissions(
              repoMap[p.pipeline.name].authInfo?.permissionsList,
            )
          : false,
        jobState,
        parents: p.details?.input ? flattenPipelineInput(p.details?.input) : [],
      },
    ];

    if (p.details?.egress) {
      nodes.push({
        name: p.details.egress.url,
        type: NodeType.EGRESS,
        access: true,
        parents: [`${pipelineName}_repo`],
        state,
        jobState,
      });
    }

    return nodes;
  });

  return [...repoNodes, ...pipelineNodes];
};

interface DeriveTransferringStateOpts {
  targetNodeType: NodeType;
  state?: JobState;
  targetNodeState?: JobState;
}

const deriveTransferringState = ({
  targetNodeType,
  state,
  targetNodeState,
}: DeriveTransferringStateOpts) => {
  // Should display transferring _from_ a corresponding egressing pipeline's output repo.
  // Because repos do not have a "state", we need to use the target node's state instead.
  return (
    (targetNodeType === NodeType.EGRESS &&
      targetNodeState === JobState.JOB_EGRESSING) ||
    (targetNodeType !== NodeType.EGRESS && state === JobState.JOB_RUNNING)
  );
};

const normalizeDAGData = async (
  vertices: (EgressVertex | PipelineVertex | RepoVertex)[],
  nodeWidth: number,
  nodeHeight: number,
  direction: DagDirection = DagDirection.RIGHT,
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
    vertices.map<LinkInputData[]>((node) => {
      if (node.type === NodeType.PIPELINE) {
        return node.parents.reduce<LinkInputData[]>((acc, parent) => {
          const parentName = `${parent.repo}_repo`;
          const sourceIndex = correspondingIndex[parentName];
          const state = vertices[sourceIndex].jobState;

          return [
            ...acc,
            {
              id: objectHash({node: node, ...parent}),
              sources: [parentName],
              targets: [node.name],
              state,
              sourceState: vertices[sourceIndex].state,
              targetstate: node.state,
              sections: [],
              transferring: deriveTransferringState({
                targetNodeType: node.type,
                state,
                targetNodeState: node.jobState,
              }),
            },
          ];
        }, []);
      }

      return node.parents.reduce<LinkInputData[]>((acc, parentName) => {
        const sourceIndex = correspondingIndex[parentName || ''];
        const state = vertices[sourceIndex].jobState;

        return [
          ...acc,
          {
            // hashing here for consistency
            id: objectHash(`${parentName}-${node.name}`),
            sources: [parentName],
            targets: [node.name],
            state,
            sourceState: vertices[sourceIndex].state,
            targetstate: node.state,
            sections: [],
            transferring: deriveTransferringState({
              targetNodeType: node.type,
              state,
              targetNodeState: node.jobState,
            }),
          },
        ];
      }, []);
    }),
  );

  // create elk children
  const elkChildren = vertices.map<NodeInputData>((node) => ({
    id: node.name,
    name: node.name,
    type: node.type,
    state: node.state,
    access: node.access,
    x: 0,
    y: 0,
    width: nodeWidth,
    height: nodeHeight,
  }));

  const horizontal = direction === DagDirection.RIGHT;

  await elk.layout(
    {id: 'root', children: elkChildren, edges: elkEdges},
    {
      layoutOptions: {
        'org.eclipse.elk.algorithm': 'disco',
        'org.eclipse.elk.aspectRatio': horizontal ? '0.2' : '10',
        'org.eclipse.elk.mergeEdges': 'false',
        'org.eclipse.elk.direction': direction,
        'org.eclipse.elk.layered.layering.strategy': 'INTERACTIVE',
        'org.eclipse.elk.disco.componentCompaction.componentLayoutAlgorithm':
          'layered',
        'spacing.componentComponent': horizontal ? '75' : '150',

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
      {args: {projectId, nodeHeight, nodeWidth, direction}},
      {pachClient, log},
    ) => {
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

      const id = minBy(repos, (r) => r.created?.seconds)?.repo?.name || '';

      const normalizedData = await normalizeDAGData(
        allVertices,
        nodeWidth,
        nodeHeight,
      );

      return {
        id,
        nodes: normalizedData.nodes,
        links: normalizedData.links,
      };
    },
  },
  Subscription: {
    dags: {
      subscribe: (
        _field,
        {args: {projectId, nodeHeight, nodeWidth, direction}},
        {pachClient, log, account},
      ) => {
        let prevReposHash = '';
        let prevPipelinesHash = '';

        const getDags = async () => {
          // TODO: Error handling
          const [repos, pipelines] = await Promise.all([
            pachClient.pfs().listRepo(),
            pachClient.pps().listPipeline(),
          ]);

          const reposHash = objectHash(repos, {unorderedArrays: true});
          const pipelinesHash = objectHash(pipelines, {unorderedArrays: true});

          if (
            prevReposHash === reposHash &&
            prevPipelinesHash === pipelinesHash
          ) {
            return;
          }

          prevReposHash = reposHash;
          prevPipelinesHash = pipelinesHash;

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

          const normalizedData = await normalizeDAGData(
            allVertices,
            nodeWidth,
            nodeHeight,
            direction,
          );

          const components = disconnectedComponents(
            normalizedData.nodes,
            normalizedData.links,
          );

          return components.map((component) => {
            const componentRepos = repos.filter((repo) =>
              component.nodes.find(
                (c) =>
                  (c.type === NodeType.OUTPUT_REPO ||
                    c.type === NodeType.INPUT_REPO) &&
                  c.name === `${repo.repo?.name}_repo`,
              ),
            );
            const id =
              minBy(componentRepos, (r) => r.created?.seconds)?.repo?.name ||
              '';

            return {
              id,
              nodes: component.nodes,
              links: component.links,
            };
          });
        };

        return withSubscription<Dag[]>({
          triggerNames: [
            `${account.id}_DAGS_${direction}_UPDATED`,
            `DAGS_${direction}_UPDATED`,
          ],
          resolver: getDags,
          intervalKey: `${projectId}-${direction}`,
        });
      },
      resolve: (result: Dag[]) => {
        return result;
      },
    },
  },
};

export default dagResolver;
