import flatMap from 'lodash/flatMap';
import keyBy from 'lodash/keyBy';
import uniqBy from 'lodash/uniqBy';
import uniqueId from 'lodash/uniqueId';
import objectHash from 'object-hash';

import flattenPipelineInput, {
  flattenPipelineInputObj,
} from '@dash-backend/lib/flattenPipelineInput';
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
import {RepoInfo, PipelineInfo} from '@dash-backend/proto';
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
  const pipelineMap = keyBy(
    pipelines,
    (p) => `${p.pipeline?.project?.name || ''}_${p.pipeline?.name}`,
  );
  const repoMap = keyBy(
    repos,
    (r) => `${r.repo?.project?.name || ''}_${r.repo?.name}`,
  );

  const repoNodes = repos.map<Vertex>((r) => ({
    id: `${r.repo?.project?.name || ''}_${r.repo?.name}_repo`,
    name: r.repo?.name || '',
    type:
      r.repo && pipelineMap[`${r.repo?.project?.name || ''}_${r.repo?.name}`]
        ? NodeType.OUTPUT_REPO
        : NodeType.INPUT_REPO,
    state: null,
    jobState: null,
    access: hasRepoReadPermissions(r.authInfo?.permissionsList),
    createdAt: r.created?.seconds,
    // detect out output repos as those name matching a pipeline
    parents:
      r.repo && pipelineMap[`${r.repo?.project?.name || ''}_${r.repo?.name}`]
        ? [`${r.repo?.project?.name || ''}_${r.repo?.name}`]
        : [],
  }));

  const pipelineNodes = flatMap(pipelines, (p) => {
    const pipelineId = `${p.pipeline?.project?.name || ''}_${p.pipeline?.name}`;
    const state = toGQLPipelineState(p.state);
    const jobState = toGQLJobState(p.lastJobState);

    const nodes: Vertex[] = [
      {
        id: pipelineId,
        name: p.pipeline?.name || '',
        type: NodeType.PIPELINE,
        state: gqlPipelineStateToNodeState(state),
        access: p.pipeline
          ? hasRepoReadPermissions(
              repoMap[`${p.pipeline?.project?.name || ''}_${p.pipeline?.name}`]
                ?.authInfo?.permissionsList,
            )
          : false,
        jobState: gqlJobStateToNodeState(jobState),
        createdAt: p.details?.createdAt?.seconds,
        parents: p.details?.input ? flattenPipelineInput(p.details?.input) : [],
      },
    ];

    if (
      p.details?.egress &&
      (p.details.egress.url ||
        p.details.egress.sqlDatabase?.url ||
        p.details.egress.objectStorage?.url)
    ) {
      nodes.push({
        id: `${p.pipeline?.project?.name || ''}_${
          p.details.egress.url ||
          p.details.egress.sqlDatabase?.url ||
          p.details.egress.objectStorage?.url ||
          ''
        }`,
        name:
          p.details.egress.url ||
          p.details.egress.sqlDatabase?.url ||
          p.details.egress.objectStorage?.url ||
          '',
        type: NodeType.EGRESS,
        access: true,
        parents: [`${pipelineId}_repo`],
        state: null,
        jobState: null,
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
        pachClient.pfs().listRepo({projectIds: [projectId]}),
        pachClient.pps().listPipeline({projectIds: [projectId]}),
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
      subscribe: (_field, {args: {projectId, jobSetId}}, {pachClient, log}) => {
        let prevDataHash = '';
        let data: Vertex[] = [];
        const getDags = async () => {
          if (jobSetId) {
            // if given a global ID, we want to construct a DAG using only
            // nodes referred to by jobs sharing that global id
            const jobSet = await pachClient
              .pps()
              .inspectJobSet({id: jobSetId, projectId, details: false});

            // stabilize elkjs inputs.
            // listRepo and listPipeline are both stabilized on 'name',
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
              (p) => `${p?.project?.name || ''}_${p?.name}`,
            );

            const vertices = sortedJobSet.reduce<Vertex[]>((acc, job) => {
              const inputs = job.details?.input
                ? flattenPipelineInputObj(job.details.input).filter(
                    (input) => input.project === projectId,
                  )
                : [];

              const inputRepoVertices: Vertex[] = inputs.map(
                ({project, repo}) => ({
                  id: `${project || ''}_${repo}_repo`,
                  parents: [],
                  type: pipelineMap[`${project || ''}_${repo}`]
                    ? NodeType.OUTPUT_REPO
                    : NodeType.INPUT_REPO,
                  state: null,
                  jobState: null,
                  access: true,
                  name: `${repo}_repo`,
                }),
              );

              const pipelineVertex: Vertex = {
                id: `${job.job?.pipeline?.project?.name || ''}_${
                  job.job?.pipeline?.name
                }`,
                parents: inputs.map(
                  ({project, repo}) => `${project || ''}_${repo}`,
                ),
                type: NodeType.PIPELINE,
                access: true,
                name: job.job?.pipeline?.name || '',
                state: null,
                jobState: gqlJobStateToNodeState(toGQLJobState(job.state)),
              };

              const outputRepoVertex: Vertex = {
                parents: [
                  `${job.job?.pipeline?.project?.name || ''}_${
                    job.job?.pipeline?.name
                  }`,
                ],
                type: NodeType.OUTPUT_REPO,
                access: true,
                state: null,
                id: `${job.job?.pipeline?.project?.name || ''}_${
                  job.job?.pipeline?.name
                }_repo`,
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
              pachClient.pfs().listRepo({projectIds: [projectId]}),
              pachClient.pps().listPipeline({projectIds: [projectId]}),
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
