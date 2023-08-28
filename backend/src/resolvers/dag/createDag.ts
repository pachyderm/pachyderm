import keyBy from 'lodash/keyBy';
import uniqBy from 'lodash/uniqBy';

import {flattenPipelineInputObj} from '@dash-backend/lib/flattenPipelineInput';
import getJobsFromJobSet from '@dash-backend/lib/getJobsFromJobSet';
import {toGQLJobState} from '@dash-backend/lib/gqlEnumMappers';
import {gqlJobStateToNodeState} from '@dash-backend/lib/nodeStateMappers';
import sortJobInfos from '@dash-backend/lib/sortJobInfos';
import {PachClient} from '@dash-backend/lib/types';
import {NodeType, Vertex} from '@graphqlTypes';

import {convertPachydermTypesToVertex} from './convertPachydermTypesToVertex';
import {postfixNameWithRepo} from './helpers';

// Creates a vertex list of pachyderm repos and pipelines
export const getProjectDAG = async ({
  projectId,
  pachClient,
}: {
  projectId: string;
  pachClient: PachClient;
}) => {
  const [repos, pipelines] = await Promise.all([
    pachClient.pfs.listRepo({projectIds: [projectId]}),
    pachClient.pps.listPipeline({projectIds: [projectId]}),
  ]);

  return convertPachydermTypesToVertex(projectId, repos, pipelines);
};

// Creates a vertex list of pachyderm repos and pipelines composed only of Global ID items
export const getProjectGlobalIdDAG = async ({
  jobSetId,
  projectId,
  pachClient,
}: {
  jobSetId: string;
  projectId: string;
  pachClient: PachClient;
}) => {
  // if given a global ID, we want to construct a DAG using only
  // nodes referred to by jobs sharing that global id
  const jobSet = await pachClient.pps.inspectJobSet({
    id: jobSetId,
    projectId,
    details: false,
  });

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

    const inputRepoVertices: Vertex[] = inputs.map(({project, repo}) => ({
      id: postfixNameWithRepo(`${project || ''}_${repo}`),
      parents: [],
      type: pipelineMap[`${project || ''}_${repo}`]
        ? NodeType.OUTPUT_REPO
        : NodeType.INPUT_REPO,
      state: null,
      jobState: null,
      nodeState: null,
      jobNodeState: null,
      access: true,
      name: postfixNameWithRepo(repo),
    }));

    const pipelineVertex: Vertex = {
      id: `${job.job?.pipeline?.project?.name || ''}_${
        job.job?.pipeline?.name
      }`,
      parents: inputs.map(({project, repo}) => `${project || ''}_${repo}`),
      type: NodeType.PIPELINE,
      access: true,
      name: job.job?.pipeline?.name || '',
      state: null,
      nodeState: null,
      jobState: toGQLJobState(job.state),
      jobNodeState: gqlJobStateToNodeState(toGQLJobState(job.state)),
    };

    const outputRepoVertex: Vertex = {
      parents: [
        `${job.job?.pipeline?.project?.name || ''}_${job.job?.pipeline?.name}`,
      ],
      type: NodeType.OUTPUT_REPO,
      access: true,
      state: null,
      jobState: null,
      nodeState: null,
      jobNodeState: null,
      id: postfixNameWithRepo(
        `${job.job?.pipeline?.project?.name || ''}_${job.job?.pipeline?.name}`,
      ),
      name: postfixNameWithRepo(job.job?.pipeline?.name),
    };

    return [...acc, ...inputRepoVertices, outputRepoVertex, pipelineVertex];
  }, []);

  return uniqBy(vertices, (v) => v.name);
};
