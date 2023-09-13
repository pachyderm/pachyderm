import uniqBy from 'lodash/uniqBy';

import flattenPipelineInput from '@dash-backend/lib/flattenPipelineInput';
import getJobsFromJobSet from '@dash-backend/lib/getJobsFromJobSet';
import {toGQLJobState} from '@dash-backend/lib/gqlEnumMappers';
import {gqlJobStateToNodeState} from '@dash-backend/lib/nodeStateMappers';
import sortJobInfos from '@dash-backend/lib/sortJobInfos';
import {PachClient} from '@dash-backend/lib/types';
import {NodeType, Vertex} from '@graphqlTypes';

import {convertPachydermTypesToVertex} from './convertPachydermTypesToVertex';
import {generateVertexId} from './helpers';

// Creates a vertex list of pachyderm repos and pipelines
export const getProjectVertices = async ({
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
export const getProjectGlobalIdVertices = async ({
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

  const vertices = sortedJobSet.reduce<Vertex[]>((acc, job) => {
    const inputs = job.details?.input
      ? flattenPipelineInput(job.details.input)
      : [];

    const inputRepoVertices: Vertex[] = inputs.map(({project, name}) => ({
      id: generateVertexId(project, name),
      project,
      name,
      parents: [],
      type: projectId !== project ? NodeType.CROSS_PROJECT_REPO : NodeType.REPO,
      state: null,
      jobState: null,
      nodeState: null,
      jobNodeState: null,
      access: true,
    }));

    const pipelineVertex: Vertex = {
      id: generateVertexId(
        job.job?.pipeline?.project?.name || '',
        job.job?.pipeline?.name || '',
      ),
      project: job.job?.pipeline?.project?.name || '',
      name: job.job?.pipeline?.name || '',
      parents: inputs,
      type: NodeType.PIPELINE,
      access: true,
      //TODO: Fix in ticket FRON-1103
      state: null,
      nodeState: null,
      jobState: toGQLJobState(job.state),
      jobNodeState: gqlJobStateToNodeState(toGQLJobState(job.state)),
    };

    return [...acc, ...inputRepoVertices, pipelineVertex];
  }, []);

  return uniqBy(vertices, (v) => v.id);
};
