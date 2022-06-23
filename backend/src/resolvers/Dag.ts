import {RepoInfo, PipelineInfo} from '@pachyderm/node-pachyderm';
import flatMap from 'lodash/flatMap';
import keyBy from 'lodash/keyBy';
import uniqBy from 'lodash/uniqBy';
import uniqueId from 'lodash/uniqueId';
import objectHash from 'object-hash';

import flattenPipelineInput from '@dash-backend/lib/flattenPipelineInput';
import getJobsFromJobSet from '@dash-backend/lib/getJobsFromJobSet';
import {
  toGQLJobState,
  toGQLPipelineState,
} from '@dash-backend/lib/gqlEnumMappers';
import hasRepoReadPermissions from '@dash-backend/lib/hasRepoReadPermissions';
import {
  gqlJobStateToNodeState,
  gqlPipelineStateToNodeState,
} from '@dash-backend/lib/nodeStateMappers';
import sortJobInfos from '@dash-backend/lib/sortJobInfos';
import withSubscription from '@dash-backend/lib/withSubscription';
import {
  NodeType,
  QueryResolvers,
  SubscriptionResolvers,
  Vertex,
} from '@graphqlTypes';

interface DagResolver {
  Query: {
    dag: QueryResolvers['dag'];
  };
  Subscription: {
    dags: SubscriptionResolvers['dags'];
  };
}

const deriveVertices = (
  repos: RepoInfo.AsObject[],
  pipelines: PipelineInfo.AsObject[],
) => {
  const pipelineMap = keyBy(pipelines, (p) => p.pipeline?.name || '');
  const repoMap = keyBy(repos, (r) => r.repo?.name || '');

  const repoNodes = repos.map<Vertex>((r) => ({
    name: `${r.repo?.name}_repo`,
    type:
      r.repo && pipelineMap[r.repo.name]
        ? NodeType.OUTPUT_REPO
        : NodeType.INPUT_REPO,
    state: null,
    access: hasRepoReadPermissions(r.authInfo?.permissionsList),
    createdAt: r.created?.seconds,
    // detect out output repos as those name matching a pipeline
    parents: r.repo && pipelineMap[r.repo.name] ? [r.repo.name] : [],
  }));

  const pipelineNodes = flatMap(pipelines, (p) => {
    const pipelineName = p.pipeline?.name || '';
    const state = toGQLPipelineState(p.state);
    const jobState = toGQLJobState(p.lastJobState);

    const nodes: Vertex[] = [
      {
        name: pipelineName,
        type: NodeType.PIPELINE,
        state: gqlPipelineStateToNodeState(state),
        access: p.pipeline
          ? hasRepoReadPermissions(
              repoMap[p.pipeline.name].authInfo?.permissionsList,
            )
          : false,
        jobState,
        createdAt: p.details?.createdAt?.seconds,
        parents: p.details?.input ? flattenPipelineInput(p.details?.input) : [],
      },
    ];

    if (p.details?.egress) {
      nodes.push({
        name:
          p.details.egress.url ||
          p.details.egress.sqlDatabase?.url ||
          p.details.egress.objectStorage?.url ||
          '',
        type: NodeType.EGRESS,
        access: true,
        parents: [`${pipelineName}_repo`],
        state: gqlPipelineStateToNodeState(state),
        jobState,
        createdAt: p.details.createdAt?.seconds,
      });
    }

    return nodes;
  });

  return [...repoNodes, ...pipelineNodes];
};

const dagResolver: DagResolver = {
  Query: {
    dag: async (_field, {args: {projectId}}, {pachClient, log}) => {
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

      return deriveVertices(repos, pipelines);
    },
  },
  Subscription: {
    dags: {
      subscribe: (
        _field,
        {args: {projectId, jobSetId}},
        {pachClient, log, account},
      ) => {
        let prevDataHash = '';
        let data: Vertex[] = [];
        const getDags = async () => {
          if (jobSetId) {
            const jobSet = await pachClient
              .pps()
              .inspectJobSet({id: jobSetId, projectId, details: false});

            // stabilize elkjs inputs.
            // listRepo and listPipeline are both stabalized on 'name',
            // but jobsets come back in a stream with random order.
            const sortedJobSet = sortJobInfos(
              await getJobsFromJobSet({
                jobSet,
                projectId,
                pachClient,
              }),
            );

            const pipelineMap = keyBy(
              sortedJobSet.map((job) => job.job?.pipeline),
              (p) => p?.name || '',
            );

            const vertices = sortedJobSet.reduce<Vertex[]>((acc, job) => {
              const inputs = job.details?.input
                ? flattenPipelineInput(job.details.input)
                : [];

              const inputRepoVertices: Vertex[] = inputs.map((input) => ({
                parents: pipelineMap[input] ? [input] : [],
                type: pipelineMap[input]
                  ? NodeType.OUTPUT_REPO
                  : NodeType.INPUT_REPO,
                state: null,
                access: true,
                name: `${input}_repo`,
              }));

              const pipelineVertex: Vertex = {
                parents: inputs,
                type: NodeType.PIPELINE,
                access: true,
                name: job.job?.pipeline?.name || '',
                state: gqlJobStateToNodeState(toGQLJobState(job.state)),
              };

              const outputRepoVertex: Vertex = {
                parents: [job.job?.pipeline?.name || ''],
                type: NodeType.OUTPUT_REPO,
                access: true,
                state: null,
                name: `${job.job?.pipeline?.name || ''}_repo`,
              };

              return [
                ...acc,
                ...inputRepoVertices,
                outputRepoVertex,
                pipelineVertex,
              ];
            }, []);

            data = uniqBy(vertices, (v) => v.name);
          } else {
            const [repos, pipelines] = await Promise.all([
              pachClient.pfs().listRepo(),
              pachClient.pps().listPipeline(),
            ]);

            log.info({
              eventSource: 'dag resolver',
              event: 'deriving vertices for dag',
              meta: {projectId},
            });

            data = deriveVertices(repos, pipelines);
          }

          const dataHash = objectHash(data, {
            unorderedArrays: true,
          });

          if (prevDataHash === dataHash) {
            return;
          }

          prevDataHash = dataHash;

          return data;
        };

        const id = uniqueId();

        return {
          [Symbol.asyncIterator]: () =>
            withSubscription<Vertex[]>({
              triggerName: `${id}_DAGS_UPDATED${
                jobSetId ? `_JOB_${jobSetId}` : ''
              }`,
              resolver: getDags,
              intervalKey: id,
            }),
        };
      },
      resolve: (result: Vertex[]) => {
        return result;
      },
    },
  },
};

export default dagResolver;
